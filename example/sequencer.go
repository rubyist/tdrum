package main

import (
	"github.com/mkb218/gosndfile/sndfile"
	"github.com/rubyist/drum"
	"log"
	"time"
)

type Sequencer struct {
	patterns    []*drum.Pattern
	instruments map[int32]*Instrument
	pattern     int
	step        int
}

func NewSequencer() *Sequencer {
	return &Sequencer{
		instruments: make(map[int32]*Instrument),
	}
}

func (s *Sequencer) Add(p *drum.Pattern) {
	s.patterns = append(s.patterns, p)
	for _, track := range p.Tracks {
		if _, ok := s.instruments[track.ID]; !ok {
			instrument, err := NewInstrument(track)
			if err != nil {
				log.Fatalf("Error adding track: %s", err)
			}
			s.instruments[track.ID] = instrument
		}
	}
}

func (s *Sequencer) Read(data []int32) {
	sum := int32(0)
	scale := int32(len(s.patterns[s.pattern].Tracks))

	for i := 0; i < len(data); i++ {
		sum = 0
		for _, instrument := range s.instruments {
			sum += instrument.Read() / scale
		}
		data[i] = sum
	}
}

func (s *Sequencer) Start() {
	period := time.Millisecond * time.Duration(((1.0/(s.patterns[s.pattern].Tempo/60.0))/4.0)*1000.0)
	go func() {
		timer := time.NewTicker(period)
		for {
			<-timer.C
			s.Tick()
		}
	}()
}

func (s *Sequencer) Tick() {
	p := s.patterns[s.pattern]
	for i := 0; i < len(p.Tracks); i++ {
		track := p.Tracks[i]
		if track.Steps[s.step] {
			s.instruments[track.ID].Hit()
		}
	}
	s.step++
	if s.step == 16 {
		s.pattern++
	}

	s.step %= 16
	s.pattern %= len(s.patterns)
}

type Instrument struct {
	sample []int32
	cursor int
}

func NewInstrument(t *drum.Track) (*Instrument, error) {
	fileName := t.Name + ".wav"
	var info sndfile.Info
	f, err := sndfile.Open(fileName, sndfile.Read, &info)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buffer := make([]int32, int(info.Frames)*int(info.Channels))
	_, err = f.ReadFrames(buffer)
	if err != nil {
		return nil, err
	}

	return &Instrument{
		sample: buffer,
		cursor: len(buffer),
	}, nil
}

func (i *Instrument) Read() int32 {
	value := int32(0)
	if i.cursor < len(i.sample) {
		value = i.sample[i.cursor]
		i.cursor++
	}
	return value
}

func (i *Instrument) Hit() {
	i.cursor = 0
}
