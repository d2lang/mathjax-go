// Copyright 2018-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
//
// Frozen values in this file were produced by D2's pinned MathJax 3.2.2
// JavaScript assets.  The source-shaped fixtures mirror the MathML emitted by
// the cancel and enclose TeX extensions.

package svg

import (
	"crypto/sha256"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/d2lang/mathjax-go/internal/jscompat"
	"github.com/d2lang/mathjax-go/internal/layout"
	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
)

func encloseFixture(notation string) *mml.Node {
	row := mml.NewNode("mrow", nil, nil, wrapperTrancheToken("mi", "x", mml.TeXClassOrd))
	row.Flags.Inferred = true
	node := mml.NewNode("menclose", nil, nil, row)
	node.TeXClass = mml.TeXClassOrd
	node.Attributes.Set("notation", notation)
	return node
}

// typesetEncloseFixture is the exact one-child Typesetter path with the
// menclose dispatch made explicit.  It lets these tests stay useful while the
// central dispatcher is owned and wired by another tranche.
func typesetEncloseFixture(t *testing.T, node *mml.Node) string {
	t.Helper()
	options := pipeline.DefaultOptions()
	r := &renderer{
		options: options,
		params:  layout.TeXParameters,
		pxPerEm: options.Ex / layout.TeXParameters.XHeight,
	}
	root := wrapperTrancheRoot(node)
	root.Attributes.Set("display", "block")
	prepareTeXClasses(root)
	rootWrapper := r.wrap(root, nil, 0, options.Display)
	target := rootWrapper.children[0]
	target.computeEncloseBBox(target.bbox)
	target.bboxComputed = true

	bbox := layout.EmptyBBox()
	bbox.Append(target.outerBBox())
	bbox.Clean()
	rootWrapper.bbox = bbox
	rootWrapper.bboxComputed = true
	px := options.Em / 1000
	width := math.Max(bbox.W, px)
	height := math.Max(bbox.H+bbox.D, px)

	g := NewElement("g").
		SetAttr("stroke", "currentColor").
		SetAttr("fill", "currentColor").
		SetAttr("stroke-width", "0").
		SetAttr("transform", "scale(1,-1)")
	mathElement := rootWrapper.standardSVG(g)
	target.encloseToSVG(mathElement)
	target.place(target.outerBBox().L*target.outerBBox().RScale, 0, target.element)

	return NewElement("svg", g).
		SetStyle("vertical-align", r.ex(-bbox.D)).
		SetAttr("xmlns", namespace).
		SetAttr("width", r.ex(width)).
		SetAttr("height", r.ex(height)).
		SetAttr("role", "img").
		SetAttr("focusable", "false").
		SetAttr("viewBox", "0 "+strings.Join([]string{
			jscompat.Fixed(-bbox.H*1000, 1),
			jscompat.Fixed(width*1000, 1),
			jscompat.Fixed(height*1000, 1),
		}, " ")).String()
}

func TestEncloseCancelFamilyFrozenSVG(t *testing.T) {
	tests := []struct {
		name     string
		notation string
		hash     string
	}{
		{"cancel", "updiagonalstrike", "b8459978ac9b7873a684dbda55a0df87c53d7f31b45100ec9106d15f8cd2a6ac"},
		{"bcancel", "downdiagonalstrike", "f45a793b3150f59567e3d957aacf2dc6830c893eb669c15c051fe247f73240ea"},
		{"xcancel", "updiagonalstrike downdiagonalstrike", "8c746f68330199c48fb379108d546850fc932255d4eb420bff9328b74ed919de"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := typesetEncloseFixture(t, encloseFixture(test.notation))
			hash := fmt.Sprintf("%x", sha256.Sum256([]byte(got)))
			if hash != test.hash {
				t.Fatalf("frozen MathJax 3.2.2 SVG SHA-256 = %s, want %s\n%s", hash, test.hash, got)
			}
		})
	}
}

