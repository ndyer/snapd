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
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// countdownDuration is how long the failure screen is held before
// snap-bootstrap returns. systemd's emergency.target reboots ~10s
// after that, so the user-visible total is roughly 40s.
const countdownDuration = 30 * time.Second

// runCountdown decrements a one-second tick from countdownDuration down
// to zero, rewriting the countdown line at pos each tick. It returns
// early if SIGINT is received so a developer at the console can drop
// into emergency shell without waiting.
//
// tick and stopSig are injection points for tests; nil values use a
// real 1s ticker and SIGINT respectively.
func runCountdown(w io.Writer, pos countdownPos, tick <-chan time.Time, stopSig <-chan os.Signal) error {
	if tick == nil {
		t := time.NewTicker(time.Second)
		defer t.Stop()
		tick = t.C
	}
	if stopSig == nil {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT)
		defer signal.Stop(ch)
		stopSig = ch
	}

	remaining := int(countdownDuration / time.Second)
	for remaining > 0 {
		select {
		case <-tick:
			remaining--
			if err := drawCountdown(w, pos, remaining); err != nil {
				return err
			}
		case <-stopSig:
			return nil
		}
	}
	return nil
}
