// Copyright ©2015 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plotter_test

import (
	"log"
	"testing"

	"github.com/gshk/plot"
	"github.com/gshk/plot/cmpimg"
	"github.com/gshk/plot/plotter"
	"github.com/gshk/plot/vg"
)

// Draw the plot logo.
func Example() {
	p, err := plot.New()
	if err != nil {
		log.Panic(err)
	}

	plotter.DefaultLineStyle.Width = vg.Points(1)
	plotter.DefaultGlyphStyle.Radius = vg.Points(3)

	p.Y.Tick.Marker = plot.ConstantTicks([]plot.Tick{
		{0, "0"}, {0.25, ""}, {0.5, "0.5"}, {0.75, ""}, {1, "1"},
	})
	p.X.Tick.Marker = plot.ConstantTicks([]plot.Tick{
		{0, "0"}, {0.25, ""}, {0.5, "0.5"}, {0.75, ""}, {1, "1"},
	})

	pts := plotter.XYs{{0, 0}, {0, 1}, {0.5, 1}, {0.5, 0.6}, {0, 0.6}}
	line, err := plotter.NewLine(pts)
	if err != nil {
		log.Panic(err)
	}
	scatter, err := plotter.NewScatter(pts)
	if err != nil {
		log.Panic(err)
	}
	p.Add(line, scatter)

	pts = plotter.XYs{{1, 0}, {0.75, 0}, {0.75, 0.75}}
	line, err = plotter.NewLine(pts)
	if err != nil {
		log.Panic(err)
	}
	scatter, err = plotter.NewScatter(pts)
	if err != nil {
		log.Panic(err)
	}
	p.Add(line, scatter)

	pts = plotter.XYs{{0.5, 0.5}, {1, 0.5}}
	line, err = plotter.NewLine(pts)
	if err != nil {
		log.Panic(err)
	}
	scatter, err = plotter.NewScatter(pts)
	if err != nil {
		log.Panic(err)
	}
	p.Add(line, scatter)

	err = p.Save(100, 100, "testdata/plotLogo.png")
	if err != nil {
		log.Panic(err)
	}
}

func TestMainExample(t *testing.T) {
	cmpimg.CheckPlot(Example, t, "plotLogo.png")
}
