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
	log.Printf("Period: %v", period)

	mixer := NewBufferedMixer(0)

	for _, track := range p.Tracks {
		fileName := track.Name + ".wav"
		var info sndfile.Info
		f, err := sndfile.Open(fileName, sndfile.Read, &info)
		if err != nil {
			log.Fatalf("error: %s", err)
		}
		log.Printf("Track %s, Frames: %d, Sample Rate: %d, Channels: %d", track.Name, info.Frames, info.Samplerate, info.Channels)
		buffer := make([]int32, int(info.Frames)*int(info.Channels))
		_, err = f.ReadFrames(buffer)
		if err != nil {
			log.Fatal(err)
		}

		mixer.Add(buffer, track.Steps)
		f.Close()
	}

	start := make(chan int)

	go func() {
		<-start
		timer := time.NewTicker(period)
		for {
			<-timer.C
			mixer.Tick()
		}
	}()

	portaudio.Initialize()
	defer portaudio.Terminate()
	var starter sync.Once
	stream, err := portaudio.OpenDefaultStream(0, 2, 44100, 0, func(o []int32) {
		starter.Do(func() { close(start) })
		mixer.Read(o)
	})
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()
	stream.Start()
	defer stream.Stop()

	start <- 1
	for {
		time.Sleep(time.Second)
	}
}
