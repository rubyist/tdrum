package main

import (
	"code.google.com/p/portaudio-go/portaudio"
	"github.com/rubyist/drum"
	"log"
	"time"
)

func main() {
	kick := &drum.Track{
		ID:   1,
		Name: "kick",
		Steps: []bool{
			true, false, false, true,
			true, false, false, false,
			true, false, true, false,
			true, false, false, true,
		},
	}

	snare := &drum.Track{
		ID:   2,
		Name: "snare",
		Steps: []bool{
			false, false, false, false,
			true, false, false, false,
			false, false, false, false,
			false, true, false, false,
		},
	}

	hat := &drum.Track{
		ID:   3,
		Name: "clhat",
		Steps: []bool{
			true, false, true, false,
			true, false, true, false,
			true, false, true, false,
			true, false, false, false,
		},
	}

	ohat := &drum.Track{
		ID:   5,
		Name: "ohat",
		Steps: []bool{
			false, false, false, false,
			false, false, false, false,
			false, false, false, false,
			false, false, true, false,
		},
	}

	crash := &drum.Track{
		ID:   4,
		Name: "crash",
		Steps: []bool{
			true, false, false, false,
			false, false, false, false,
			false, false, false, false,
			false, false, false, false,
		},
	}

	p := &drum.Pattern{
		Version: "0.909",
		Tempo:   60.0,
		Tracks:  []*drum.Track{kick, snare, hat, ohat, crash},
	}

	log.Print(p.String())

	drum.Encode(p, "test.splice")

	//	pat, err := drum.DecodeFile("bass.splice")
	//	if err != nil {
	//		log.Fatal(err)
	//	}

	sequencer := NewSequencer()
	sequencer.Add(p)

	portaudio.Initialize()
	defer portaudio.Terminate()
	stream, err := portaudio.OpenDefaultStream(0, 2, 44100, 0, func(o []int32) {
		sequencer.Read(o)
	})
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()
	stream.Start()
	defer stream.Stop()

	sequencer.Start()

	for {
		time.Sleep(time.Second)
	}
}
