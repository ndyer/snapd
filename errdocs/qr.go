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

import "rsc.io/qr"

// Grid is a square 2D matrix of QR modules. grid[y][x] is true when the
// module is "dark" (the part a scanner reads as black).
type Grid [][]bool

// Size returns the grid's side length in modules.
func (g Grid) Size() int { return len(g) }

// EncodeQR returns the QR module grid for the given URL at the lowest
// error-correction level that fits. Lower levels yield smaller grids,
// which is what we want on a constrained terminal.
func EncodeQR(url string) (Grid, error) {
	code, err := qr.Encode(url, qr.L)
	if err != nil {
		return nil, err
	}
	size := code.Size
	g := make(Grid, size)
	for y := 0; y < size; y++ {
		row := make([]bool, size)
		for x := 0; x < size; x++ {
			row[x] = code.Black(x, y)
		}
		g[y] = row
	}
	return g, nil
}
