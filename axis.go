// Copyright ©2015 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package plot

import (
	"fmt"
	"image/color"
	"math"
	"strings"
	"time"

	"github.com/gshk/plot/vg"
	"github.com/gshk/plot/vg/draw"
)

// Ticker creates Ticks in a specified range
type Ticker interface {
	// Ticks returns Ticks in a specified range
	Ticks(min, max float64) []Tick
}

// Normalizer rescales values from the data coordinate system to the
// normalized coordinate system.
type Normalizer interface {
	// Normalize transforms a value x in the data coordinate system to
	// the normalized coordinate system.
	Normalize(min, max, x float64) float64
}

// An Axis represents either a horizontal or vertical
// axis of a plot.
type Axis struct {
	// Min and Max are the minimum and maximum data
	// values represented by the axis.
	Min, Max float64

	Label struct {
		// Text is the axis label string.
		Text string

		// TextStyle is the style of the axis label text.
		// For the vertical axis, one quarter turn
		// counterclockwise will be added to the label
		// text before drawing.
		draw.TextStyle
	}

	// LineStyle is the style of the axis line.
	draw.LineStyle

	// Padding between the axis line and the data.  Having
	// non-zero padding ensures that the data is never drawn
	// on the axis, thus making it easier to see.
	Padding vg.Length

	Tick struct {
		// Label is the TextStyle on the tick labels.
		Label draw.TextStyle

		// LineStyle is the LineStyle of the tick lines.
		draw.LineStyle

		// Length is the length of a major tick mark.
		// Minor tick marks are half of the length of major
		// tick marks.
		Length vg.Length

		// Marker returns the tick marks.  Any tick marks
		// returned by the Marker function that are not in
		// range of the axis are not drawn.
		Marker Ticker
	}

	// Scale transforms a value given in the data coordinate system
	// to the normalized coordinate system of the axis—its distance
	// along the axis as a fraction of the axis range.
	Scale Normalizer
}

// makeAxis returns a default Axis.
//
// The default range is (∞, ­∞), and thus any finite
// value is less than Min and greater than Max.
func makeAxis(orientation bool) (Axis, error) {
	labelFont, err := vg.MakeFont(DefaultFont, vg.Points(12))
	if err != nil {
		return Axis{}, err
	}

	tickFont, err := vg.MakeFont(DefaultFont, vg.Points(10))
	if err != nil {
		return Axis{}, err
	}

	a := Axis{
		Min: math.Inf(+1),
		Max: math.Inf(-1),
		LineStyle: draw.LineStyle{
			Color: color.Black,
			Width: vg.Points(0.5),
		},
		Padding: vg.Points(5),
		Scale:   LinearScale{},
	}
	a.Label.TextStyle = draw.TextStyle{
		Color:  color.Black,
		Font:   labelFont,
		XAlign: draw.XCenter,
		YAlign: draw.YBottom,
	}
	var xalign, yalign = draw.XCenter, draw.YTop
	if orientation == vertical {
		xalign, yalign = draw.XRight, draw.YCenter
	}
	a.Tick.Label = draw.TextStyle{
		Color:  color.Black,
		Font:   tickFont,
		XAlign: xalign,
		YAlign: yalign,
	}
	a.Tick.LineStyle = draw.LineStyle{
		Color: color.Black,
		Width: vg.Points(0.5),
	}
	a.Tick.Length = vg.Points(8)
	a.Tick.Marker = DefaultTicks{}

	return a, nil
}

// sanitizeRange ensures that the range of the
// axis makes sense.
func (a *Axis) sanitizeRange() {
	if math.IsInf(a.Min, 0) || math.IsNaN(a.Min) {
		a.Min = 0
	}
	if math.IsInf(a.Max, 0) || math.IsNaN(a.Max) {
		a.Max = 0
	}
	if a.Min > a.Max {
		a.Min, a.Max = a.Max, a.Min
	}
	if a.Min == a.Max {
		a.Min--
		a.Max++
	}
}

// LinearScale an be used as the value of an Axis.Scale function to
// set the axis to a standard linear scale.
type LinearScale struct{}

var _ Normalizer = LinearScale{}

// Normalize returns the fractional distance of x between min and max.
func (LinearScale) Normalize(min, max, x float64) float64 {
	return (x - min) / (max - min)
}

// LogScale can be used as the value of an Axis.Scale function to
// set the axis to a log scale.
type LogScale struct{}

var _ Normalizer = LogScale{}

