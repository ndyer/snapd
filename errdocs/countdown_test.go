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
	"bytes"
	"os"
	"strings"
	"syscall"
	"time"

	. "gopkg.in/check.v1"
)

type countdownSuite struct{}

var _ = Suite(&countdownSuite{})

func (s *countdownSuite) TestRunCountdownFiresEachTick(c *C) {
	var buf bytes.Buffer
	pos := countdownPos{row: 20, col: 27, width: len("Restarting in NN seconds...")}
	ticks := make(chan time.Time, 32)
	// Feed all 30 ticks immediately and let the goroutine drain.
	for i := 0; i < 30; i++ {
		ticks <- time.Now()
	}
	stop := make(chan os.Signal)

	c.Assert(runCountdown(&buf, pos, ticks, stop), IsNil)

	out := buf.String()
	// First tick drops counter to 29, last drops to 0; both must
	// appear somewhere in the stream.
	c.Check(strings.Contains(out, "Restarting in 29 seconds..."), Equals, true)
	c.Check(strings.Contains(out, "Restarting in  0 seconds..."), Equals, true)
}

func (s *countdownSuite) TestRunCountdownAbortsOnSignal(c *C) {
	var buf bytes.Buffer
	pos := countdownPos{row: 20, col: 27, width: len("Restarting in NN seconds...")}
	ticks := make(chan time.Time)
	stop := make(chan os.Signal, 1)
	stop <- syscall.SIGINT

	done := make(chan error, 1)
	go func() { done <- runCountdown(&buf, pos, ticks, stop) }()

	select {
	case err := <-done:
		c.Check(err, IsNil)
	case <-time.After(2 * time.Second):
		c.Fatal("runCountdown did not abort on signal")
	}
	// No ticks fired, so no countdown line should have been written.
	c.Check(buf.Len(), Equals, 0)
}