func TestEncloseFrozenColorBackgroundAndPadding(t *testing.T) {
	node := encloseFixture("updiagonalstrike")
	// Preserve the keyval parser's source order; notation is appended last.
	node.Attributes = mml.NewAttributes(nil, nil)
	node.Attributes.Set("mathcolor", "red")
	node.Attributes.Set("mathbackground", "yellow")
	node.Attributes.Set("data-padding", ".3em")
	node.Attributes.Set("notation", "updiagonalstrike")
	got := typesetEncloseFixture(t, node)
	const want = "819644d28f70fff0d61b4cbc72c03c08b9188f19a92151295122910fcac72c02"
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(got)))
	if hash != want {
		t.Fatalf("frozen colored cancel SVG SHA-256 = %s, want %s\n%s", hash, want, got)
	}
}

func TestEncloseCancelToFrozenMencloseSubtree(t *testing.T) {
	// The base of msup keeps the enclosing display/script state unchanged, so
	// its menclose subtree is identical to this source-shaped direct fixture.
	// The frozen hash is the balanced menclose subtree extracted from
	// \cancelto{0}{x}, not a hash produced by the Go renderer.
	node := encloseFixture("updiagonalstrike updiagonalarrow northeastarrow")
	output := typesetEncloseFixture(t, node)
	subtree := encloseGroup(output)
	if subtree == "" {
		t.Fatalf("cancelto fixture produced no menclose subtree:\n%s", output)
	}
	const want = "a6c062c7efe2b93a76b8a8cb71f6f66cb04ded252f0b55a662f62ac7fedf2cf1"
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(subtree)))
	if hash != want {
		t.Fatalf("frozen cancelto menclose subtree SHA-256 = %s, want %s\n%s", hash, want, subtree)
	}

	parameterized := encloseFixture("updiagonalstrike updiagonalarrow northeastarrow")
	parameterized.Attributes = mml.NewAttributes(nil, nil)
	parameterized.Attributes.Set("data-padding", ".3em")
	parameterized.Attributes.Set("data-thickness", ".08em")
	parameterized.Attributes.Set("data-arrowhead", "5 3 2")
	parameterized.Attributes.Set("notation", "updiagonalstrike updiagonalarrow northeastarrow")
	subtree = encloseGroup(typesetEncloseFixture(t, parameterized))
	const parameterizedWant = "6f962b2dc86206ee28c64516bee3fb25cc2413cd140af99d25cea17175a73844"
	hash = fmt.Sprintf("%x", sha256.Sum256([]byte(subtree)))
	if hash != parameterizedWant {
		t.Fatalf("frozen parameterized cancelto subtree SHA-256 = %s, want %s\n%s", hash, parameterizedWant, subtree)
	}
}

func encloseGroup(svg string) string {
	start := strings.Index(svg, `<g data-mml-node="menclose"`)
	if start < 0 {
		return ""
	}
	depth, position := 0, start
	for position < len(svg) {
		open := strings.Index(svg[position:], "<g")
		close := strings.Index(svg[position:], "</g>")
		if open >= 0 && (close < 0 || open < close) {
			depth++
			position += open + 2
			continue
		}
		if close < 0 {
			return ""
		}
		depth--
		position += close + len("</g>")
		if depth == 0 {
			return svg[start:position]
		}
	}
	return ""
}

func TestEncloseNotationRegistryAndPrimitives(t *testing.T) {
	names := []string{
		"top", "right", "bottom", "left", "actuarial", "madruwb",
		"updiagonalstrike", "downdiagonalstrike", "horizontalstrike", "verticalstrike",
		"box", "roundedbox", "circle", "phasorangle",
		"uparrow", "downarrow", "leftarrow", "rightarrow", "updownarrow", "leftrightarrow",
		"updiagonalarrow", "northeastarrow", "southeastarrow", "northwestarrow", "southwestarrow",
		"northeastsouthwestarrow", "northwestsoutheastarrow", "longdiv", "radical",
	}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			w := wrapperTrancheWrapper(encloseFixture(name))
			w.computeEncloseBBox(w.bbox)
			w.bboxComputed = true
			if w.bbox.W <= 0 || w.bbox.H+w.bbox.D <= 0 {
				t.Fatalf("%s bbox = %#v", name, w.bbox)
			}
			root := NewElement("root")
			w.encloseToSVG(root)
			if output := root.String(); !strings.Contains(output, `data-mml-node="menclose"`) {
				t.Fatalf("%s produced no menclose element: %s", name, output)
			}
		})
	}
}

