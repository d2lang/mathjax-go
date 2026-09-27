// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package svg

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/d2lang/mathjax-go/internal/layout"
	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
)

// The retained constructed-MathML capture distinguishes the original default
// from 1/math.Sqrt2 by one ULP, even when SVG formatting hides the difference.
func TestScriptScaleRetainedReference(t *testing.T) {
	var fixture struct {
		MathjaxGitCommit     string
		ScriptSizeMultiplier float64
		SubscriptBBox        layout.BBox
	}
	data, err := os.ReadFile("../../testdata/script_scaling_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || fixture.ScriptSizeMultiplier != 0.7071067811865475 {
		t.Fatal("unbound original script scale")
	}
	n := wrapperTrancheToken("mi", "i", mml.TeXClassOrd)
	n.Attributes.Set("id", "sub")
	n.Attributes.SetInherited("scriptlevel", 1)
	n.Attributes.SetInherited("mathvariant", "italic")
	n.Attributes.Globals().Set("scriptsizemultiplier", fixture.ScriptSizeMultiplier)
	n.Attributes.Globals().Set("scriptminsize", "8px")
	w := wrapperTrancheWrapper(n)
	if got := *w.outerBBox(); got != fixture.SubscriptBBox {
		t.Fatalf("original subscript bbox differs:\n got %+v\nwant %+v", got, fixture.SubscriptBBox)
	}
}

func TestScriptScaleEffectiveAttributes(t *testing.T) {
	for _, test := range []struct {
		name                                     string
		level                                    int
		globalMultiplier, inheritedMultiplier    any
		explicitMultiplier                       any
		globalMinimum, inheritedMinimum, minimum string
		mathsize                                 string
		parentScale, wantScale, wantRelative     float64
		inferred                                 bool
	}{
		{name: "level-zero-ignores-minimum", minimum: "2em", wantScale: 1, wantRelative: 1},
		{name: "raw-default-level-one", level: 1, wantScale: 0.7071067811865475, wantRelative: 0.7071067811865475},
		{name: "raw-default-level-two", level: 2, wantScale: 0.4999999999999999, wantRelative: 0.4999999999999999},
		{name: "level-cap", level: 4, wantScale: 0.4999999999999999, wantRelative: 0.4999999999999999},
		{name: "explicit-multiplier", level: 1, explicitMultiplier: .8, wantScale: .8, wantRelative: .8},
		{name: "explicit-multiplier-level-two", level: 2, explicitMultiplier: .8, wantScale: .6400000000000001, wantRelative: .6400000000000001},
		{name: "inherited-multiplier", level: 1, inheritedMultiplier: .75, wantScale: .75, wantRelative: .75},
		{name: "explicit-precedence", level: 1, globalMultiplier: .8, inheritedMultiplier: .75, explicitMultiplier: .6, wantScale: .6, wantRelative: .6},
		{name: "global-attributes", level: 1, globalMultiplier: .8, globalMinimum: ".9em", wantScale: .9, wantRelative: .9},
		{name: "minimum-em", level: 2, explicitMultiplier: .5, minimum: ".7em", wantScale: .7, wantRelative: .7},
		{name: "inherited-minimum", level: 1, inheritedMultiplier: .75, inheritedMinimum: ".9em", wantScale: .9, wantRelative: .9},
		{name: "minimum-px", level: 2, explicitMultiplier: .5, minimum: "10px", wantScale: .5, wantRelative: .5},
		{name: "minimum-precedence", level: 2, explicitMultiplier: .5, inheritedMinimum: "1.2em", minimum: ".2em", wantScale: .25, wantRelative: .25},
		{name: "mathsize-parent-relative", level: 1, explicitMultiplier: .75, mathsize: "2", parentScale: 2, wantScale: 1.5, wantRelative: .75},
		{name: "mathsize-after-minimum", level: 2, explicitMultiplier: .5, minimum: ".6em", mathsize: "2", wantScale: 1.2, wantRelative: 1.2},
		{name: "inferred-parent-scale", level: 1, explicitMultiplier: .75, mathsize: "2", parentScale: 2, inferred: true, wantScale: 2, wantRelative: 1},
		{name: "numeric-string-multiplier", level: 2, explicitMultiplier: ".75", wantScale: .5625, wantRelative: .5625},
		{name: "negative-level", level: -1, explicitMultiplier: .8, wantScale: 1.25, wantRelative: 1.25},
	} {
		t.Run(test.name, func(t *testing.T) {
			n := wrapperTrancheToken("mi", "x", mml.TeXClassOrd)
			if test.inferred {
				n = mml.NewNode("mrow", nil, nil)
				n.Flags.Inferred = true
			}
			n.Attributes.Set("scriptlevel", test.level)
			if test.globalMultiplier != nil {
				n.Attributes.Globals().Set("scriptsizemultiplier", test.globalMultiplier)
			}
			if test.inheritedMultiplier != nil {
				n.Attributes.SetInherited("scriptsizemultiplier", test.inheritedMultiplier)
			}
			if test.explicitMultiplier != nil {
				n.Attributes.Set("scriptsizemultiplier", test.explicitMultiplier)
			}
			if test.globalMinimum != "" {
				n.Attributes.Globals().Set("scriptminsize", test.globalMinimum)
			}
			if test.inheritedMinimum != "" {
				n.Attributes.SetInherited("scriptminsize", test.inheritedMinimum)
			}
			if test.minimum != "" {
				n.Attributes.Set("scriptminsize", test.minimum)
			}
			if test.mathsize != "" {
				n.Attributes.Set("mathsize", test.mathsize)
			}
			parentScale := test.parentScale
			if parentScale == 0 {
				parentScale = 1
			}
			parent := &wrapper{node: mml.NewNode("mrow", nil, nil), bbox: &layout.BBox{Scale: parentScale}}
			r := &renderer{options: pipeline.DefaultOptions(), params: layout.TeXParameters, pxPerEm: 20}
			w := r.wrap(n, parent, 0, false)
			if w.bbox.Scale != test.wantScale || w.bbox.RScale != test.wantRelative {
				t.Fatalf("scale/relative = %.17g/%.17g, want %.17g/%.17g", w.bbox.Scale, w.bbox.RScale, test.wantScale, test.wantRelative)
			}
		})
	}
}
