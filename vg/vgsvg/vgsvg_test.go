// Copyright ©2019 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vgsvg_test

import (
	"bytes"
	"io/ioutil"
	"log"
	"testing"

	"github.com/gshk/plot"
	"github.com/gshk/plot/cmpimg"
	"github.com/gshk/plot/plotter"
	"github.com/gshk/plot/vg"
	"github.com/gshk/plot/vg/draw"
	"github.com/gshk/plot/vg/vgsvg"
)

func Example() {
	p, err := plot.New()
	if err != nil {
		log.Fatalf("could not create plot: %v", err)
	}
	p.Title.Text = "Scatter plot"
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"

	scatter, err := plotter.NewScatter(plotter.XYs{{1, 1}, {0, 1}, {0, 0}})
	if err != nil {
		log.Fatalf("could not create scatter: %v", err)
	}
	p.Add(scatter)

	err = p.Save(5*vg.Centimeter, 5*vg.Centimeter, "testdata/scatter.svg")
	if err != nil {
		log.Fatalf("could not save SVG plot: %v", err)
	}
}

func TestSVG(t *testing.T) {
	cmpimg.CheckPlot(Example, t, "scatter.svg")
}

func TestNewWith(t *testing.T) {
	p, err := plot.New()
	if err != nil {
		t.Fatalf("could not create plot: %v", err)
	}
	p.Title.Text = "Scatter plot"
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"

	scatter, err := plotter.NewScatter(plotter.XYs{{1, 1}, {0, 1}, {0, 0}})
	if err != nil {
		t.Fatalf("could not create scatter: %v", err)
	}
	p.Add(scatter)

	c := vgsvg.NewWith(vgsvg.UseWH(5*vg.Centimeter, 5*vg.Centimeter))
	p.Draw(draw.New(c))

	b := new(bytes.Buffer)
	if _, err = c.WriteTo(b); err != nil {
		t.Fatal(err)
	}

	want, err := ioutil.ReadFile("testdata/scatter_golden.svg")
	if err != nil {
		t.Fatal(err)
	}

	ok, err := cmpimg.Equal("svg", b.Bytes(), want)
	if err != nil {
		t.Fatalf("could not compare images: %v", err)
	}
	if !ok {
		t.Fatalf("images differ:\ngot:\n%s\nwant:\n%s\n", b.Bytes(), want)
	}
}

func TestHtmlEscape(t *testing.T) {
	p, err := plot.New()
	if err != nil {
		t.Fatalf("could not create plot: %v", err)
	}
	p.Title.Text = "Scatter & line plot"
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"

	scatter, err := plotter.NewScatter(plotter.XYs{{1, 1}, {0, 1}, {0, 0}})
	if err != nil {
		t.Fatalf("could not create scatter: %v", err)
	}
	p.Add(scatter)

	line, err := plotter.NewLine(plotter.XYs{{1, 1}, {0, 1}, {0, 0}})
	if err != nil {
		t.Fatalf("could not create scatter: %v", err)
	}
	line.Width = 0.5
	p.Add(line)

	c := vgsvg.NewWith(vgsvg.UseWH(5*vg.Centimeter, 5*vg.Centimeter))
	p.Draw(draw.New(c))

	b := new(bytes.Buffer)
	if _, err = c.WriteTo(b); err != nil {
		t.Fatal(err)
	}

	want, err := ioutil.ReadFile("testdata/scatter_line_golden.svg")
	if err != nil {
		t.Fatal(err)
	}

	ok, err := cmpimg.Equal("svg", b.Bytes(), want)
	if err != nil {
		t.Fatalf("could not compare images: %v", err)
	}
	if !ok {
		t.Fatalf("images differ:\ngot:\n%s\nwant:\n%s\n", b.Bytes(), want)
	}
}
