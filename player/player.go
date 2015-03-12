package main

import (
	"code.google.com/p/portaudio-go/portaudio"
	"github.com/rubyist/drum"
	"log"
	"time"
)

func main() {
	pat1, err := drum.DecodeFile("test.splice")
	if err != nil {
		log.Fatal(err)
	}
	log.Print(pat1.String())

	pat2, err := drum.DecodeFile("test2.splice")
	if err != nil {
		log.Fatal(err)
	}
	log.Print(pat2.String())

	sequencer := NewSequencer()
	sequencer.Add(pat1)
	sequencer.Add(pat2)

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
