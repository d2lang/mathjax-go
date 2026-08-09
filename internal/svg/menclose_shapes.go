// Copyright 2018-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.
//
// Source: ts/output/svg/Notation.ts and
// ts/output/svg/Wrappers/menclose.ts.

package svg

import (
	"math"
	"strings"

	"github.com/d2lang/mathjax-go/internal/jscompat"
)

func (e *encloseLayout) lineData(kind, offset string) [4]float64 {
	bbox := e.w.getBBox()
	h, d, width, halfT := bbox.H, bbox.D, bbox.W, e.thickness/2
	var data [4]float64
	switch kind {
	case "top":
		data = [4]float64{0, h - halfT, width, h - halfT}
	case "right":
		data = [4]float64{width - halfT, -d, width - halfT, h}
	case "bottom":
		data = [4]float64{0, halfT - d, width, halfT - d}
	case "left":
		data = [4]float64{halfT, -d, halfT, h}
	case "vertical":
		data = [4]float64{width / 2, h, width / 2, -d}
	case "horizontal":
		data = [4]float64{0, (h - d) / 2, width, (h - d) / 2}
	case "up":
		data = [4]float64{halfT, halfT - d, width - halfT, h - halfT}
	case "down":
		data = [4]float64{halfT, h - halfT, width - halfT, halfT - d}
	}
	if offset != "" {
		amount := e.getOffset(offset)
		if amount != 0 {
			if offset == "X" {
				data[0], data[2] = data[0]-amount, data[2]-amount
			} else {
				data[1], data[3] = data[1]-amount, data[3]-amount
			}
		}
	}
	return data
}

func (e *encloseLayout) line(data [4]float64) *Element {
	return NewElement("line").
		SetAttr("x1", fixed(data[0])).
		SetAttr("y1", fixed(data[1])).
		SetAttr("x2", fixed(data[2])).
		SetAttr("y2", fixed(data[3])).
		SetAttr("stroke-width", fixed(e.thickness))
}

func (e *encloseLayout) box(radius float64) *Element {
	bbox, thickness := e.w.getBBox(), e.thickness
	box := NewElement("rect").
		SetAttr("x", fixed(thickness/2)).
		SetAttr("y", fixed(thickness/2-bbox.D)).
		SetAttr("width", fixed(bbox.W-thickness)).
		SetAttr("height", fixed(bbox.H+bbox.D-thickness)).
		SetAttr("fill", "none").
		SetAttr("stroke-width", fixed(thickness))
	if radius != 0 {
		box.SetAttr("rx", fixed(radius))
	}
	return box
}

func (e *encloseLayout) ellipse() *Element {
	bbox, thickness := e.w.getBBox(), e.thickness
	return NewElement("ellipse").
		SetAttr("rx", fixed((bbox.W-thickness)/2)).
		SetAttr("ry", fixed((bbox.H+bbox.D-thickness)/2)).
		SetAttr("cx", fixed(bbox.W/2)).
		SetAttr("cy", fixed((bbox.H-bbox.D)/2)).
		SetAttr("fill", "none").
		SetAttr("stroke-width", fixed(thickness))
}

func (e *encloseLayout) phasor() *Element {
	bbox := e.w.getBBox()
	angle, _ := encloseArgMod(1.75*e.padding, bbox.H+bbox.D)
	halfT := e.thickness / 2
	height := bbox.H + bbox.D
	cos := math.Cos(angle)
	return e.path("mitre",
		"M", bbox.W, halfT-bbox.D,
		"L", halfT+cos*halfT, halfT-bbox.D,
		"L", cos*height+halfT, height-bbox.D-halfT,
	)
}

func (e *encloseLayout) longdiv() *Element {
	bbox, halfT := e.w.getBBox(), e.thickness/2
	return e.path("round",
		"M", halfT, halfT-bbox.D,
		"a", e.padding-halfT/2, (bbox.H+bbox.D)/2-4*halfT, 0, "0,1", 0, bbox.H+bbox.D-2*halfT,
		"L", bbox.W-halfT, bbox.H-halfT,
	)
}

func (e *encloseLayout) path(join string, parts ...any) *Element {
	return NewElement("path").
		SetAttr("d", enclosePathData(parts)).
		SetStyle("stroke-width", fixed(e.thickness)).
		SetAttr("stroke-linecap", "round").
		SetAttr("stroke-linejoin", join).
		SetAttr("fill", "none")
}

func (e *encloseLayout) fill(parts ...any) *Element {
	return NewElement("path").SetAttr("d", enclosePathData(parts))
}

func enclosePathData(parts []any) string {
	data := make([]string, len(parts))
	for i, part := range parts {
		switch value := part.(type) {
		case string:
			data[i] = value
		case float64:
			data[i] = fixed(value)
		case int:
			data[i] = fixed(float64(value))
		default:
			data[i] = fixed(encloseNumberValue(value))
		}
	}
	return strings.Join(data, " ")
}

func (e *encloseLayout) arrow(width, angle float64, double bool, offset string, distance float64) *Element {
	bbox := e.w.getBBox()
	dw := (width - bbox.W) / 2
	middle := (bbox.H - bbox.D) / 2
	t, halfT := e.thickness, e.thickness/2
	x, y, dx := t*e.arrowX, t*e.arrowY, t*e.arrowDX
	var arrow *Element
	if double {
		arrow = e.fill(
			"M", bbox.W+dw, middle,
			"l", -(x + dx), y, "l", dx, halfT-y,
			"L", x-dw, middle+halfT,
			"l", dx, y-halfT, "l", -(x + dx), -y,
			"l", x+dx, -y, "l", -dx, y-halfT,
			"L", bbox.W+dw-x, middle-halfT,
			"l", -dx, halfT-y, "Z",
		)
	} else {
		arrow = e.fill(
			"M", bbox.W+dw, middle,
			"l", -(x + dx), y, "l", dx, halfT-y,
			"L", -dw, middle+halfT, "l", 0, -t,
			"L", bbox.W+dw-x, middle-halfT,
			"l", -dx, halfT-y, "Z",
		)
	}
	transforms := make([]string, 0, 2)
	if distance != 0 {
		if offset == "X" {
			transforms = append(transforms, "translate("+fixed(-distance)+" 0)")
		} else {
			transforms = append(transforms, "translate(0 "+fixed(distance)+")")
		}
	}
	if angle != 0 {
		degrees := jscompat.Fixed(-angle*180/math.Pi, 3)
		transforms = append(transforms, "rotate("+degrees+" "+fixed(bbox.W/2)+" "+fixed(middle)+")")
	}
	if len(transforms) != 0 {
		arrow.SetAttr("transform", strings.Join(transforms, " "))
	}
	return arrow
}
