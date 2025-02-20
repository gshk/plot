// Copyright ©2015 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plotter

import (
	"image/color"

	"github.com/gshk/plot"
	"github.com/gshk/plot/vg"
	"github.com/gshk/plot/vg/draw"
)

// GlyphBoxes implements the Plotter interface, drawing
// all of the glyph boxes of the plot.  This is intended for
// debugging.
type GlyphBoxes struct {
	draw.LineStyle
}

func NewGlyphBoxes() *GlyphBoxes {
	g := new(GlyphBoxes)
	g.Color = color.RGBA{R: 255, A: 255}
	g.Width = vg.Points(0.25)
	return g
}

func (g GlyphBoxes) Plot(c draw.Canvas, plt *plot.Plot) {
	for _, b := range plt.GlyphBoxes(plt) {
		x := c.X(b.X) + b.Rectangle.Min.X
		y := c.Y(b.Y) + b.Rectangle.Min.Y
		c.StrokeLines(g.LineStyle, []vg.Point{
			{x, y},
			{x + b.Rectangle.Size().X, y},
			{x + b.Rectangle.Size().X, y + b.Rectangle.Size().Y},
			{x, y + b.Rectangle.Size().Y},
			{x, y},
		})
	}
}
