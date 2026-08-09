// Copyright 2018-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.
//
// Source: ts/output/common/Notation.ts, ts/output/svg/Notation.ts, and
// ts/output/svg/Wrappers/menclose.ts.

package svg

import "math"

var encloseNotationNames = map[string]bool{
	"top": true, "right": true, "bottom": true, "left": true,
	"actuarial": true, "madruwb": true,
	"updiagonalstrike": true, "downdiagonalstrike": true,
	"horizontalstrike": true, "verticalstrike": true,
	"box": true, "roundedbox": true, "circle": true, "phasorangle": true,
	"uparrow": true, "downarrow": true, "leftarrow": true, "rightarrow": true,
	"updownarrow": true, "leftrightarrow": true,
	"updiagonalarrow": true, "northeastarrow": true,
	"southeastarrow": true, "northwestarrow": true, "southwestarrow": true,
	"northeastsouthwestarrow": true, "northwestsoutheastarrow": true,
	"longdiv": true, "radical": true,
}

func encloseNotationSupported(name string) bool { return encloseNotationNames[name] }

func encloseNotationRemove(name string) string {
	switch name {
	case "actuarial":
		return "top right"
	case "madruwb":
		return "bottom right"
	case "box":
		return "left right top bottom"
	case "phasorangle":
		return "bottom"
	case "uparrow":
		return "verticalstrike"
	case "downarrow":
		// The misspelling is present in MathJax 3.2.2 and observable when
		// downarrow and verticalstrike are requested together.
		return "verticakstrike"
	case "leftarrow", "rightarrow":
		return "horizontalstrike"
	case "updownarrow":
		return "verticalstrike uparrow downarrow"
	case "leftrightarrow":
		return "horizontalstrike leftarrow rightarrow"
	case "updiagonalarrow":
		return "updiagonalstrike northeastarrow"
	case "northeastarrow":
		return "updiagonalstrike updiagonalarrow"
	case "southeastarrow", "northwestarrow":
		return "downdiagonalstrike"
	case "southwestarrow":
		return "updiagonalstrike"
	case "northeastsouthwestarrow":
		return "updiagonalstrike northeastarrow updiagonalarrow southwestarrow"
	case "northwestsoutheastarrow":
		return "downdiagonalstrike northwestarrow southeastarrow"
	default:
		return ""
	}
}

func (e *encloseLayout) notationBBox(name string) encloseTRBL {
	p, t := e.padding, e.thickness
	switch name {
	case "top":
		return encloseTRBL{t + p, 0, 0, 0}
	case "right":
		return encloseTRBL{0, t + p, 0, 0}
	case "bottom":
		return encloseTRBL{0, 0, t + p, 0}
	case "left":
		return encloseTRBL{0, 0, 0, t + p}
	case "actuarial":
		return encloseTRBL{t + p, t + p, 0, 0}
	case "madruwb":
		return encloseTRBL{0, t + p, t + p, 0}
	case "updiagonalstrike", "downdiagonalstrike", "box", "roundedbox", "circle":
		return encloseTRBL{t + p, t + p, t + p, t + p}
	case "horizontalstrike":
		return encloseTRBL{0, p, 0, p}
	case "verticalstrike":
		return encloseTRBL{p, 0, p, 0}
	case "phasorangle":
		q := p / 2
		return encloseTRBL{2 * q, q, q + t, 3*q + t}
	case "uparrow":
		return e.arrowBBoxW(encloseTRBL{e.arrowHead(), 0, p, 0})
	case "downarrow":
		return e.arrowBBoxW(encloseTRBL{p, 0, e.arrowHead(), 0})
	case "rightarrow":
		return e.arrowBBoxHD(encloseTRBL{0, e.arrowHead(), 0, p})
	case "leftarrow":
		return e.arrowBBoxHD(encloseTRBL{0, p, 0, e.arrowHead()})
	case "updownarrow":
		return e.arrowBBoxW(encloseTRBL{e.arrowHead(), 0, e.arrowHead(), 0})
	case "leftrightarrow":
		return e.arrowBBoxHD(encloseTRBL{0, e.arrowHead(), 0, e.arrowHead()})
	case "updiagonalarrow", "northeastarrow", "southeastarrow", "northwestarrow", "southwestarrow", "northeastsouthwestarrow", "northwestsoutheastarrow":
		a, _, x, y := e.arrowData()
		b, radius := encloseArgMod(e.arrowX+e.arrowDX, e.arrowY)
		dy := y
		if b > a {
			dy += t * radius * math.Sin(b-a)
		}
		dx := x
		if b > math.Pi/2-a {
			dx += t * radius * math.Sin(b+a-math.Pi/2)
		}
		return encloseTRBL{dy, dx, dy, dx}
	case "longdiv":
		return encloseTRBL{p + t, p, p, 2*p + t/2}
	case "radical":
		return e.sqrtTRBL()
	default:
		return encloseTRBL{}
	}
}

