package main

import (
	"code.google.com/p/portaudio-go/portaudio"
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/rubyist/drum"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	background   = 0x12
	titleBG      = 0x22
	textBG       = 0xa3
	hitFG        = 0xc5
	tracksBG     = 0x3a
	curStepBG    = 0xf6
	cornerTL     = '\u256d'
	cornerTR     = '\u256e'
	cornerBL     = '\u2570'
	cornerBR     = '\u256f'
	titleLeaderL = '\u257c'
	titleLeaderR = '\u257e'
	hLine        = '\u2500'
	vLine        = '\u2502'
	hit          = '\u2055'
	noHit        = '-'
)

var (
	nameT  = []rune{'n', 'a', 'm', 'e'}
	tempoT = []rune{'t', 'e', 'm', 'p', 'o'}
	timeT  = []rune{'t', 'i', 'm', 'e'}
)

var sequencer *Sequencer

func box(column, row, width, height int, fill termbox.Attribute) {
	// Top left
	termbox.SetCell(column, row, cornerTL, termbox.ColorDefault, background)

	// hline
	for c := column + 1; c < width; c++ {
		termbox.SetCell(c, row, hLine, termbox.ColorDefault, background)
	}

	// Top right
	termbox.SetCell(column+width-1, row, cornerTR, termbox.ColorDefault, background)

	// vlines
	for r := row + 1; r < row+height; r++ {
		termbox.SetCell(column, r, vLine, termbox.ColorDefault, background)         // left side
		termbox.SetCell(column+width-1, r, vLine, termbox.ColorDefault, background) // right side
	}

	// Bottom left
	termbox.SetCell(column, row+height, cornerBL, termbox.ColorDefault, background)

	// hline
	for c := 1; c < width; c++ {
		termbox.SetCell(c, row+height, hLine, termbox.ColorDefault, background)
	}

	// Bottom right
	termbox.SetCell(column+width-1, row+height, cornerBR, termbox.ColorDefault, background)

	// Fill color
	for c := column + 1; c < width-1; c++ {
		for r := row + 1; r < row+height; r++ {
			termbox.SetCell(c, r, ' ', termbox.ColorDefault, fill)
		}
	}
}

func textBox(row, column, width int, title, value string) {
	col := column
	stop := column + width - 1

	// title line
	// corner
	termbox.SetCell(col, row, cornerTL, termbox.ColorDefault, background)
	col++

	if title != "" {
		// leader
		termbox.SetCell(col, row, titleLeaderL, termbox.ColorDefault, background)
		col++

		// title
		for _, c := range title {
			termbox.SetCell(col, row, c, termbox.ColorDefault, titleBG)
			col++
		}

		// leader
		termbox.SetCell(col, row, titleLeaderR, termbox.ColorDefault, background)
		col++
	}

	// line padding
	for i := col; i < stop; i++ {
		termbox.SetCell(col, row, hLine, termbox.ColorDefault, background)
		col++
	}

	// corner
	termbox.SetCell(col, row, cornerTR, termbox.ColorDefault, background)
	col++

	// text line
	col = column
	row++

	// bar
	termbox.SetCell(col, row, '\u2502', termbox.ColorDefault, background)
	col++

	// one space padding
	termbox.SetCell(col, row, ' ', termbox.ColorDefault, textBG)
	col++

	// value
	for _, c := range value {
		termbox.SetCell(col, row, c, termbox.ColorDefault, textBG)
		col++
	}

	// end padding
	for i := col; i < stop; i++ {
		termbox.SetCell(col, row, ' ', termbox.ColorDefault, textBG)
		col++
	}

	// bar
	termbox.SetCell(col, row, '\u2502', termbox.ColorDefault, background)
	col++

	// bottom line
	col = column
	row++

	// corner
	termbox.SetCell(col, row, cornerBL, termbox.ColorDefault, background)
	col++

	// line padding
	for i := col; i < stop; i++ {
		termbox.SetCell(col, row, hLine, termbox.ColorDefault, background)
		col++
	}

	// corner
	termbox.SetCell(col, row, cornerBR, termbox.ColorDefault, background)
}

func drawSteps(row int, steps []bool) {
	if len(steps) != 16 {
		panic("invalid set of steps")
	}

	curStep := sequencer.Step

	w, _ := termbox.Size()
	start := w - 41
	col := start

	for i := 0; i < 16; i++ {
		if i%4 == 0 {
			termbox.SetCell(col, row, vLine, termbox.ColorDefault, tracksBG)
			col++

			termbox.SetCell(col, row, ' ', termbox.ColorDefault, tracksBG)
			col++
		}

		bg := termbox.Attribute(tracksBG)
		if i == curStep {
			bg = curStepBG
		}

		if steps[i] {
			termbox.SetCell(col, row, hit, hitFG, bg)
			col++
		} else {
			termbox.SetCell(col, row, noHit, termbox.ColorDefault, bg)
			col++
		}

		termbox.SetCell(col, row, ' ', termbox.ColorDefault, tracksBG)
		col++
	}
}