// Normalize returns the fractional logarithmic distance of
// x between min and max.
func (LogScale) Normalize(min, max, x float64) float64 {
	if min <= 0 || max <= 0 || x <= 0 {
		panic("Values must be greater than 0 for a log scale.")
	}
	logMin := math.Log(min)
	return (math.Log(x) - logMin) / (math.Log(max) - logMin)
}

// InvertedScale can be used as the value of an Axis.Scale function to
// invert the axis using any Normalizer.
type InvertedScale struct{ Normalizer }

var _ Normalizer = InvertedScale{}

// Normalize returns a normalized [0, 1] value for the position of x.
func (is InvertedScale) Normalize(min, max, x float64) float64 {
	return is.Normalizer.Normalize(max, min, x)
}

// Norm returns the value of x, given in the data coordinate
// system, normalized to its distance as a fraction of the
// range of this axis.  For example, if x is a.Min then the return
// value is 0, and if x is a.Max then the return value is 1.
func (a Axis) Norm(x float64) float64 {
	return a.Scale.Normalize(a.Min, a.Max, x)
}

// drawTicks returns true if the tick marks should be drawn.
func (a Axis) drawTicks() bool {
	return a.Tick.Width > 0 && a.Tick.Length > 0
}

// A horizontalAxis draws horizontally across the bottom
// of a plot.
type horizontalAxis struct {
	Axis
}

// size returns the height of the axis.
func (a horizontalAxis) size() (h vg.Length) {
	if a.Label.Text != "" { // We assume that the label isn't rotated.
		h -= a.Label.Font.Extents().Descent
		h += a.Label.Height(a.Label.Text)
	}

	marks := a.Tick.Marker.Ticks(a.Min, a.Max)
	if len(marks) > 0 {
		if a.drawTicks() {
			h += a.Tick.Length
		}
		h += tickLabelHeight(a.Tick.Label, marks)
	}
	h += a.Width / 2
	h += a.Padding

	return h
}

// draw draws the axis along the lower edge of a draw.Canvas.
func (a horizontalAxis) draw(c draw.Canvas) {
	y := c.Min.Y
	if a.Label.Text != "" {
		y -= a.Label.Font.Extents().Descent
		c.FillText(a.Label.TextStyle, vg.Point{X: c.Center().X, Y: y}, a.Label.Text)
		y += a.Label.Height(a.Label.Text)
	}

	marks := a.Tick.Marker.Ticks(a.Min, a.Max)
	ticklabelheight := tickLabelHeight(a.Tick.Label, marks)
	for _, t := range marks {
		x := c.X(a.Norm(t.Value))
		if !c.ContainsX(x) || t.IsMinor() {
			continue
		}
		c.FillText(a.Tick.Label, vg.Point{X: x, Y: y + ticklabelheight}, t.Label)
	}

	if len(marks) > 0 {
		y += ticklabelheight
	} else {
		y += a.Width / 2
	}

	if len(marks) > 0 && a.drawTicks() {
		len := a.Tick.Length
		for _, t := range marks {
			x := c.X(a.Norm(t.Value))
			if !c.ContainsX(x) {
				continue
			}
			start := t.lengthOffset(len)
			c.StrokeLine2(a.Tick.LineStyle, x, y+start, x, y+len)
		}
		y += len
	}

	c.StrokeLine2(a.LineStyle, c.Min.X, y, c.Max.X, y)
}

// GlyphBoxes returns the GlyphBoxes for the tick labels.
func (a horizontalAxis) GlyphBoxes(*Plot) []GlyphBox {
	var boxes []GlyphBox
	for _, t := range a.Tick.Marker.Ticks(a.Min, a.Max) {
		if t.IsMinor() {
			continue
		}
		box := GlyphBox{
			X:         a.Norm(t.Value),
			Rectangle: a.Tick.Label.Rectangle(t.Label),
		}
		boxes = append(boxes, box)
	}
	return boxes
}

// A verticalAxis is drawn vertically up the left side of a plot.
type verticalAxis struct {
	Axis
}

// size returns the width of the axis.
func (a verticalAxis) size() (w vg.Length) {
	if a.Label.Text != "" { // We assume that the label isn't rotated.
		w -= a.Label.Font.Extents().Descent
		w += a.Label.Height(a.Label.Text)
	}

	marks := a.Tick.Marker.Ticks(a.Min, a.Max)
	if len(marks) > 0 {
		if lwidth := tickLabelWidth(a.Tick.Label, marks); lwidth > 0 {
			w += lwidth
			w += a.Label.Width(" ")
		}
		if a.drawTicks() {
			w += a.Tick.Length
		}
	}
	w += a.Width / 2
	w += a.Padding

	return w
}

