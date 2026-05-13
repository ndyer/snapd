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
	"errors"
	"fmt"
	"testing"

	. "gopkg.in/check.v1"

	"github.com/snapcore/snapd/errdocs"
)

func Test(t *testing.T) { TestingT(t) }

type codedSuite struct{}

var _ = Suite(&codedSuite{})

func (s *codedSuite) TestSentinelMessageAndSlug(c *C) {
	err := errdocs.Sentinel("no sealed keys", "no-sealed-keys")
	c.Check(err.Error(), Equals, "no sealed keys")
	c.Check(errdocs.SlugOf(err), Equals, "no-sealed-keys")
}

func (s *codedSuite) TestSentinelImplementsCoded(c *C) {
	err := errdocs.Sentinel("msg", "slug")
	var c1 errdocs.Coded
	c.Assert(errors.As(err, &c1), Equals, true)
	c.Check(c1.DocSlug(), Equals, "slug")
}

func (s *codedSuite) TestSlugOfWrapped(c *C) {
	inner := errdocs.Sentinel("inner", "inner-slug")
	wrapped := fmt.Errorf("wrap: %w", inner)
	c.Check(errdocs.SlugOf(wrapped), Equals, "inner-slug")
}

func (s *codedSuite) TestSlugOfDoubleWrapped(c *C) {
	inner := errdocs.Sentinel("inner", "inner-slug")
	wrapped := fmt.Errorf("once: %w", fmt.Errorf("twice: %w", inner))
	c.Check(errdocs.SlugOf(wrapped), Equals, "inner-slug")
}

func (s *codedSuite) TestSlugOfNonCoded(c *C) {
	c.Check(errdocs.SlugOf(errors.New("plain")), Equals, "")
}

func (s *codedSuite) TestSlugOfNil(c *C) {
	c.Check(errdocs.SlugOf(nil), Equals, "")
}

func (s *codedSuite) TestSentinelIdentityPreserved(c *C) {
	// Replacing errors.New with errdocs.Sentinel must preserve the
	// errors.Is identity semantics relied on by existing callers:
	// the same package-level var compares equal to itself.
	pkgErr := errdocs.Sentinel("boom", "boom")
	wrapped := fmt.Errorf("ctx: %w", pkgErr)
	c.Check(errors.Is(wrapped, pkgErr), Equals, true)

	// Two distinct Sentinel() calls with identical args remain
	// distinct (matches errors.New behaviour).
	a := errdocs.Sentinel("x", "x")
	b := errdocs.Sentinel("x", "x")
	c.Check(errors.Is(a, b), Equals, false)
}

// A typed error that implements Coded directly, as is the pattern for
// struct errors like AuthQualityError.
type typedErr struct{ inner error }

func (e *typedErr) Error() string   { return e.inner.Error() }
func (e *typedErr) DocSlug() string { return "typed-slug" }

func (s *codedSuite) TestTypedCodedError(c *C) {
	err := &typedErr{inner: errors.New("typed-msg")}
	c.Check(errdocs.SlugOf(err), Equals, "typed-slug")
}

func (s *codedSuite) TestTypedCodedErrorWrapped(c *C) {
	err := &typedErr{inner: errors.New("typed-msg")}
	wrapped := fmt.Errorf("outer: %w", err)
	c.Check(errdocs.SlugOf(wrapped), Equals, "typed-slug")
}
