package drum

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
// TODO: implement
func DecodeFile(path string) (*Pattern, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	header := make([]byte, 6)
	file.Read(header)
	if string(header) != "SPLICE" {
		return nil, errors.New("Invalid splice file")
	}

	byReader := bufio.NewReader(file)
	var byLeft int64
	err = binary.Read(byReader, binary.BigEndian, &byLeft)
	if err != nil {
		return nil, err
	}

	version := make([]byte, 32)
	_, err = byReader.Read(version)
	if err != nil {
		return nil, err
	}
	byLeft -= 32

	var tempo float32
	err = binary.Read(byReader, binary.LittleEndian, &tempo)
	if err != nil {
		return nil, err
	}
	byLeft -= 4

	var tracks []*Track
	for byLeft > 0 {
		var id int32
		err = binary.Read(byReader, binary.LittleEndian, &id)
		if err != nil {
			return nil, err
		}
		byLeft -= 4

		l, err := byReader.ReadByte()
		if err != nil {
			return nil, err
		}
		byLeft--

		name := make([]byte, l)
		_, err = byReader.Read(name)
		if err != nil {
			return nil, err
		}
		byLeft -= int64(l)

		steps := make([]bool, 16)
		for i := 0; i < 16; i++ {
			s, err := byReader.ReadByte()
			if err != nil {
				return nil, err
			}
			steps[i] = s == 1
			byLeft--
		}

		track := &Track{
			Id:    id,
			Name:  string(name),
			Steps: steps,
		}
		tracks = append(tracks, track)
	}

	p := &Pattern{
		Version: string(bytes.Trim(version, "\x00")),
		Tempo:   tempo,
		Tracks:  tracks,
	}
	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
// TODO: implement
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []*Track
}

func (p *Pattern) String() string {
	s := "Saved with HW Version: " + p.Version + "\n"
	s += fmt.Sprintf("Tempo: %s\n", strings.TrimSuffix(fmt.Sprintf("%.1f", p.Tempo), ".0"))
	for _, track := range p.Tracks {
		s += track.String()
	}

	return s
}

// Track is the instrument track within the Pattern.
type Track struct {
	ID    int32
	Name  string
	Steps []bool
}

func (t *Track) String() string {
	s := fmt.Sprintf("(%d) %s\t", t.ID, t.Name)
	for i, step := range t.Steps {
		if i%4 == 0 {
			s += "|"
		}
		if step {
			s += "x"
		} else {
			s += "-"
		}
	}
	s += "|\n"
	return s
}