func drawTrack(row int, track *drum.Track) {
	col := 1

	termbox.SetCell(col, row, ' ', termbox.ColorDefault, tracksBG)
	col++

	id := fmt.Sprintf("%03d", track.ID)
	for _, c := range id {
		termbox.SetCell(col, row, c, termbox.ColorDefault, tracksBG)
		col++
	}

	termbox.SetCell(col, row, ' ', termbox.ColorDefault, tracksBG)
	col++

	termbox.SetCell(col, row, vLine, termbox.ColorDefault, tracksBG)
	col++

	termbox.SetCell(col, row, ' ', termbox.ColorDefault, tracksBG)
	col++

	for _, c := range track.Name {
		termbox.SetCell(col, row, c, termbox.ColorDefault, tracksBG)
		col++
	}

	drawSteps(row, track.Steps)
}

func draw(pattern *drum.Pattern) {
	w, h := termbox.Size()
	termbox.Clear(termbox.ColorDefault, background)

	// Name box
	name := strings.TrimSuffix(filepath.Base(os.Args[1]), ".splice")
	textBox(0, 0, w-12-10, "name", name)

	// Tempo box
	textBox(0, w-12-10, 10, "tempo", fmt.Sprintf("%v", pattern.Tempo))

	// Time box
	textBox(0, w-12, 12, "time", time.Now().Format("15:04:05"))

	// Version box?
	textBox(h-3, 0, w, "", fmt.Sprintf("tDrum v0.0.0 (HW Version %s)", pattern.Version))

	// Steps outline
	box(0, 3, w, h-7, tracksBG)

	trackRow := 4

	for _, track := range pattern.Tracks {
		drawTrack(trackRow, track)
		trackRow++
	}

	// Remaining bar lines
	termbox.SetCell(6, 3, '\u252c', termbox.ColorDefault, background)
	termbox.SetCell(w-11, 3, '\u252c', termbox.ColorDefault, background)
	termbox.SetCell(w-21, 3, '\u252c', termbox.ColorDefault, background)
	termbox.SetCell(w-31, 3, '\u252c', termbox.ColorDefault, background)
	termbox.SetCell(w-41, 3, '\u252c', termbox.ColorDefault, background)
	termbox.SetCell(6, h-4, '\u2534', termbox.ColorDefault, background)
	termbox.SetCell(w-11, h-4, '\u2534', termbox.ColorDefault, background)
	termbox.SetCell(w-21, h-4, '\u2534', termbox.ColorDefault, background)
	termbox.SetCell(w-31, h-4, '\u2534', termbox.ColorDefault, background)
	termbox.SetCell(w-41, h-4, '\u2534', termbox.ColorDefault, background)

	for r := trackRow; r < h-4; r++ {
		termbox.SetCell(6, r, vLine, termbox.ColorDefault, tracksBG)
		termbox.SetCell(w-11, r, vLine, termbox.ColorDefault, tracksBG)
		termbox.SetCell(w-21, r, vLine, termbox.ColorDefault, tracksBG)
		termbox.SetCell(w-31, r, vLine, termbox.ColorDefault, tracksBG)
		termbox.SetCell(w-41, r, vLine, termbox.ColorDefault, tracksBG)
	}

	// step columns
	stepCols := []int{
		w - 39, w - 37, w - 35, w - 33,
		w - 29, w - 27, w - 25, w - 23,
		w - 19, w - 17, w - 15, w - 13,
		w - 9, w - 7, w - 5, w - 3,
	}
	step := sequencer.Step
	for r := trackRow; r < h-4; r++ {
		for i, c := range stepCols {
			if i == step {
				termbox.SetCell(c, r, ' ', termbox.ColorDefault, curStepBG)
			}
		}
	}

	termbox.Flush()
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage: tdrum file.splice")
		os.Exit(1)
	}

	pattern, err := drum.DecodeFile(os.Args[1])
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}

	sequencer = NewSequencer()
	if err := sequencer.Add(pattern); err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}

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

	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.SetOutputMode(termbox.Output256)

	eq := make(chan termbox.Event)
	go func() {
		for {
			eq <- termbox.PollEvent()
		}
	}()

	draw(pattern)
loop:
	for {
		select {
		case ev := <-eq:
			if ev.Type == termbox.EventKey && ev.Key == termbox.KeyEsc {
				break loop
			}
			if ev.Type == termbox.EventKey && ev.Key == termbox.KeySpace {
				if sequencer.Running {
					sequencer.Stop()
				} else {
					sequencer.Start()
				}
			}
		default:
			draw(pattern)
			time.Sleep(time.Millisecond * 2)
		}
	}
}
