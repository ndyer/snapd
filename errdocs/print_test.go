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
	"bytes"
	"errors"
	"strings"

	. "gopkg.in/check.v1"

	"github.com/snapcore/snapd/errdocs"
)

type printSuite struct{}

var _ = Suite(&printSuite{})

func (s *printSuite) TestPrintBootErrorPlainFallback(c *C) {
	// bytes.Buffer is not a *os.File, so PrintBootError takes the
	// plain-text path.
	var buf bytes.Buffer
	err := errdocs.Sentinel("no sealed keys", "no-sealed-keys")
	errdocs.PrintBootError(&buf, "https://documentation.ubuntu.com/core/", err)

	out := buf.String()
	c.Check(out, Equals, "error: no sealed keys\nFor help: https://documentation.ubuntu.com/core/errors/no-sealed-keys/\n")
	c.Check(strings.Contains(out, "\x1b["), Equals, false)
}

func (s *printSuite) TestPrintBootErrorFallbackURLForUnknown(c *C) {
	var buf bytes.Buffer
	errdocs.PrintBootError(&buf, "https://documentation.ubuntu.com/core/", errors.New("mystery"))
	out := buf.String()
	c.Check(strings.Contains(out, "/errors/\n"), Equals, true, Commentf("got %q", out))
	c.Check(strings.Contains(out, "/errors/mystery/"), Equals, false)
}

func (s *printSuite) TestPrintBootErrorNilErrorIsNoOp(c *C) {
	var buf bytes.Buffer
	errdocs.PrintBootError(&buf, "https://documentation.ubuntu.com/core/", nil)
	c.Check(buf.Len(), Equals, 0)
}

func (s *printSuite) TestPrintBootErrorBaseWithoutTrailingSlash(c *C) {
	var buf bytes.Buffer
	err := errdocs.Sentinel("x", "x")
	errdocs.PrintBootError(&buf, "https://example.com", err)
	c.Check(strings.Contains(buf.String(), "https://example.com/errors/x/"), Equals, true)
}
