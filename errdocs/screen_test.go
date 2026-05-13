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
	"strings"

	. "gopkg.in/check.v1"
)

type screenSuite struct{}

var _ = Suite(&screenSuite{})

func (s *screenSuite) TestPaintPlain(c *C) {
	var buf bytes.Buffer
	c.Assert(paintPlain(&buf, "boom", "https://docs.example/errors/boom/"), IsNil)
	out := buf.String()
	c.Check(out, Equals, "error: boom\nFor help: https://docs.example/errors/boom/\n")
	c.Check(strings.Contains(out, "\x1b["), Equals, false, Commentf("plain output must not contain ANSI escapes"))
}

func (s *screenSuite) TestPaintScreenContainsTheme(c *C) {
	var buf bytes.Buffer
	qr, err := EncodeQR("https://documentation.ubuntu.com/core/errors/no-sealed-keys/")
	c.Assert(err, IsNil)

	pos, err := paintScreen(&buf, "no sealed keys", "https://documentation.ubuntu.com/core/errors/no-sealed-keys/", qr, 80, 25)
	c.Assert(err, IsNil)
	out := buf.String()

	c.Check(strings.Contains(out, ansiBgAubergine), Equals, true)
	c.Check(strings.Contains(out, ansiBgWhite), Equals, true)
	c.Check(strings.Contains(out, ansiFgWhite), Equals, true)
	c.Check(strings.Contains(out, ansiFgBlack), Equals, true)

	c.Check(strings.Contains(out, "[ Ubuntu Core ]"), Equals, true)
	c.Check(strings.Contains(out, "Error: no sealed keys"), Equals, true)
	c.Check(strings.Contains(out, "https://documentation.ubuntu.com/core/errors/no-sealed-keys/"), Equals, true)
	c.Check(strings.Contains(out, "Restarting in 30 seconds..."), Equals, true)

	c.Check(pos.row > 0, Equals, true)
	c.Check(pos.col > 0, Equals, true)
	c.Check(pos.width >= len("Restarting in 30 seconds..."), Equals, true)
}

func (s *screenSuite) TestDrawCountdownUpdatesDigit(c *C) {
	var buf bytes.Buffer
	pos := countdownPos{row: 20, col: 27, width: len("Restarting in NN seconds...")}
	c.Assert(drawCountdown(&buf, pos, 5), IsNil)
	out := buf.String()
	c.Check(strings.Contains(out, "Restarting in  5 seconds..."), Equals, true)
	c.Check(strings.Contains(out, "\x1b[20;27H"), Equals, true)
}
