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

// Package errdocs renders a failure screen with a QR code for boot-blocking
// errors. Errors that should surface a documentation link implement the
// Coded interface; the slug they return is combined with a docs base URL
// supplied by the calling binary to form the URL embedded in the QR.
package errdocs

import "errors"

// Coded is implemented by errors that have a corresponding redirect entry
// in one of the docs repos. The returned slug, combined with the caller's
// docs base, forms the URL embedded in the failure screen's QR code.
//
// Slugs are kebab-case, with no leading or trailing slash. The fallback
// slug used when no Coded error is found in the chain is the empty string,
// which yields the bare "errors/" URL.
type Coded interface {
	DocSlug() string
}

// SlugOf walks the error chain and returns the slug of the first Coded
// error found, or "" if none is present.
func SlugOf(err error) string {
	var c Coded
	if errors.As(err, &c) {
		return c.DocSlug()
	}
	return ""
}

// Sentinel returns an error whose Error() is msg and DocSlug() is slug.
// Use it as a drop-in replacement for errors.New() when an error should
// also surface a QR code on the failure screen.
//
//	var ErrNoSealedKeys = errdocs.Sentinel("no sealed keys", "no-sealed-keys")
func Sentinel(msg, slug string) error {
	return &sentinelError{msg: msg, slug: slug}
}

type sentinelError struct {
	msg  string
	slug string
}

func (e *sentinelError) Error() string   { return e.msg }
func (e *sentinelError) DocSlug() string { return e.slug }
