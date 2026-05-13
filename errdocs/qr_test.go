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

package errdocs_test

import (
	. "gopkg.in/check.v1"

	"github.com/snapcore/snapd/errdocs"
)

type qrSuite struct{}

var _ = Suite(&qrSuite{})

func (s *qrSuite) TestEncodeQRBasic(c *C) {
	g, err := errdocs.EncodeQR("https://documentation.ubuntu.com/core/errors/no-sealed-keys/")
	c.Assert(err, IsNil)
	c.Check(g.Size() > 0, Equals, true)
	c.Check(g.Size(), Equals, len(g[0]))

	// Finder patterns: the three 7x7 squares in the corners are
	// always present in a QR code. Quick sanity check on the
	// top-left corner.
	c.Check(g[0][0], Equals, true)
	c.Check(g[6][6], Equals, true)
	c.Check(g[0][6], Equals, true)
	c.Check(g[6][0], Equals, true)
	c.Check(g[1][1], Equals, false)
}

func (s *qrSuite) TestEncodeQRFitsTypicalTerminal(c *C) {
	// QR for a typical errdocs URL should fit comfortably in the
	// inset panel of an 80-column terminal (allowing for borders +
	// padding, we want < 60 modules per side).
	g, err := errdocs.EncodeQR("https://documentation.ubuntu.com/core/errors/recovery-mode-unsupported/")
	c.Assert(err, IsNil)
	c.Check(g.Size() < 60, Equals, true, Commentf("QR is %dx%d", g.Size(), g.Size()))
}
