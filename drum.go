// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"fmt"
	"time"
)

func (p *Pattern) PlayStep(step int) {
	for _, track := range p.Tracks {
		if track.Steps[step] {
			fmt.Printf("(%d) %s\n", step, track.Name)
		}
	}
}

func Play(p *Pattern) {
	period := time.Millisecond * time.Duration(((1.0/(p.Tempo/60.0))/4.0)*1000.0)
	fmt.Printf("tempo: %v period: %v\n", p.Tempo, period)
	for i := 0; ; i++ {
		p.PlayStep(i % 16)
		time.Sleep(period)
	}
}

type Sequence struct {
	Patterns []*Pattern
}

func NewSequence() *Sequence {
	return &Sequence{}
}

func (s *Sequence) Add(p *Pattern) {
	s.Patterns = append(s.Patterns, p)
}

func (s *Sequence) Play() {
}
