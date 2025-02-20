// Copyright ©2015 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package vgsvg uses svgo (github.com/ajstarks/svgo)
// as a backend for vg.
package vgsvg // import "github.com/gshk/plot/vg/vgsvg"

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"html"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"

	svgo "github.com/ajstarks/svgo"

	"github.com/gshk/plot/vg"
)

// pr is the precision to use when outputting float64s.
const pr = 5

const (
	// DefaultWidth and DefaultHeight are the default canvas
	// dimensions.
	DefaultWidth  = 4 * vg.Inch
	DefaultHeight = 4 * vg.Inch
)

type Canvas struct {
	svg  *svgo.SVG
	w, h vg.Length

	buf   *bytes.Buffer
	stack []context
}

type context struct {
	color      color.Color
	dashArray  []vg.Length
	dashOffset vg.Length
	lineWidth  vg.Length
	gEnds      int
}

type option func(*Canvas)

// UseWH specifies the width and height of the canvas.
func UseWH(w, h vg.Length) option {
	return func(c *Canvas) {
		if w <= 0 || h <= 0 {
			panic("vgsvg: w and h must both be > 0")
		}
		c.w = w
		c.h = h
	}
}

// New returns a new image canvas.
func New(w, h vg.Length) *Canvas {
	return NewWith(UseWH(w, h))
}

// NewWith returns a new image canvas created according to the specified
// options. The currently accepted options is UseWH. If size is not
// specified, the default is used.
func NewWith(opts ...option) *Canvas {
	buf := new(bytes.Buffer)
	c := &Canvas{
		svg:   svgo.New(buf),
		w:     DefaultWidth,
		h:     DefaultHeight,
		buf:   buf,
		stack: []context{{}},
	}

	for _, opt := range opts {
		opt(c)
	}

	// This is like svg.Start, except it uses floats
	// and specifies the units.
	fmt.Fprintf(c.buf, `<?xml version="1.0"?>
<!-- Generated by SVGo and Plotinum VG -->
<svg width="%.*gpt" height="%.*gpt" viewBox="0 0 %.*g %.*g"
	xmlns="http://www.w3.org/2000/svg"
	xmlns:xlink="http://www.w3.org/1999/xlink">`+"\n",
		pr, c.w,
		pr, c.h,
		pr, c.w,
		pr, c.h,
	)

	// Swap the origin to the bottom left.
	// This must be matched with a </g> when saving,
	// before the closing </svg>.
	c.svg.Gtransform(fmt.Sprintf("scale(1, -1) translate(0, -%.*g)", pr, c.h.Points()))

	vg.Initialize(c)
	return c
}

func (c *Canvas) Size() (w, h vg.Length) {
	return c.w, c.h
}

func (c *Canvas) context() *context {
	return &c.stack[len(c.stack)-1]
}

func (c *Canvas) SetLineWidth(w vg.Length) {
	c.context().lineWidth = w
}

func (c *Canvas) SetLineDash(dashes []vg.Length, offs vg.Length) {
	c.context().dashArray = dashes
	c.context().dashOffset = offs
}

func (c *Canvas) SetColor(clr color.Color) {
	c.context().color = clr
}

func (c *Canvas) Rotate(rot float64) {
	rot = rot * 180 / math.Pi
	c.svg.Rotate(rot)
	c.context().gEnds++
}

func (c *Canvas) Translate(pt vg.Point) {
	c.svg.Gtransform(fmt.Sprintf("translate(%.*g, %.*g)", pr, pt.X.Points(), pr, pt.Y.Points()))
	c.context().gEnds++
}

func (c *Canvas) Scale(x, y float64) {
	c.svg.ScaleXY(x, y)
	c.context().gEnds++
}

func (c *Canvas) Push() {
	top := *c.context()
	top.gEnds = 0
	c.stack = append(c.stack, top)
}

func (c *Canvas) Pop() {
	for i := 0; i < c.context().gEnds; i++ {
		c.svg.Gend()
	}
	c.stack = c.stack[:len(c.stack)-1]
}

func (c *Canvas) Stroke(path vg.Path) {
	if c.context().lineWidth.Points() <= 0 {
		return
	}
	c.svg.Path(c.pathData(path),
		style(elm("fill", "#000000", "none"),
			elm("stroke", "none", colorString(c.context().color)),
			elm("stroke-opacity", "1", opacityString(c.context().color)),
			elm("stroke-width", "1", "%.*g", pr, c.context().lineWidth.Points()),
			elm("stroke-dasharray", "none", dashArrayString(c)),
			elm("stroke-dashoffset", "0", "%.*g", pr, c.context().dashOffset.Points())))
}

func (c *Canvas) Fill(path vg.Path) {
	c.svg.Path(c.pathData(path),
		style(elm("fill", "#000000", colorString(c.context().color)),
			elm("fill-opacity", "1", opacityString(c.context().color))))
}

