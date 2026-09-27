// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package svg

import (
	"reflect"
	"strings"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

func bevelledTestFraction(display bool, level int, size string) *mml.Node {
	numerator := wrapperTrancheToken("mi", "x", mml.TeXClassOrd)
	denominator := wrapperTrancheToken("mi", "y", mml.TeXClassOrd)
	frac := mml.NewNode("mfrac", nil, nil, numerator, denominator)
	frac.Attributes.Set("bevelled", true)
	frac.Attributes.SetInherited("displaystyle", display)
	frac.Attributes.SetInherited("scriptlevel", level)
	if size != "" {
		for _, n := range []*mml.Node{frac, numerator, denominator} {
			n.Attributes.SetInherited("mathsize", size)
		}
	}
	return frac
}

func TestBevelledFractionSlashSizing(t *testing.T) {
	// Original delimiter sizes are 1, 1.2, 1.8, 2.4, 3em. These expected
	// choices use the x/y metrics and the original .15/.4em clearance.
	for _, c := range []struct {
		name    string
		display bool
		level   int
		size    string
		variant int
	}{
		{"inline", false, 0, "", 0},
		{"display", true, 0, "", 2},
		{"script", true, 1, "", 0},
		{"scaled-display", true, 0, "2", 3},
	} {
		t.Run(c.name, func(t *testing.T) {
			w := wrapperTrancheWrapper(bevelledTestFraction(c.display, c.level, c.size))
			if w.bevel == nil || !w.bevel.sizeSet || w.bevel.size != c.variant {
				t.Fatalf("slash size = %v, want variant %d", w.bevel, c.variant)
			}
			if w.bevel.bbox.Scale != w.bbox.Scale || w.bevel.bbox.RScale != 1 {
				t.Fatalf("slash scale = %g/%g, fraction scale = %g", w.bevel.bbox.Scale, w.bevel.bbox.RScale, w.bbox.Scale)
			}
			slash := w.bevel
			before := *w.getBBox()
			num, den := *w.children[0].outerBBox(), *w.children[1].outerBBox()
			w.bboxComputed = false
			if after := *w.getBBox(); after != before || w.bevel != slash {
				t.Fatalf("repeat bbox changed the fraction or recreated its slash: %v / %v", before, after)
			}
			if *w.children[0].outerBBox() != num || *w.children[1].outerBBox() != den {
				t.Fatal("fraction measurement changed its authored child boxes")
			}
		})
	}
}

func TestBevelledFractionGeneratedOwnershipAndInheritance(t *testing.T) {
	frac := bevelledTestFraction(true, 0, "2")
	frac.Attributes.Set("id", "authored-fraction")
	frac.Attributes.Set("mathvariant", "bold")
	frac.Attributes.Set("mathcolor", "red")
	frac.Attributes.Set("style", "font-family: serif")
	frac.SetProperty("texprimestyle", true)
	children := append([]*mml.Node(nil), frac.Children...)
	w := wrapperTrancheWrapper(frac)
	if w.bevel == nil || w.bevel.parent != w || w.bevel.node.Parent != nil ||
		len(w.children) != 2 || !reflect.DeepEqual(frac.Children, children) {
		t.Fatal("generated slash must be wrapped by the fraction without becoming an authored child")
	}
	for i, child := range children {
		if child.Parent != frac || w.children[i].node != child || w.children[i].parent != w {
			t.Fatal("authored numerator/denominator ownership changed")
		}
	}
	slash := w.bevel.node
	if got := slash.Attributes.ExplicitNames(); !reflect.DeepEqual(got, []string{"stretchy", "symmetric"}) {
		t.Fatalf("generated slash copied authored attributes: %v", got)
	}
	if got := slash.Attributes.InheritedNames(); !reflect.DeepEqual(got, []string{"displaystyle", "scriptlevel", "mathsize"}) {
		t.Fatalf("generated slash inherited unexpected attributes: %v", got)
	}
	if prime, _ := slash.Property("texprimestyle"); prime != true {
		t.Fatal("generated slash lost prime style")
	}
	if slash.Children[0].Parent != slash || nodeText(slash) != "/" {
		t.Fatal("generated slash text ownership differs")
	}
	container := NewElement("g")
	w.toSVG(container)
	if got := strings.Count(container.String(), `id="authored-fraction"`); got != 1 {
		t.Fatalf("authored fraction ID appears %d times", got)
	}
	if len(w.element.Children) != 3 || w.element.Children[0] != w.children[0].element ||
		w.element.Children[1] != w.bevel.element || w.element.Children[2] != w.children[1].element {
		t.Fatal("fraction SVG order is not numerator, generated slash, denominator")
	}
}

func TestBevelledFractionEmptyAndOrdinaryControls(t *testing.T) {
	for _, ruled := range []bool{false, true} {
		frac := bevelledTestFraction(true, 0, "")
		frac.Attributes.Set("bevelled", false)
		if !ruled {
			frac.Attributes.Set("linethickness", "0")
		}
		w := wrapperTrancheWrapper(frac)
		if w.bevel != nil || len(w.children) != 2 || w.getBBox().W <= 0 {
			t.Fatal("ordinary fraction acquired a generated slash or lost its content")
		}
		container := NewElement("g")
		w.toSVG(container)
		if got := strings.Contains(container.String(), "<rect"); got != ruled {
			t.Fatalf("ordinary fraction rule presence = %v, want %v", got, ruled)
		}
	}
	for count := 0; count < 2; count++ {
		frac := bevelledTestFraction(true, 0, "")
		frac.SetChildren(frac.Children[:count])
		w := wrapperTrancheWrapper(frac)
		if w.bevel != nil || w.getBBox().W != 0 {
			t.Fatal("incomplete fraction must retain its empty rendering fallback")
		}
		w.toSVG(NewElement("g"))
	}
}
