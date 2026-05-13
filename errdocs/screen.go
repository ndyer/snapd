// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2026 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package errdocs

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

// ANSI escape sequences used by the screen renderer.
const (
	// Aubergine: Canonical brand colour. 256-colour 89 is the closest ANSI
	// match for the #772953 hex value.
	ansiBgAubergine = "\x1b[48;5;89m"
	ansiBgWhite     = "\x1b[107m"
	ansiFgWhite     = "\x1b[97m"
	ansiFgBlack     = "\x1b[30m"
	ansiReset       = "\x1b[0m"
	ansiClear       = "\x1b[2J"
	ansiHome        = "\x1b[H"
	ansiCursorHide  = "\x1b[?25l"
	ansiCursorShow  = "\x1b[?25h"
)

// Default terminal dimensions when TIOCGWINSZ fails.
const (
	defaultCols = 80
	defaultRows = 25
)

// IsTerminal reports whether the given file descriptor is a terminal.
func IsTerminal(f *os.File) bool {
	return terminal.IsTerminal(int(f.Fd()))
}

// termSize returns the size of the terminal attached to f, or the
// defaults if the ioctl fails.
func termSize(f *os.File) (cols, rows int) {
	c, r, err := terminal.GetSize(int(f.Fd()))
	if err != nil || c <= 0 || r <= 0 {
		return defaultCols, defaultRows
	}
	return c, r
}

// countdownPos is the cursor position and field width of the
// "Restarting in NN seconds..." line on the painted screen. The
// countdown loop uses it to rewrite the line in place without
// repainting the rest of the screen.
type countdownPos struct {
	row, col, width int
}

// paintScreen renders the full BSOD screen on a TTY-attached writer and
// returns the cursor position of the countdown line. The caller must
// have already emitted ansiClear+ansiHome+ansiCursorHide; paintScreen
// only emits the body.
func paintScreen(w io.Writer, errMsg, url string, qr Grid, cols, rows int) (countdownPos, error) {
	// First wipe the screen to aubergine. We do this by filling every
	// row with spaces using the aubergine background.
	if _, err := fmt.Fprint(w, ansiBgAubergine, ansiFgWhite); err != nil {
		return countdownPos{}, err
	}
	for r := 1; r <= rows; r++ {
		if _, err := fmt.Fprintf(w, "\x1b[%d;1H%s", r, strings.Repeat(" ", cols)); err != nil {
			return countdownPos{}, err
		}
	}

	// Layout: lines of body text, then the QR panel, then the
	// countdown. Top padding adapts to terminal height so things
	// stay roughly centered vertically.
	body := []string{
		"[ Ubuntu Core ]",
		"",
		"Your device has encountered a problem and needs to restart.",
		"",
		"Error: " + errMsg,
		"",
		"For help, scan the QR code below or visit:",
		"  " + url,
	}

	// QR panel size in terminal cells:
	//   width  = qr.Size() + 2 chars of horizontal padding on each side
	//   height = ceil(qr.Size() / 2) + 1 row of vertical padding on each side
	qrSize := qr.Size()
	panelW := qrSize + 4
	panelH := (qrSize+1)/2 + 2

	// Total content height: body + 1 blank + panel + 1 blank + countdown.
	contentH := len(body) + 1 + panelH + 1 + 1
	topPad := (rows - contentH) / 2
	if topPad < 1 {
		topPad = 1
	}

	row := topPad
	for _, line := range body {
		col := (cols - len(line)) / 2
		if col < 1 {
			col = 1
		}
		// "[ Ubuntu Core ]" and the centred body share the centre line.
		// The error/URL lines look better left-aligned with a small
		// indent rather than perfectly centred, but for v1 we keep
		// them centred for simplicity.
		if _, err := fmt.Fprintf(w, "\x1b[%d;%dH%s", row, col, line); err != nil {
			return countdownPos{}, err
		}
		row++
	}
	row++ // blank line

	panelCol := (cols - panelW) / 2
	if panelCol < 1 {
		panelCol = 1
	}
	if err := paintQRPanel(w, qr, row, panelCol); err != nil {
		return countdownPos{}, err
	}
	row += panelH
	row++ // blank line

	// Countdown line. Field is a fixed width so the digits never shift
	// the surrounding text when they go from 2 chars to 1.
	const cdField = "Restarting in NN seconds..."
	cdCol := (cols - len(cdField)) / 2
	if cdCol < 1 {
		cdCol = 1
	}
	pos := countdownPos{row: row, col: cdCol, width: len(cdField)}
	if err := drawCountdown(w, pos, 30); err != nil {
		return countdownPos{}, err
	}

	return pos, nil
}

// paintQRPanel paints the white inset panel containing the QR code at
// the given top-left cell. Inside the panel: white background, black
// foreground for the QR modules drawn with Unicode half-blocks.
func paintQRPanel(w io.Writer, qr Grid, row, col int) error {
	size := qr.Size()
	panelW := size + 4

	if _, err := fmt.Fprint(w, ansiBgWhite, ansiFgBlack); err != nil {
		return err
	}
	// Top padding row.
	if _, err := fmt.Fprintf(w, "\x1b[%d;%dH%s", row, col, strings.Repeat(" ", panelW)); err != nil {
		return err
	}
	// QR rows: 2 modules per row via half-blocks.
	for y := 0; y < size; y += 2 {
		line := make([]byte, 0, size*3+8)
		line = append(line, "  "...) // left padding inside panel
		for x := 0; x < size; x++ {
			top := qr[y][x]
			var bot bool
			if y+1 < size {
				bot = qr[y+1][x]
			}
			switch {
			case top && bot:
				line = append(line, "\xe2\x96\x88"...) // █ full block
			case top && !bot:
				line = append(line, "\xe2\x96\x80"...) // ▀ upper half
			case !top && bot:
				line = append(line, "\xe2\x96\x84"...) // ▄ lower half
			default:
				line = append(line, ' ')
			}
		}
		line = append(line, "  "...) // right padding inside panel
		if _, err := fmt.Fprintf(w, "\x1b[%d;%dH%s", row+1+y/2, col, line); err != nil {
			return err
		}
	}
	// Bottom padding row.
	if _, err := fmt.Fprintf(w, "\x1b[%d;%dH%s", row+(size+1)/2+1, col, strings.Repeat(" ", panelW)); err != nil {
		return err
	}
	// Restore the outer screen colours before returning.
	_, err := fmt.Fprint(w, ansiBgAubergine, ansiFgWhite)
	return err
}

// drawCountdown writes the "Restarting in NN seconds..." line in place
// at the recorded position. Pads to a fixed width so the surrounding
// layout stays stable.
func drawCountdown(w io.Writer, pos countdownPos, seconds int) error {
	text := fmt.Sprintf("Restarting in %2d seconds...", seconds)
	if len(text) < pos.width {
		text += strings.Repeat(" ", pos.width-len(text))
	}
	_, err := fmt.Fprintf(w, "\x1b[%d;%dH%s%s%s", pos.row, pos.col, ansiBgAubergine+ansiFgWhite, text, ansiReset)
	return err
}

// paintPlain renders a non-TTY fallback: just the error and URL, no
// escapes, no countdown sleep.
func paintPlain(w io.Writer, errMsg, url string) error {
	_, err := fmt.Fprintf(w, "error: %s\nFor help: %s\n", errMsg, url)
	return err
}
