package drum

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
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

	var tempo float32
	err = binary.Read(byReader, binary.LittleEndian, &tempo)
	if err != nil {
		return nil, err
	}

	p := &Pattern{
		Version: string(version),
		Tempo:   tempo,
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
	return ""
}

type Track struct {
	Id    int
	Name  string
	Steps [16]bool
}