func (c *Canvas) pathData(path vg.Path) string {
	buf := new(bytes.Buffer)
	var x, y float64
	for _, comp := range path {
		switch comp.Type {
		case vg.MoveComp:
			fmt.Fprintf(buf, "M%.*g,%.*g", pr, comp.Pos.X.Points(), pr, comp.Pos.Y.Points())
			x = comp.Pos.X.Points()
			y = comp.Pos.Y.Points()
		case vg.LineComp:
			fmt.Fprintf(buf, "L%.*g,%.*g", pr, comp.Pos.X.Points(), pr, comp.Pos.Y.Points())
			x = comp.Pos.X.Points()
			y = comp.Pos.Y.Points()
		case vg.ArcComp:
			r := comp.Radius.Points()
			x0 := comp.Pos.X.Points() + r*math.Cos(comp.Start)
			y0 := comp.Pos.Y.Points() + r*math.Sin(comp.Start)
			if x0 != x || y0 != y {
				fmt.Fprintf(buf, "L%.*g,%.*g", pr, x0, pr, y0)
			}
			if math.Abs(comp.Angle) >= 2*math.Pi {
				x, y = circle(buf, c, &comp)
			} else {
				x, y = arc(buf, c, &comp)
			}
		case vg.CurveComp:
			switch len(comp.Control) {
			case 1:
				fmt.Fprintf(buf, "Q%.*g,%.*g,%.*g,%.*g",
					pr, comp.Control[0].X.Points(), pr, comp.Control[0].Y.Points(),
					pr, comp.Pos.X.Points(), pr, comp.Pos.Y.Points())
			case 2:
				fmt.Fprintf(buf, "C%.*g,%.*g,%.*g,%.*g,%.*g,%.*g",
					pr, comp.Control[0].X.Points(), pr, comp.Control[0].Y.Points(),
					pr, comp.Control[1].X.Points(), pr, comp.Control[1].Y.Points(),
					pr, comp.Pos.X.Points(), pr, comp.Pos.Y.Points())
			default:
				panic("vgsvg: invalid number of control points")
			}
			x = comp.Pos.X.Points()
			y = comp.Pos.Y.Points()
		case vg.CloseComp:
			buf.WriteString("Z")
		default:
			panic(fmt.Sprintf("Unknown path component type: %d\n", comp.Type))
		}
	}
	return buf.String()
}

// circle adds circle path data to the given writer.
// Circles must be drawn using two arcs because
// SVG disallows the start and end point of an arc
// from being at the same location.
func circle(w io.Writer, c *Canvas, comp *vg.PathComp) (x, y float64) {
	angle := 2 * math.Pi
	if comp.Angle < 0 {
		angle = -2 * math.Pi
	}
	angle += remainder(comp.Angle, 2*math.Pi)
	if angle >= 4*math.Pi {
		panic("Impossible angle")
	}

	r := comp.Radius.Points()
	x0 := comp.Pos.X.Points() + r*math.Cos(comp.Start+angle/2)
	y0 := comp.Pos.Y.Points() + r*math.Sin(comp.Start+angle/2)
	x = comp.Pos.X.Points() + r*math.Cos(comp.Start+angle)
	y = comp.Pos.Y.Points() + r*math.Sin(comp.Start+angle)

	fmt.Fprintf(w, "A%.*g,%.*g 0 %d %d %.*g,%.*g", pr, r, pr, r,
		large(angle/2), sweep(angle/2), pr, x0, pr, y0) //
	fmt.Fprintf(w, "A%.*g,%.*g 0 %d %d %.*g,%.*g", pr, r, pr, r,
		large(angle/2), sweep(angle/2), pr, x, pr, y)
	return
}

// remainder returns the remainder of x/y.
// We don't use math.Remainder because it
// seems to return incorrect values due to how
// IEEE defines the remainder operation…
func remainder(x, y float64) float64 {
	return (x/y - math.Trunc(x/y)) * y
}

// arc adds arc path data to the given writer.
// Arc can only be used if the arc's angle is
// less than a full circle, if it is greater then
// circle should be used instead.
func arc(w io.Writer, c *Canvas, comp *vg.PathComp) (x, y float64) {
	r := comp.Radius.Points()
	x = comp.Pos.X.Points() + r*math.Cos(comp.Start+comp.Angle)
	y = comp.Pos.Y.Points() + r*math.Sin(comp.Start+comp.Angle)
	fmt.Fprintf(w, "A%.*g,%.*g 0 %d %d %.*g,%.*g", pr, r, pr, r,
		large(comp.Angle), sweep(comp.Angle), pr, x, pr, y)
	return
}

// sweep returns the arc sweep flag value for
// the given angle.
func sweep(a float64) int {
	if a < 0 {
		return 0
	}
	return 1
}

// large returns the arc's large flag value for
// the given angle.
func large(a float64) int {
	if math.Abs(a) >= math.Pi {
		return 1
	}
	return 0
}

// FillString draws str at position pt using the specified font.
// Text passed to FillString is escaped with html.EscapeString.
func (c *Canvas) FillString(font vg.Font, pt vg.Point, str string) {
	fontStr, ok := fontMap[font.Name()]
	if !ok {
		panic(fmt.Sprintf("Unknown font: %s", font.Name()))
	}
	sty := style(fontStr,
		elm("font-size", "medium", "%.*gpx", pr, font.Size.Points()),
		elm("fill", "#000000", colorString(c.context().color)))
	if sty != "" {
		sty = "\n\t" + sty
	}
	fmt.Fprintf(c.buf, `<text x="%.*g" y="%.*g" transform="scale(1, -1)"%s>%s</text>`+"\n",
		pr, pt.X.Points(), pr, -pt.Y.Points(), sty, html.EscapeString(str))
}