// draw draws the axis along the left side of a draw.Canvas.
func (a verticalAxis) draw(c draw.Canvas) {
	x := c.Min.X
	if a.Label.Text != "" {
		sty := a.Label.TextStyle
		sty.Rotation += math.Pi / 2
		x += a.Label.Height(a.Label.Text)
		c.FillText(sty, vg.Point{X: x, Y: c.Center().Y}, a.Label.Text)
		x += -a.Label.Font.Extents().Descent
	}
	marks := a.Tick.Marker.Ticks(a.Min, a.Max)
	if w := tickLabelWidth(a.Tick.Label, marks); len(marks) > 0 && w > 0 {
		x += w
	}

	major := false
	for _, t := range marks {
		y := c.Y(a.Norm(t.Value))
		if !c.ContainsY(y) || t.IsMinor() {
			continue
		}
		c.FillText(a.Tick.Label, vg.Point{X: x, Y: y}, t.Label)
		major = true
	}
	if major {
		x += a.Tick.Label.Width(" ")
	}
	if a.drawTicks() && len(marks) > 0 {
		len := a.Tick.Length
		for _, t := range marks {
			y := c.Y(a.Norm(t.Value))
			if !c.ContainsY(y) {
				continue
			}
			start := t.lengthOffset(len)
			c.StrokeLine2(a.Tick.LineStyle, x+start, y, x+len, y)
		}
		x += len
	}

	c.StrokeLine2(a.LineStyle, x, c.Min.Y, x, c.Max.Y)
}

// GlyphBoxes returns the GlyphBoxes for the tick labels
func (a verticalAxis) GlyphBoxes(*Plot) []GlyphBox {
	var boxes []GlyphBox
	for _, t := range a.Tick.Marker.Ticks(a.Min, a.Max) {
		if t.IsMinor() {
			continue
		}
		box := GlyphBox{
			Y:         a.Norm(t.Value),
			Rectangle: a.Tick.Label.Rectangle(t.Label),
		}
		boxes = append(boxes, box)
	}
	return boxes
}

// DefaultTicks is suitable for the Tick.Marker field of an Axis,
// it returns a reasonable default set of tick marks.
type DefaultTicks struct {
	SuggestedTick int
}

var _ Ticker = DefaultTicks{}

