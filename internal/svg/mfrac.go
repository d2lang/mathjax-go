// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package svg

import (
	"math"

	"github.com/d2lang/mathjax-go/internal/font"
	"github.com/d2lang/mathjax-go/internal/layout"
)

func (w *wrapper) fractionPad() float64 {
	if value, ok := w.node.Property("withDelims"); ok && truthy(value) {
		return 0
	}
	return w.renderer.params.NullDelimiterSpace
}

func (w *wrapper) fractionDisplay() bool { return w.displayStyle && w.scriptLevel == 0 }

func (w *wrapper) fractionThickness() float64 {
	return layout.Length2Em(attribute(w.node, "linethickness", "medium"), .06, w.bbox.Scale, w.renderer.pxPerEm)
}

func (w *wrapper) computeFractionBBox(bbox *layout.BBox) {
	bbox.Empty()
	if len(w.children) < 2 {
		bbox.Clean()
		return
	}
	display := w.fractionDisplay()
	thickness := w.fractionThickness()
	if boolAttributeDefault(w.node, "bevelled", false) {
		w.computeBevelledBBox(bbox, display)
	} else if thickness == 0 {
		u, v, _, numerator, denominator := w.fractionUVQ(display)
		bbox.Combine(numerator, 0, u)
		bbox.Combine(denominator, 0, -v)
		bbox.W += 2 * w.fractionPad()
	} else {
		numerator := w.children[0].outerBBox()
		denominator := w.children[1].outerBBox()
		a := w.renderer.params.Axis
		T, u, v := w.fractionTUV(display, thickness)
		bbox.Combine(numerator, 0, a+T+math.Max(numerator.D*numerator.RScale, u))
		bbox.Combine(denominator, 0, a-T-math.Max(denominator.H*denominator.RScale, v))
		bbox.W += 2*w.fractionPad() + .2
	}
	bbox.Clean()
}

func (w *wrapper) fractionTUV(display bool, thickness float64) (T, u, v float64) {
	params := w.renderer.params
	if display {
		T = 3.5 * thickness
		u = params.Num1 - params.Axis - T
		v = params.Denom1 + params.Axis - T
	} else {
		T = 1.5 * thickness
		u = params.Num2 - params.Axis - T
		v = params.Denom2 + params.Axis - T
	}
	return
}

func (w *wrapper) fractionUVQ(display bool) (u, v, q float64, numerator, denominator *layout.BBox) {
	numerator = w.children[0].outerBBox()
	denominator = w.children[1].outerBBox()
	params := w.renderer.params
	if display {
		u, v = params.Num1, params.Denom1
	} else {
		u, v = params.Num3, params.Denom2
	}
	minimum := 3 * params.Rule
	if display {
		minimum = 7 * params.Rule
	}
	q = (u - numerator.D*numerator.Scale) - (denominator.H*denominator.Scale - v)
	if q < minimum {
		u += (minimum - q) / 2
		v += (minimum - q) / 2
		q = minimum
	}
	return
}

func (w *wrapper) computeBevelledBBox(bbox *layout.BBox, display bool) {
	// The bevel glyph is added by the stretchy-delimiter tranche. Retain the
	// exact numerator/denominator offsets and a conservative slash width until
	// that internal wrapper is available.
	u, v, delta, numerator, denominator := w.bevelData(display)
	bbox.Combine(numerator, 0, u)
	bbox.Combine(layout.NewBBox(.5, .75, .25), bbox.W-delta/2, 0)
	bbox.Combine(denominator, bbox.W-delta/2, v)
}

func (w *wrapper) bevelData(display bool) (u, v, delta float64, numerator, denominator *layout.BBox) {
	numerator = w.children[0].outerBBox()
	denominator = w.children[1].outerBBox()
	delta = .15
	if display {
		delta = .4
	}
	a := w.renderer.params.Axis
	u = numerator.Scale*(numerator.D-numerator.H)/2 + a + delta
	v = denominator.Scale*(denominator.D-denominator.H)/2 + a - delta
	return
}

