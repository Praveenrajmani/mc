/*
 * MinIO Client (C) 2014, 2015 MinIO, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"runtime"
	"strings"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/fatih/color"

	"github.com/minio/mc/pkg/console"
)

// progress extender.
type progressBar struct {
	*pb.ProgressBar
}

// newProgressBar - instantiate a progress bar.
func newProgressBar(total int64) *progressBar {
	// Progress bar speific theme customization.
	console.SetColor("Bar", color.New(color.FgGreen, color.Bold))

	pgbar := progressBar{}

	// get the new original progress bar.
	bar := pb.New64(total)

	// Set new human friendly print units.
	bar.SetUnits(pb.U_BYTES)

	// Refresh rate for progress bar is set to 125 milliseconds.
	bar.SetRefreshRate(time.Millisecond * 125)

	// Do not print a newline by default handled, it is handled manually.
	bar.NotPrint = true

	// Show current speed is true.
	bar.ShowSpeed = true

	// Custom callback with colorized bar.
	bar.Callback = func(s string) {
		console.Print(console.Colorize("Bar", "\r"+s))
	}

	// Use different unicodes for Linux, OS X and Windows.
	switch runtime.GOOS {
	case "linux":
		// Need to add '\x00' as delimiter for unicode characters.
		bar.Format("┃\x00▓\x00█\x00░\x00┃")
	case "darwin":
		// Need to add '\x00' as delimiter for unicode characters.
		bar.Format(" \x00▓\x00 \x00░\x00 ")
	default:
		// Default to non unicode characters.
		bar.Format("[=> ]")
	}

	// Start the progress bar.
	if bar.Total > 0 {
		bar.Start()
	}

	// Copy for future
	pgbar.ProgressBar = bar

	// Return new progress bar here.
	return &pgbar
}

// Set caption.
func (p *progressBar) SetCaption(caption string) *progressBar {
	caption = fixateBarCaption(caption, getFixedWidth(p.ProgressBar.GetWidth(), 18))
	p.ProgressBar.Prefix(caption)
	return p
}

func (p *progressBar) Set64(length int64) *progressBar {
	p.ProgressBar = p.ProgressBar.Set64(length)
	return p
}

func (p *progressBar) Read(buf []byte) (n int, err error) {
	defer func() {
		// After updating the internal progress bar, make sure that its
		// current progress doesn't exceed the specified total progress
		currentProgress := p.ProgressBar.Get()
		if currentProgress > p.ProgressBar.Total {
			p.ProgressBar.Set64(p.ProgressBar.Total)
		}
	}()

	return p.ProgressBar.Read(buf)
}

func (p *progressBar) SetTotal(total int64) *progressBar {
	p.ProgressBar.Total = total
	return p
}

// cursorAnimate - returns a animated rune through read channel for every read.
func cursorAnimate() <-chan string {
	cursorCh := make(chan string)
	var cursors string

	switch runtime.GOOS {
	case "linux":
		// cursors = "➩➪➫➬➭➮➯➱"
		// cursors = "▁▃▄▅▆▇█▇▆▅▄▃"
		cursors = "◐◓◑◒"
		// cursors = "←↖↑↗→↘↓↙"
		// cursors = "◴◷◶◵"
		// cursors = "◰◳◲◱"
		//cursors = "⣾⣽⣻⢿⡿⣟⣯⣷"
	case "darwin":
		cursors = "◐◓◑◒"
	default:
		cursors = "|/-\\"
	}
	go func() {
		for {
			for _, cursor := range cursors {
				cursorCh <- string(cursor)
			}
		}
	}()
	return cursorCh
}

// fixateBarCaption - fancify bar caption based on the terminal width.
func fixateBarCaption(caption string, width int) string {
	switch {
	case len(caption) > width:
		// Trim caption to fit within the screen
		trimSize := len(caption) - width + 3
		if trimSize < len(caption) {
			caption = "..." + caption[trimSize:]
		}
	case len(caption) < width:
		caption += strings.Repeat(" ", width-len(caption))
	}
	return caption
}

// getFixedWidth - get a fixed width based for a given percentage.
func getFixedWidth(width, percent int) int {
	return width * percent / 100
}
