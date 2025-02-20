// Copyright ©2016 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vgtex_test

import (
	"log"
	"os"
	"testing"

	"github.com/gshk/plot"
	"github.com/gshk/plot/cmpimg"
	"github.com/gshk/plot/plotter"
	"github.com/gshk/plot/vg"
	"github.com/gshk/plot/vg/draw"
	"github.com/gshk/plot/vg/vgtex"
)

// An example of making a LaTeX plot.
func Example() {
	scatter, err := plotter.NewScatter(plotter.XYs{{1, 1}, {0, 1}, {0, 0}})
	if err != nil {
		log.Fatal(err)
	}
	p, err := plot.New()
	if err != nil {
		log.Fatal(err)
	}
	p.Add(scatter)
	// p.HideAxes()
	p.Title.Text = `A scatter plot: $\sqrt{\frac{e^{3i\pi}}{2\cos 3\pi}}$`
	p.X.Label.Text = `$x = \eta$`
	p.Y.Label.Text = `$y$ is some $\Phi$`

	c := vgtex.NewDocument(5*vg.Centimeter, 5*vg.Centimeter)
	p.Draw(draw.New(c))
	c.FillString(p.Title.Font, vg.Point{2.5 * vg.Centimeter, 2.5 * vg.Centimeter}, "x")

	f, err := os.Create("testdata/scatter.tex")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if _, err = c.WriteTo(f); err != nil {
		log.Fatal(err)
	}
}

func TestTexCanvas(t *testing.T) {
	cmpimg.CheckPlot(Example, t, "scatter.tex")
}