func (w *wrapper) fractionToSVG(parent *Element) {
	element := w.standardSVG(parent)
	if len(w.children) < 2 {
		return
	}
	if boolAttributeDefault(w.node, "bevelled", false) {
		w.bevelledToSVG(element, w.fractionDisplay())
		return
	}
	thickness := w.fractionThickness()
	if thickness == 0 {
		w.atopToSVG(element, w.fractionDisplay())
	} else {
		w.ruledFractionToSVG(element, w.fractionDisplay(), thickness)
	}
}

func (w *wrapper) ruledFractionToSVG(element *Element, display bool, thickness float64) {
	numerator, denominator := w.children[0], w.children[1]
	nbox, dbox := numerator.outerBBox(), denominator.outerBBox()
	width := math.Max((nbox.L+nbox.W+nbox.R)*nbox.RScale, (dbox.L+dbox.W+dbox.R)*dbox.RScale)
	pad := w.fractionPad()
	nx := alignX(width, nbox, stringAttribute(w.node, "numalign", "center")) + .1 + pad
	dx := alignX(width, dbox, stringAttribute(w.node, "denomalign", "center")) + .1 + pad
	T, u, v := w.fractionTUV(display, thickness)
	a := w.renderer.params.Axis
	numerator.toSVG(element)
	numerator.place(nx, a+T+math.Max(nbox.D*nbox.RScale, u), numerator.element)
	denominator.toSVG(element)
	denominator.place(dx, a-T-math.Max(dbox.H*dbox.RScale, v), denominator.element)
	element.Append(NewElement("rect").
		SetAttr("width", fixed(width+2*.1)).
		SetAttr("height", fixed(thickness)).
		SetAttr("x", fixed(pad)).
		SetAttr("y", fixed(a-thickness/2)))
}

func (w *wrapper) atopToSVG(element *Element, display bool) {
	numerator, denominator := w.children[0], w.children[1]
	nbox, dbox := numerator.outerBBox(), denominator.outerBBox()
	width := math.Max((nbox.L+nbox.W+nbox.R)*nbox.RScale, (dbox.L+dbox.W+dbox.R)*dbox.RScale)
	pad := w.fractionPad()
	nx := alignX(width, nbox, stringAttribute(w.node, "numalign", "center")) + pad
	dx := alignX(width, dbox, stringAttribute(w.node, "denomalign", "center")) + pad
	u, v, _, _, _ := w.fractionUVQ(display)
	numerator.toSVG(element)
	numerator.place(nx, u, numerator.element)
	denominator.toSVG(element)
	denominator.place(dx, -v, denominator.element)
}

func (w *wrapper) bevelledToSVG(element *Element, display bool) {
	// Stretchy slash assembly is completed together with mo. This fallback is
	// deterministic and keeps child placement/source order intact.
	numerator, denominator := w.children[0], w.children[1]
	u, v, delta, nbox, dbox := w.bevelData(display)
	width := (nbox.L + nbox.W + nbox.R) * nbox.RScale
	numerator.toSVG(element)
	numerator.place(nbox.L*nbox.RScale, u, numerator.element)
	slash := NewElement("g").SetAttr("data-mml-node", "mo")
	element.Append(slash)
	temporary := &wrapper{renderer: w.renderer, variant: font.Normal}
	temporary.placeChar('/', 0, 0, slash, font.Normal)
	temporary.place(width-delta/2, 0, slash)
	denominator.toSVG(element)
	denominator.place(width+.5+dbox.L*dbox.RScale-delta, v, denominator.element)
}

func alignX(width float64, bbox *layout.BBox, align string) float64 {
	switch align {
	case "right":
		return width - (bbox.W+bbox.R)*bbox.RScale
	case "left":
		return bbox.L * bbox.RScale
	default:
		return (width - bbox.W*bbox.RScale) / 2
	}
}
