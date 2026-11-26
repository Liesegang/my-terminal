package main

import (
	"bufio"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/creack/pty"
	"io"
	"os"
	"os/exec"
	"time"
)

const MaxBufferSize = 16

func main() {
	a := app.New()
	w := a.NewWindow("germ")

	ui := widget.NewTextGrid()

	os.Setenv("TERM", "myterm")
	c := exec.Command("/bin/bash")
	p, err := pty.Start(c)

	if err != nil {
		fyne.LogError("Failed to open pty", err)
		os.Exit(1)
	}

	defer c.Process.Kill()

	onTypedKey := func(e *fyne.KeyEvent) {
		if e.Name == fyne.KeyEnter || e.Name == fyne.KeyReturn {
			_, _ = p.Write([]byte{'\r'})
		}
	}

	onTypedRune := func(r rune) {
		_, _ = p.WriteString(string(r))
	}

	w.Canvas().SetOnTypedKey(onTypedKey)
	w.Canvas().SetOnTypedRune(onTypedRune)

	buffer := [][]rune{}
	reader := bufio.NewReader(p)

	go func() {
		line := []rune{}
		buffer = append(buffer, line)
		for {
			r, _, err := reader.ReadRune()

			if err != nil {
				if err == io.EOF {
					return
				}
				os.Exit(0)
			}

			line = append(line, r)
			buffer[len(buffer)-1] = line
			if r == '\n' {
				if len(buffer) > MaxBufferSize {
					buffer = buffer[1:]
				}

				line = []rune{}
				buffer = append(buffer, line)
			}
		}
	}()

	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			ui.SetText("")
			var lines string
			for _, line := range buffer {
				lines = lines + string(line)
			}
			ui.SetText(lines)
		}
	}()

	// Create a new container with a wrapped layout
	// set the layout width to 420, height to 200
	w.SetContent(
		container.New(
			layout.NewGridWrapLayout(fyne.NewSize(900, 325)),
			ui,
		),
	)

	w.ShowAndRun()
}
