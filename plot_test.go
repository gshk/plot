// Copyright ©2015 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plot_test

import (
	"bytes"
	"fmt"
	"image/color"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/gshk/plot"
	"github.com/gshk/plot/plotter"
	"github.com/gshk/plot/vg"
	"github.com/gshk/plot/vg/draw"
	"github.com/gshk/plot/vg/recorder"
)

func TestLegendAlignment(t *testing.T) {
	const fontSize = 10.189054726368159 // This font size gives an entry height of 10.
	font, err := vg.MakeFont(plot.DefaultFont, fontSize)
	if err != nil {
		t.Fatalf("failed to create font: %v", err)
	}
	l := plot.Legend{
		ThumbnailWidth: vg.Points(20),
		TextStyle:      draw.TextStyle{Font: font},
	}
	for _, n := range []string{"A", "B", "C", "D"} {
		b, err := plotter.NewBarChart(plotter.Values{0}, 1)
		if err != nil {
			t.Fatalf("failed to create bar chart %q: %v", n, err)
		}
		l.Add(n, b)
	}

	var r recorder.Canvas
	c := draw.NewCanvas(&r, 100, 100)
	l.Draw(draw.Canvas{
		Canvas: c.Canvas,
		Rectangle: vg.Rectangle{
			Min: vg.Point{X: 0, Y: 0},
			Max: vg.Point{X: 100, Y: 100},
		},
	})

	got := r.Actions

	// want is a snapshot of the actions for the code above when the
	// graphical output has been visually confirmed to be correct for
	// the bar charts example show in gonum/plot#25.
	want := []recorder.Action{
		&recorder.SetColor{
			Color: color.Gray16{},
		},
		&recorder.Fill{
			Path: vg.Path{
				{Type: vg.MoveComp, Pos: vg.Point{X: 80, Y: 30}},
				{Type: vg.LineComp, Pos: vg.Point{X: 80, Y: 40}},
				{Type: vg.LineComp, Pos: vg.Point{X: 100, Y: 40}},
				{Type: vg.LineComp, Pos: vg.Point{X: 100, Y: 30}},
				{Type: vg.CloseComp},
			},
		},
		&recorder.SetColor{
			Color: color.Gray16{},
		},
		&recorder.SetLineWidth{
			Width: 1,
		},
		&recorder.SetLineDash{},
		&recorder.Stroke{
			Path: vg.Path{
				{Type: vg.MoveComp, Pos: vg.Point{X: 80, Y: 30}},
				{Type: vg.LineComp, Pos: vg.Point{X: 80, Y: 40}},
				{Type: vg.LineComp, Pos: vg.Point{X: 100, Y: 40}},
				{Type: vg.LineComp, Pos: vg.Point{X: 100, Y: 30}},
				{Type: vg.LineComp, Pos: vg.Point{X: 80, Y: 30}},
			},
		},
		&recorder.SetColor{},
		&recorder.FillString{
			Font:   string("Times-Roman"),
			Size:   fontSize,
			Point:  vg.Point{X: 70.09452736318407, Y: 30.18905472636816},
			String: "A",
		},
		&recorder.SetColor{
			Color: color.Gray16{},
		},
		&recorder.Fill{
			Path: vg.Path{
				{Type: vg.MoveComp, Pos: vg.Point{X: 80, Y: 20}},
				{Type: vg.LineComp, Pos: vg.Point{X: 80, Y: 30}},
				{Type: vg.LineComp, Pos: vg.Point{X: 100, Y: 30}},
				{Type: vg.LineComp, Pos: vg.Point{X: 100, Y: 20}},
				{Type: vg.CloseComp},
			},
		},
		&recorder.SetColor{
			Color: color.Gray16{},
		},
		&recorder.SetLineWidth{
			Width: 1,
		},
		&recorder.SetLineDash{},
		&recorder.Stroke{
			Path: vg.Path{
				{Type: vg.MoveComp, Pos: vg.Point{X: 80, Y: 20}},
				{Type: vg.LineComp, Pos: vg.Point{X: 80, Y: 30}},
				{Type: vg.LineComp, Pos: vg.Point{X: 100, Y: 30}},
				{Type: vg.LineComp, Pos: vg.Point{X: 100, Y: 20}},
				{Type: vg.LineComp, Pos: vg.Point{X: 80, Y: 20}},
			},
		},
		&recorder.SetColor{},
		&recorder.FillString{
			Font:   string("Times-Roman"),
			Size:   fontSize,
			Point:  vg.Point{X: 70.65671641791045, Y: 20.18905472636816},
			String: "B",
		},
		&recorder.SetColor{
			Color: color.Gray16{
				Y: uint16(0),
			},
		},
		&recorder.Fill{
			Path: vg.Path{
				{Type: vg.MoveComp, Pos: vg.Point{X: 80, Y: 10}},
				{Type: vg.LineComp, Pos: vg.Point{X: 80, Y: 20}},
				{Type: vg.LineComp, Pos: vg.Point{X: 100, Y: 20}},
				{Type: vg.LineComp, Pos: vg.Point{X: 100, Y: 10}},
				{Type: vg.CloseComp},
			},
		},
		&recorder.SetColor{
			Color: color.Gray16{},
		},
		&recorder.SetLineWidth{
			Width: 1,
		},
		&recorder.SetLineDash{},
		&recorder.Stroke{
			Path: vg.Path{
				{Type: vg.MoveComp, Pos: vg.Point{X: 80, Y: 10}},
				{Type: vg.LineComp, Pos: vg.Point{X: 80, Y: 20}},
				{Type: vg.LineComp, Pos: vg.Point{X: 100, Y: 20}},
				{Type: vg.LineComp, Pos: vg.Point{X: 100, Y: 10}},
				{Type: vg.LineComp, Pos: vg.Point{X: 80, Y: 10}},
			},
		},
		&recorder.SetColor{},
		&recorder.FillString{
			Font:   string("Times-Roman"),
			Size:   fontSize,
			Point:  vg.Point{X: 70.65671641791045, Y: 10.189054726368159},
			String: "C",
		},
		&recorder.SetColor{
			Color: color.Gray16{},
		},
		&recorder.Fill{
			Path: vg.Path{
				{Type: vg.MoveComp, Pos: vg.Point{X: 80, Y: 0}},
				{Type: vg.LineComp, Pos: vg.Point{X: 80, Y: 10}},
				{Type: vg.LineComp, Pos: vg.Point{X: 100, Y: 10}},
				{Type: vg.LineComp, Pos: vg.Point{X: 100, Y: 0}},
				{Type: vg.CloseComp},
			},
		},
		&recorder.SetColor{
			Color: color.Gray16{},
		},
		&recorder.SetLineWidth{
			Width: 1,
		},
		&recorder.SetLineDash{},
		&recorder.Stroke{
			Path: vg.Path{
				{Type: vg.MoveComp, Pos: vg.Point{X: 80, Y: 0}},
				{Type: vg.LineComp, Pos: vg.Point{X: 80, Y: 10}},
				{Type: vg.LineComp, Pos: vg.Point{X: 100, Y: 10}},
				{Type: vg.LineComp, Pos: vg.Point{X: 100, Y: 0}},
				{Type: vg.LineComp, Pos: vg.Point{X: 80, Y: 0}},
			},
		},
		&recorder.SetColor{},
		&recorder.FillString{
			Font:   string("Times-Roman"),
			Size:   fontSize,
			Point:  vg.Point{X: 70.09452736318407, Y: 0.189054726368159},
			String: "D",
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected legend actions:\ngot:\n%s\nwant:\n%s", formatActions(got), formatActions(want))
		t.Errorf("First diff:\n%s", printFirstDiff(got, want))
	}
}

func formatActions(actions []recorder.Action) string {
	var buf bytes.Buffer
	for _, a := range actions {
		fmt.Fprintf(&buf, "\t%s\n", a.Call())
	}
	return buf.String()
}

// printFirstDiff prints the first line that is different between two actions.
func printFirstDiff(got, want []recorder.Action) string {
	var buf bytes.Buffer
	for i, g := range got {
		if i >= len(want) {
			fmt.Fprintf(&buf, "line %d:\n\tgot: %s\n\twant is empty", i, g.Call())
			break
		}
		w := want[i]
		if w.Call() != g.Call() {
			fmt.Fprintf(&buf, "line %d:\n\tgot: %s\n\twant: %s", i, g.Call(), w.Call())
			break
		}
	}
	if len(want) > len(got) {
		fmt.Fprintf(&buf, "line %d:\n\twant: %s\n\tgot is empty", len(got), want[len(got)].Call())
	}
	return buf.String()
}

func TestIssue514(t *testing.T) {
	for _, ulp := range []int{
		0,
		+1, +2, +3, +4, +5, +6, +7, +8, +9, +10, +11, +12, +13, +14, +15, +16, +17, +18, +19, +20, +21, +22,
		-1, -2, -3, -4, -5, -6, -7, -8, -9, -10, -11, -12, -13, -14, -15, -16, -17, -18, -19, -20, -21, -22,
	} {
		t.Run(fmt.Sprintf("ulps%+02d", ulp), func(t *testing.T) {
			done := make(chan int)
			go func() {
				defer close(done)

				p, err := plot.New()
				if err != nil {
					t.Fatalf("could not create plot: %v", err)
				}

				var (
					y1 = 100.0
					y2 = y1
				)

				switch {
				case ulp < 0:
					y2 = math.Float64frombits(math.Float64bits(y1) - uint64(-ulp))
				case ulp > 0:
					y2 = math.Float64frombits(math.Float64bits(y1) + uint64(ulp))
				}

				pts, err := plotter.NewScatter(plotter.XYs{
					{X: 1, Y: y1},
					{X: 1, Y: y2},
				})
				if err != nil {
					t.Fatalf("could not create scatter: %v", err)
				}

				p.Add(pts)

				c := draw.NewCanvas(&recorder.Canvas{}, 100, 100)
				p.Draw(c)
			}()

			timeout := time.NewTimer(100 * time.Millisecond)
			defer timeout.Stop()

			select {
			case <-done:
			case <-timeout.C:
				t.Fatalf("could not create plot with small axis range within allotted time")
			}
		})
	}
}
