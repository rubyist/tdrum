package drum

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"os"
	"unsafe"
)

type spliceWriter struct {
	w     io.Writer
	err   error
	count uint64
	b     bytes.Buffer
}

func (p *spliceWriter) write(buf []byte) {
	if p.err != nil {
		return
	}

	n, err := p.b.Write(buf)
	if err != nil {
		p.err = err
		return
	}

	log.Printf("Increasing count by %d", n)
	p.count += uint64(n)
}

func (p *spliceWriter) bwrite(order binary.ByteOrder, data interface{}) {
	if p.err != nil {
		return
	}

	// Should really use reflect and get the size?
	err := binary.Write(&p.b, order, data)
	if err != nil {
		p.err = err
		return
	}

	c := unsafe.Sizeof(data)

	log.Printf("(b) Increasing count by %d", c)
	p.count += uint64(c)
}

func (p *spliceWriter) flush() error {
	if p.err != nil {
		return p.err
	}

	// should handle short writes
	if _, err := p.w.Write([]byte(spliceHeader)); err != nil {
		return err
	}

	log.Printf("Writing a count of %d\n", p.count)
	if err := binary.Write(p.w, binary.BigEndian, p.count); err != nil {
		return err
	}

	if _, err := io.Copy(p.w, &p.b); err != nil {
		return err
	}

	return p.err
}

// Encode encodes the pattern to the file found at the provided path.
func Encode(pat *Pattern, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := &spliceWriter{w: file}

	// Write version string with null padding
	buf.write([]byte(version))
	buf.write(bytes.Repeat([]byte{0}, 32-len(version)))

	// Write tempo
	buf.bwrite(binary.LittleEndian, pat.Tempo)

	// Write tracks
	for _, track := range pat.Tracks {
		buf.bwrite(binary.LittleEndian, track.ID)
		buf.bwrite(binary.BigEndian, uint8(len(track.Name)))
		buf.write([]byte(track.Name))
		for _, step := range track.Steps {
			if step {
				buf.bwrite(binary.BigEndian, []byte{1})
			} else {
				buf.bwrite(binary.BigEndian, []byte{0})
			}
		}
	}

	return buf.flush()
}