func TestEncloseDefaultsEmptyAndRedundancy(t *testing.T) {
	omitted := encloseFixture("longdiv")
	omitted.Attributes = mml.NewAttributes(nil, nil)
	omittedState := newEncloseLayout(wrapperTrancheWrapper(omitted))
	if got := omittedState.activeNotations(); len(got) != 1 || got[0] != "longdiv" {
		t.Fatalf("omitted notation = %v, want [longdiv]", got)
	}

	emptyState := newEncloseLayout(wrapperTrancheWrapper(encloseFixture("")))
	if got := emptyState.activeNotations(); len(got) != 0 {
		t.Fatalf("explicit empty notation = %v, want none", got)
	}

	actuarial := newEncloseLayout(wrapperTrancheWrapper(encloseFixture("top right actuarial")))
	if got := actuarial.activeNotations(); len(got) != 1 || got[0] != "actuarial" {
		t.Fatalf("actuarial redundancy = %v, want [actuarial]", got)
	}

	// Preserve MathJax 3.2.2's observable verticakstrike typo.
	down := newEncloseLayout(wrapperTrancheWrapper(encloseFixture("verticalstrike downarrow")))
	if got := down.activeNotations(); len(got) != 2 || got[0] != "verticalstrike" || got[1] != "downarrow" {
		t.Fatalf("downarrow redundancy = %v, want both notations", got)
	}

	cancelTo := newEncloseLayout(wrapperTrancheWrapper(encloseFixture(
		"updiagonalstrike updiagonalarrow northeastarrow",
	)))
	if got := cancelTo.activeNotations(); len(got) != 1 || got[0] != "updiagonalarrow" {
		t.Fatalf("cancelto redundancy = %v, want [updiagonalarrow]", got)
	}
}

func TestEncloseParametersPaddingAndOrder(t *testing.T) {
	node := encloseFixture("roundedbox circle phasorangle longdiv")
	node.Attributes.Set("data-padding", ".3em")
	node.Attributes.Set("data-thickness", ".08em")
	node.Attributes.Set("data-arrowhead", "5 3 2")
	w := wrapperTrancheWrapper(node)
	state := newEncloseLayout(w)
	if state.padding != .3 || state.thickness != .08 || state.arrowX != 5 || state.arrowY != 3 || state.arrowDX != 2 {
		t.Fatalf("parameters = p:%g t:%g arrow:%g %g %g", state.padding, state.thickness, state.arrowX, state.arrowY, state.arrowDX)
	}
	w.computeEncloseBBox(w.bbox)
	w.bboxComputed = true
	root := NewElement("root")
	w.encloseToSVG(root)
	output := root.String()
	indices := []int{
		strings.Index(output, "<rect"),
		strings.Index(output, "<ellipse"),
		strings.Index(output, `stroke-linejoin="mitre"`),
		strings.Index(output, `stroke-linejoin="round"`),
	}
	for i, index := range indices {
		if index < 0 || i > 0 && index <= indices[i-1] {
			t.Fatalf("primitive order differs (%v):\n%s", indices, output)
		}
	}
}

func TestEncloseBoxPaddingExcludesBorder(t *testing.T) {
	state := newEncloseLayout(wrapperTrancheWrapper(encloseFixture("box")))
	padding := state.getPadding()
	for i, got := range padding {
		if math.Abs(got-enclosePadding) > 1e-15 {
			t.Fatalf("box padding[%d] = %g, want %g", i, got, enclosePadding)
		}
	}
}
