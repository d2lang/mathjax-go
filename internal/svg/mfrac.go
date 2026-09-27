// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package svg

import (
	"math"

	"github.com/d2lang/mathjax-go/internal/font"
	"github.com/d2lang/mathjax-go/internal/layout"
	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/ordered"
)

// CommonMfrac constructs the slash after its authored children are wrapped.
// Keep its node separate from the input MathML, and size it once, just as the
// original constructor does rather than recreating it during bbox queries.
func (w *wrapper) initializeBevel() {
	if len(w.children) < 2 || !boolAttributeDefault(w.node, "bevelled", false) {
		return
	}
	_, _, delta, numerator, denominator := w.bevelData(w.fractionDisplay())
	height := math.Max(numerator.Scale*(numerator.H+numerator.D),
		denominator.Scale*(denominator.H+denominator.D)) + 2*delta

	defaults := ordered.New[mml.Property]()
	defaults.Set("mathvariant", "normal")
	defaults.Set("mathsize", mml.Inherit)
	mo := mml.NewNode("mo", defaults, w.node.Attributes.Globals(), mml.NewText("/"))
	mo.Flags.Token, mo.Flags.Embellished = true, true
	mo.Attributes.Set("stretchy", true)
	// AbstractMmlNode.inheritAttributesFrom copies only these effective values,
	// not authored IDs, fonts, colors, or other fraction attributes.
	mo.Attributes.SetInherited("displaystyle", attribute(w.node, "displaystyle", w.displayStyle))
	mo.Attributes.SetInherited("scriptlevel", attribute(w.node, "scriptlevel", w.scriptLevel))
	if w.node.Attributes.IsSet("mathsize") {
		mo.Attributes.SetInherited("mathsize", attribute(w.node, "mathsize", "normal"))
	}
	if prime, ok := w.node.Property("texprimestyle"); ok && truthy(prime) {
		mo.SetProperty("texprimestyle", true)
	}
	w.bevel = w.renderer.wrap(mo, w, w.scriptLevel, w.displayStyle)
	mo.Attributes.Set("symmetric", true)
	w.bevel.canStretch(font.DirectionVertical)
	w.bevel.getStretchedVariant([]float64{height}, true)
}

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
	u, v, delta, numerator, denominator := w.bevelData(display)
	bbox.Combine(numerator, 0, u)
	bbox.Combine(w.bevel.outerBBox(), bbox.W-delta/2, 0)
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
	numerator.place(nx, a+T+math.Max(nbox.D*nbox.RScale, u))
	denominator.toSVG(element)
	denominator.place(dx, a-T-math.Max(dbox.H*dbox.RScale, v))
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
	numerator.place(nx, u)
	denominator.toSVG(element)
	denominator.place(dx, -v)
}

func (w *wrapper) bevelledToSVG(element *Element, display bool) {
	numerator, denominator := w.children[0], w.children[1]
	u, v, delta, nbox, dbox := w.bevelData(display)
	width := (nbox.L + nbox.W + nbox.R) * nbox.RScale
	numerator.toSVG(element)
	w.bevel.toSVG(element)
	denominator.toSVG(element)
	numerator.place(nbox.L*nbox.RScale, u)
	w.bevel.place(width-delta/2, 0)
	denominator.place(width+w.bevel.outerBBox().W+dbox.L*dbox.RScale-delta, v)
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
