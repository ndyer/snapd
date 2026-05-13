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
)

// PrintBootError renders the Ubuntu Core "screen of death" for a
// boot-blocking error and holds it for 30 seconds before returning.
//
// base is the docs site root (with trailing slash) for the calling
// binary: "https://documentation.ubuntu.com/core/" for snap-bootstrap,
// "https://snapcraft.io/docs/" for the snapd daemon when wired in.
//
// If w is not a TTY, falls back to a plain-text error+URL line and
// returns immediately with no countdown.
func PrintBootError(w io.Writer, base string, err error) {
	if err == nil {
		return
	}
	url := urlFor(base, err)
	msg := err.Error()

	f, isFile := w.(*os.File)
	if !isFile || !IsTerminal(f) {
		_ = paintPlain(w, msg, url)
		return
	}

	cols, rows := termSize(f)
	qr, qrErr := EncodeQR(url)
	if qrErr != nil {
		_ = paintPlain(w, msg, url)
		return
	}

	// Enter screen mode.
	if _, e := fmt.Fprint(w, ansiClear, ansiHome, ansiCursorHide); e != nil {
		return
	}
	pos, paintErr := paintScreen(w, msg, url, qr, cols, rows)
	if paintErr != nil {
		_, _ = fmt.Fprint(w, ansiReset, ansiCursorShow)
		return
	}

	_ = runCountdown(w, pos, nil, nil)

	// Restore terminal state. The caller will exit shortly after,
	// but be polite in case anything is still using the TTY.
	_, _ = fmt.Fprint(w, ansiReset, ansiCursorShow, "\n")
}

// urlFor builds the docs URL for an error: <base>errors/<slug>/, with
// the slug coming from a Coded error in err's chain. Unknown errors
// fall through to the bare errors/ path which the RTD wildcard then
// redirects to the bug-report page.
func urlFor(base string, err error) string {
	b := strings.TrimRight(base, "/")
	slug := SlugOf(err)
	if slug == "" {
		return b + "/errors/"
	}
	return b + "/errors/" + slug + "/"
}