// DrawImage implements the vg.Canvas.DrawImage method.
func (c *Canvas) DrawImage(rect vg.Rectangle, img image.Image) {
	buf := new(bytes.Buffer)
	err := png.Encode(buf, img)
	if err != nil {
		panic(fmt.Errorf("vgsvg: error encoding image to PNG: %v\n", err))
	}
	str := "data:image/jpg;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
	rsz := rect.Size()
	min := rect.Min
	var (
		width  = rsz.X.Points()
		height = rsz.Y.Points()
		xmin   = min.X.Points()
		ymin   = min.Y.Points()
	)
	fmt.Fprintf(
		c.buf,
		`<image x="%v" y="%v" width="%v" height="%v" xlink:href="%s" %s />`+"\n",
		xmin,
		-ymin-height,
		width,
		height,
		str,
		// invert y so image is not upside-down
		`transform="scale(1, -1)"`,
	)
}

var (
	// fontMap maps Postscript-style font names to their
	// corresponding SVG style string.
	fontMap = map[string]string{
		"Courier":               "font-family:Courier;font-weight:normal;font-style:normal",
		"Courier-Bold":          "font-family:Courier;font-weight:bold;font-style:normal",
		"Courier-Oblique":       "font-family:Courier;font-weight:normal;font-style:oblique",
		"Courier-BoldOblique":   "font-family:Courier;font-weight:bold;font-style:oblique",
		"Helvetica":             "font-family:Helvetica;font-weight:normal;font-style:normal",
		"Helvetica-Bold":        "font-family:Helvetica;font-weight:bold;font-style:normal",
		"Helvetica-Oblique":     "font-family:Helvetica;font-weight:normal;font-style:oblique",
		"Helvetica-BoldOblique": "font-family:Helvetica;font-weight:bold;font-style:oblique",
		"Times-Roman":           "font-family:Times;font-weight:normal;font-style:normal",
		"Times-Bold":            "font-family:Times;font-weight:bold;font-style:normal",
		"Times-Italic":          "font-family:Times;font-weight:normal;font-style:italic",
		"Times-BoldItalic":      "font-family:Times;font-weight:bold;font-style:italic",
	}
)

// WriteTo writes the canvas to an io.Writer.
func (c *Canvas) WriteTo(w io.Writer) (int64, error) {
	b := bufio.NewWriter(w)
	n, err := c.buf.WriteTo(b)
	if err != nil {
		return n, err
	}

	// Close the groups and svg in the output buffer
	// so that the Canvas is not closed and can be
	// used again if needed.
	for i := 0; i < c.nEnds(); i++ {
		m, err := fmt.Fprintln(b, "</g>")
		n += int64(m)
		if err != nil {
			return n, err
		}
	}

	m, err := fmt.Fprintln(b, "</svg>")
	n += int64(m)
	if err != nil {
		return n, err
	}

	return n, b.Flush()
}

// nEnds returns the number of group ends
// needed before the SVG is saved.
func (c *Canvas) nEnds() int {
	n := 1 // close the transform that moves the origin
	for _, ctx := range c.stack {
		n += ctx.gEnds
	}
	return n
}

// style returns a style string composed of
// all of the given elements.  If the elements
// are all empty then the empty string is
// returned.
func style(elms ...string) string {
	str := ""
	for _, e := range elms {
		if e == "" {
			continue
		}
		if str != "" {
			str += ";"
		}
		str += e
	}
	if str == "" {
		return ""
	}
	return "style=\"" + str + "\""
}

// elm returns a style element string with the
// given key and value.  If the value matches
// default then the empty string is returned.
func elm(key, def, f string, vls ...interface{}) string {
	value := fmt.Sprintf(f, vls...)
	if value == def {
		return ""
	}
	return key + ":" + value
}

// dashArrayString returns a string representing the
// dash array specification.
func dashArrayString(c *Canvas) string {
	str := ""
	for i, d := range c.context().dashArray {
		str += fmt.Sprintf("%.*g", pr, d.Points())
		if i < len(c.context().dashArray)-1 {
			str += ","
		}
	}
	if str == "" {
		str = "none"
	}
	return str
}

// colorString returns the hexadecimal string representation of the color
func colorString(clr color.Color) string {
	if clr == nil {
		clr = color.Black
	}
	r, g, b, _a := clr.RGBA()
	a := 255.0 / float64(_a)
	return fmt.Sprintf("#%02X%02X%02X", int(float64(r)*a),
		int(float64(g)*a), int(float64(b)*a))
}

// opacityString returns the opacity value of the given color.
func opacityString(clr color.Color) string {
	if clr == nil {
		clr = color.Black
	}
	_, _, _, a := clr.RGBA()
	return fmt.Sprintf("%.*g", pr, float64(a)/math.MaxUint16)
}
