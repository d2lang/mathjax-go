// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package svg

import (
	"fmt"
	"math"
	"strings"

	"github.com/d2lang/mathjax-go/internal/jscompat"
	"github.com/d2lang/mathjax-go/internal/layout"
	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
)

const namespace = "http://www.w3.org/2000/svg"

// Typesetter is the stateless MathJax 3.2.2 SVG output stage.
type Typesetter struct{}

// NewTypesetter constructs an SVG typesetter.
func NewTypesetter() *Typesetter { return &Typesetter{} }

var _ pipeline.Typesetter = (*Typesetter)(nil)

type renderer struct {
	options pipeline.Options
	params  layout.Parameters
	pxPerEm float64
}

// Typeset lays out root and returns D2's bare SVG element.
func (t *Typesetter) Typeset(root *mml.Node, options pipeline.Options) (string, error) {
	if root == nil {
		return "", fmt.Errorf("mathjax-go: typeset a nil MathML root")
	}
	if err := options.Validate(); err != nil {
		return "", err
	}
	r := &renderer{
		options: options,
		params:  layout.TeXParameters,
		pxPerEm: options.Ex / layout.TeXParameters.XHeight,
	}
	prepareTeXClasses(root)
	wrapper := r.wrap(root, nil, 0, options.Display)
	bbox := wrapper.outerBBox()
	px := options.Em / 1000
	width := math.Max(bbox.W, px)
	height := math.Max(bbox.H+bbox.D, px)

	g := NewElement("g").
		SetAttr("stroke", "currentColor").
		SetAttr("fill", "currentColor").
		SetAttr("stroke-width", "0").
		SetAttr("transform", "scale(1,-1)")
	wrapper.toSVG(g)

	svg := NewElement("svg", g).
		SetStyle("vertical-align", r.ex(-bbox.D)).
		SetAttr("xmlns", namespace).
		SetAttr("width", r.ex(width)).
		SetAttr("height", r.ex(height)).
		SetAttr("role", "img").
		SetAttr("focusable", "false").
		SetAttr("viewBox", strings.Join([]string{
			jscompat.Fixed(-bbox.H*1000, 1),
			jscompat.Fixed(width*1000, 1),
			jscompat.Fixed(height*1000, 1),
		}, " "))
	// MathJax's viewBox starts at x=0.
	svg.Attributes[len(svg.Attributes)-1].Value = "0 " + svg.Attributes[len(svg.Attributes)-1].Value
	if bbox.PWidth != "" {
		// SVGOutputJax.createRoot uses a responsive width with no viewBox for
		// full-width labeled tables.  The inverse root scale preserves the same
		// 1000-unit glyph coordinate system inside that responsive viewport.
		svg.SetStyle("min-width", r.ex(width))
		svg.SetAttr("width", bbox.PWidth)
		svg.RemoveAttr("viewBox")
		scale := jscompat.Fixed(options.Ex/(r.params.XHeight*1000), 6)
		g.SetAttr("transform", fmt.Sprintf("scale(%s,-%s) translate(0, %s)",
			scale, scale, jscompat.Fixed(-bbox.H*1000, 1)))
	}
	return svg.String(), nil
}

func (r *renderer) ex(value float64) string {
	value /= r.params.XHeight
	if math.Abs(value) < .001 {
		return "0"
	}
	return strings.TrimRight(strings.TrimRight(jscompat.ToFixed(value, 3), "0"), ".") + "ex"
}

func fixed(value float64, digits ...int) string {
	n := 1
	if len(digits) != 0 {
		n = digits[0]
	}
	return jscompat.Fixed(value*1000, n)
}