// Ticks returns Ticks in the specified range.
func (d DefaultTicks) Ticks(min, max float64) []Tick {
	if max <= min {
		panic("illegal range")
	}

	if d.SuggestedTick == 0 {
		d.SuggestedTick = 3
	}

	labels, step, q, mag := talbotLinHanrahan(min, max, d.SuggestedTick, withinData, nil, nil, nil)
	majorDelta := step * math.Pow10(mag)
	if q == 0 {
		// Simple fall back was chosen, so
		// majorDelta is the label distance.
		majorDelta = labels[1] - labels[0]
	}

	// Choose a reasonable, but ad
	// hoc formatting for labels.
	var off int
	if mag < -1 || 6 < mag {
		off = 1
	}
	if math.Trunc(q) != q {
		off += 2
	}
	prec := minInt(6, maxInt(off, -mag))
	var ticks []Tick
	for _, v := range labels {
		ticks = append(ticks, Tick{Value: v, Label: formatFloatTick(v, prec)})
	}

	var minorDelta float64
	// See talbotLinHanrahan for the values used here.
	switch step {
	case 1, 2.5:
		minorDelta = majorDelta / 5
	case 2, 3, 4, 5:
		minorDelta = majorDelta / step
	default:
		if majorDelta/2 < dlamchP {
			return ticks
		}
		minorDelta = majorDelta / 2
	}

	// Find the first minor tick not greater
	// than the lowest data value.
	var i float64
	for labels[0]+(i-1)*minorDelta > min {
		i--
	}
	// Add ticks at minorDelta intervals when
	// they are not within minorDelta/2 of a
	// labelled tick.
	for {
		val := labels[0] + i*minorDelta
		if val > max {
			break
		}
		found := false
		for _, t := range ticks {
			if math.Abs(t.Value-val) < minorDelta/2 {
				found = true
			}
		}
		if !found {
			ticks = append(ticks, Tick{Value: val})
		}
		i++
	}

	return ticks
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// LogTicks is suitable for the Tick.Marker field of an Axis,
// it returns tick marks suitable for a log-scale axis.
type LogTicks struct{}

var _ Ticker = LogTicks{}

// Ticks returns Ticks in a specified range
func (LogTicks) Ticks(min, max float64) []Tick {
	if min <= 0 || max <= 0 {
		panic("Values must be greater than 0 for a log scale.")
	}

	val := math.Pow10(int(math.Log10(min)))
	max = math.Pow10(int(math.Ceil(math.Log10(max))))
	var ticks []Tick
	for val < max {
		for i := 1; i < 10; i++ {
			if i == 1 {
				ticks = append(ticks, Tick{Value: val, Label: formatFloatTick(val, -1)})
			}
			ticks = append(ticks, Tick{Value: val * float64(i)})
		}
		val *= 10
	}
	ticks = append(ticks, Tick{Value: val, Label: formatFloatTick(val, -1)})

	return ticks
}

// ConstantTicks is suitable for the Tick.Marker field of an Axis.
// This function returns the given set of ticks.
type ConstantTicks []Tick

var _ Ticker = ConstantTicks{}

// Ticks returns Ticks in a specified range
func (ts ConstantTicks) Ticks(float64, float64) []Tick {
	return ts
}

// UnixTimeIn returns a time conversion function for the given location.
func UnixTimeIn(loc *time.Location) func(t float64) time.Time {
	return func(t float64) time.Time {
		return time.Unix(int64(t), 0).In(loc)
	}
}

// UTCUnixTime is the default time conversion for TimeTicks.
var UTCUnixTime = UnixTimeIn(time.UTC)

// TimeTicks is suitable for axes representing time values.
type TimeTicks struct {
	// Ticker is used to generate a set of ticks.
	// If nil, DefaultTicks will be used.
	Ticker Ticker

	// Format is the textual representation of the time value.
	// If empty, time.RFC3339 will be used
	Format string

	// Time takes a float64 value and converts it into a time.Time.
	// If nil, UTCUnixTime is used.
	Time func(t float64) time.Time
}

var _ Ticker = TimeTicks{}

// Ticks implements plot.Ticker.
func (t TimeTicks) Ticks(min, max float64) []Tick {
	if t.Ticker == nil {
		t.Ticker = DefaultTicks{}
	}
	if t.Format == "" {
		t.Format = time.RFC3339
	}
	if t.Time == nil {
		t.Time = UTCUnixTime
	}

	ticks := t.Ticker.Ticks(min, max)
	for i := range ticks {
		tick := &ticks[i]
		if tick.Label == "" {
			continue
		}
		tick.Label = t.Time(tick.Value).Format(t.Format)
	}
	return ticks
}

// A Tick is a single tick mark on an axis.
type Tick struct {
	// Value is the data value marked by this Tick.
	Value float64

	// Label is the text to display at the tick mark.
	// If Label is an empty string then this is a minor
	// tick mark.
	Label string
}

// IsMinor returns true if this is a minor tick mark.
func (t Tick) IsMinor() bool {
	return t.Label == ""
}

// lengthOffset returns an offset that should be added to the
// tick mark's line to accout for its length.  I.e., the start of
// the line for a minor tick mark must be shifted by half of
// the length.
func (t Tick) lengthOffset(len vg.Length) vg.Length {
	if t.IsMinor() {
		return len / 2
	}
	return 0
}

// tickLabelHeight returns height of the tick mark labels.
func tickLabelHeight(sty draw.TextStyle, ticks []Tick) vg.Length {
	maxHeight := vg.Length(0)
	for _, t := range ticks {
		if t.IsMinor() {
			continue
		}
		r := sty.Rectangle(t.Label)
		h := r.Max.Y - r.Min.Y
		if h > maxHeight {
			maxHeight = h
		}
	}
	return maxHeight
}

// tickLabelWidth returns the width of the widest tick mark label.
func tickLabelWidth(sty draw.TextStyle, ticks []Tick) vg.Length {
	maxWidth := vg.Length(0)
	for _, t := range ticks {
		if t.IsMinor() {
			continue
		}
		r := sty.Rectangle(t.Label)
		w := r.Max.X - r.Min.X
		if w > maxWidth {
			maxWidth = w
		}
	}
	return maxWidth
}

// formatFloatTick returns a g-formated string representation of v
// to the specified precision.
func formatFloatTick(v float64, prec int) string {
	if prec < 1 {
		prec = 1
	}
	s := fmt.Sprintf("%."+fmt.Sprint(prec)+"f", v)
	return strings.TrimRight(strings.TrimRight(s, "0"), ".")
}

// TickerFunc is suitable for the Tick.Marker field of an Axis.
// It is an adapter which allows to quickly setup a Ticker using a function with an appropriate signature.
type TickerFunc func(min, max float64) []Tick

var _ Ticker = TickerFunc(nil)

// Ticks implements plot.Ticker.
func (f TickerFunc) Ticks(min, max float64) []Tick {
	return f(min, max)
}