func (e *encloseLayout) notationBorder(name string) encloseTRBL {
	t := e.thickness
	switch name {
	case "top":
		return encloseTRBL{t, 0, 0, 0}
	case "right":
		return encloseTRBL{0, t, 0, 0}
	case "bottom":
		return encloseTRBL{0, 0, t, 0}
	case "left":
		return encloseTRBL{0, 0, 0, t}
	case "actuarial":
		return encloseTRBL{t, t, 0, 0}
	case "madruwb":
		return encloseTRBL{0, t, t, 0}
	case "box":
		return encloseTRBL{t, t, t, t}
	case "phasorangle":
		return encloseTRBL{0, 0, t, 0}
	default:
		return encloseTRBL{}
	}
}

func (e *encloseLayout) arrowHead() float64 {
	return math.Max(e.padding, e.thickness*(e.arrowX+e.arrowDX+1))
}

func (e *encloseLayout) arrowBBoxHD(trbl encloseTRBL) encloseTRBL {
	child := e.childBBox()
	value := math.Max(0, e.thickness*e.arrowY-(child.H+child.D)/2)
	trbl[0], trbl[2] = value, value
	return trbl
}

func (e *encloseLayout) arrowBBoxW(trbl encloseTRBL) encloseTRBL {
	child := e.childBBox()
	value := math.Max(0, e.thickness*e.arrowY-child.W/2)
	trbl[1], trbl[3] = value, value
	return trbl
}

func (e *encloseLayout) renderNotation(name string, element *Element) {
	switch name {
	case "top", "right", "bottom", "left":
		element.Append(e.line(e.lineData(name, "")))
	case "actuarial":
		element.Append(e.line(e.lineData("top", "")), e.line(e.lineData("right", "")))
	case "madruwb":
		element.Append(e.line(e.lineData("bottom", "")), e.line(e.lineData("right", "")))
	case "updiagonalstrike":
		element.Append(e.line(e.lineData("up", "")))
	case "downdiagonalstrike":
		element.Append(e.line(e.lineData("down", "")))
	case "horizontalstrike":
		element.Append(e.line(e.lineData("horizontal", "Y")))
	case "verticalstrike":
		element.Append(e.line(e.lineData("vertical", "X")))
	case "box":
		element.Append(e.box(0))
	case "roundedbox":
		element.Append(e.box(e.thickness + e.padding))
	case "circle":
		element.Append(e.ellipse())
	case "phasorangle":
		element.Append(e.phasor())
	case "uparrow":
		e.renderStraightArrow(element, -math.Pi/2, false, true)
	case "downarrow":
		e.renderStraightArrow(element, math.Pi/2, false, true)
	case "leftarrow":
		e.renderStraightArrow(element, math.Pi, false, false)
	case "rightarrow":
		e.renderStraightArrow(element, 0, false, false)
	case "updownarrow":
		e.renderStraightArrow(element, math.Pi/2, true, true)
	case "leftrightarrow":
		e.renderStraightArrow(element, 0, true, false)
	case "updiagonalarrow", "northeastarrow":
		e.renderDiagonalArrow(element, -1, 0, false)
	case "southeastarrow":
		e.renderDiagonalArrow(element, 1, 0, false)
	case "northwestarrow":
		e.renderDiagonalArrow(element, 1, math.Pi, false)
	case "southwestarrow":
		e.renderDiagonalArrow(element, -1, math.Pi, false)
	case "northeastsouthwestarrow":
		e.renderDiagonalArrow(element, -1, 0, true)
	case "northwestsoutheastarrow":
		e.renderDiagonalArrow(element, 1, 0, true)
	case "longdiv":
		element.Append(e.longdiv())
	}
}

func (e *encloseLayout) renderStraightArrow(element *Element, angle float64, double, vertical bool) {
	bbox := e.w.getBBox()
	width, offset := bbox.W, "Y"
	if vertical {
		width, offset = bbox.H+bbox.D, "X"
	}
	element.Append(e.arrow(width, angle, double, offset, e.getOffset(offset)))
}

func (e *encloseLayout) renderDiagonalArrow(element *Element, coefficient, pi float64, double bool) {
	angle, width := e.arrowAW()
	element.Append(e.arrow(width, coefficient*(angle-pi), double, "", 0))
}
