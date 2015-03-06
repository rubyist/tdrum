package drum

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	spliceHeader = "SPLICE"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	buf := bufio.NewReader(file)

	// Read the SPLICE header
	header := make([]byte, 6)
	buf.Read(header)
	if string(header) != spliceHeader {
		return nil, errors.New("Invalid splice file")
	}

	// Next up is an integer containing the number of bytes left to read
	var remaining int64
	err = binary.Read(buf, binary.BigEndian, &remaining)
	if err != nil {
		return nil, err
	}

	// The 32 byte long version string
	version := make([]byte, 32)
	_, err = io.ReadFull(buf, version)
	if err != nil {
		return nil, err
	}
	version = bytes.Trim(version, "\x00")
	remaining -= 32

	// Tempo is a little endian 32 bit floating point
	var tempo float32
	err = binary.Read(buf, binary.LittleEndian, &tempo)
	if err != nil {
		return nil, err
	}
	remaining -= 4

	var tracks []*Track
	for remaining > 0 {
		// Track id is a little endian 32 bit int
		var id int32
		err = binary.Read(buf, binary.LittleEndian, &id)
		if err != nil {
			return nil, err
		}
		remaining -= 4

		// A byte indicating the length of the instrument name
		l, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}
		remaining--

		// The instrument name
		name := make([]byte, l)
		_, err = io.ReadFull(buf, name)
		if err != nil {
			return nil, err
		}
		remaining -= int64(l)

		// 16 steps, 1 byte each
		steps := make([]bool, 16)
		for i := 0; i < 16; i++ {
			s, err := buf.ReadByte()
			if err != nil {
				return nil, err
			}
			steps[i] = s == 1
			remaining--
		}

		track := &Track{
			ID:    id,
			Name:  string(name),
			Steps: steps,
		}
		tracks = append(tracks, track)
	}

	p := &Pattern{
		Version: string(version),
		Tempo:   tempo,
		Tracks:  tracks,
	}
	return p, nil
}

func Encode(pat *Pattern, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := bufio.NewWriter(file)
	// Write header
	_, err = buf.WriteString(spliceHeader)
	if err != nil {
		return err
	}

	// Write bytes left
	var total uint64
	total = 32                           // version string
	total += 4                           // tempo
	total += uint64(len(pat.Tracks) * 4) // track ids
	total += uint64(len(pat.Tracks) * 1) // inst lengths
	for _, track := range pat.Tracks {
		total += uint64(len(track.Name)) // inst names
	}
	total += uint64(len(pat.Tracks)) * 16 // steps
	err = binary.Write(buf, binary.BigEndian, total)
	if err != nil {
		return err
	}

	// Write version string
	v := "0.808-alpha"
	buf.WriteString(v)
	for i := 0; i < 32-len(v); i++ { // sucks
		err := buf.WriteByte(0)
		if err != nil {
			return err
		}
	}

	// Write tempo
	err = binary.Write(buf, binary.LittleEndian, pat.Tempo)
	if err != nil {
		return err
	}

	// Write tracks
	for _, track := range pat.Tracks {
		binary.Write(buf, binary.LittleEndian, track.ID)
		binary.Write(buf, binary.BigEndian, uint8(len(track.Name)))
		buf.WriteString(track.Name)
		for _, step := range track.Steps {
			if step {
				buf.WriteByte(1)
			} else {
				buf.WriteByte(0)
			}
		}
	}

	buf.Flush()
	return nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []*Track
}

func (p *Pattern) String() string {
	s := "Saved with HW Version: " + p.Version + "\n"
	s += fmt.Sprintf("Tempo: %v\n", p.Tempo)
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
