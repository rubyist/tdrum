package main

import (
	"code.google.com/p/portaudio-go/portaudio"
	"github.com/mkb218/gosndfile/sndfile"
	"github.com/rubyist/drum"
	"log"
	"sync"
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

	//drum.Encode(p, "bass.splice")

	//	pat, err := drum.DecodeFile("bass.splice")
	//	if err != nil {
	//		log.Fatal(err)
	//	}

	//drum.Play(pat)
	period := time.Millisecond * time.Duration(((1.0/(p.Tempo/60.0))/4.0)*1000.0)

	numTracks := len(p.Tracks)
	buffers := make([][]int32, numTracks)
	cursors := make([]int, numTracks)
	for i := range cursors {
		cursors[i] = -1
	}

	for i, track := range p.Tracks {
		fileName := track.Name + ".wav"
		var info sndfile.Info
		f, err := sndfile.Open(fileName, sndfile.Read, &info)
		if err != nil {
			log.Fatalf("error: %s", err)
		}
		buffer := make([]int32, int(info.Frames)*int(info.Channels))
		_, err = f.ReadItems(buffer)
		if err != nil {
			log.Fatal(err)
		}
		buffers[i] = buffer
		f.Close()
	}

	start := make(chan int)

	go func() {
		<-start
		step := 0
		timer := time.NewTicker(period)
		for {
			<-timer.C
			for i := 0; i < numTracks; i++ {
				if p.Tracks[i].Steps[step] {
					cursors[i] = 0
				}
			}

			step++
			step = step % 16
		}
	}()

	portaudio.Initialize()
	defer portaudio.Terminate()
	//stream, err := portaudio.OpenDefaultStream(0, int(i.Channels), float64(i.Samplerate), 0, kick)
	var v int32
	var starter sync.Once
	stream, err := portaudio.OpenDefaultStream(0, 2, 44100, 0, func(o []int32) {
		starter.Do(func() { close(start) })

		for i := range o {
			v = 0
			for i := 0; i < numTracks; i++ {
				c := cursors[i]
				if c >= 0 && c < len(buffers[i]) {
					v += (buffers[i][c] / int32(numTracks+1))
					cursors[i] = c + 1
				}
			}

			o[i] = v
		}
	})
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	//if err = stream.Start(); err != nil {
	//	log.Fatal(err)
	//}

	stream.Start()
	/*
		if err := stream.Write(); err != nil {
			log.Printf("Got an error: %s", err)
		}
	*/
	defer stream.Stop()
	//if err := stream.Write(); err != nil {
	//	log.Fatalf("Play error: %s", err)
	//}

	//stream.Stop()

	start <- 1
	for {
		time.Sleep(time.Second)
	}
}
