// Copyright 2018-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package svg

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/d2lang/mathjax-go/internal/layout"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/tex"
)

type enclosureObservedBox struct {
	W      float64 `json:"w"`
	H      float64 `json:"h"`
	D      float64 `json:"d"`
	Scale  float64 `json:"scale"`
	RScale float64 `json:"rscale"`
	L      float64 `json:"l"`
	R      float64 `json:"r"`
	PWidth string  `json:"pwidth"`
}

type enclosureObservedState struct {
	Event         string                `json:"event"`
	Cached        bool                  `json:"cached"`
	Owned         enclosureObservedBox  `json:"owned"`
	Result        *enclosureObservedBox `json:"result"`
	ResultIsOwned *bool                 `json:"resultIsOwned"`
}

type enclosureFixtureCase struct {
	TeX      string `json:"tex"`
	Display  bool   `json:"display"`
	Notation string `json:"notation"`
	Original struct {
		SVG string `json:"svg"`
	} `json:"original"`
	Lifecycle []struct {
		States []enclosureObservedState `json:"states"`
	} `json:"lifecycle"`
}

func readEnclosureFixtures(t *testing.T, name string, count int) []enclosureFixtureCase {
	t.Helper()
	data, err := os.ReadFile("../../testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []enclosureFixtureCase
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != count {
		t.Fatal("unbound enclosure observations")
	}
	return fixture.Cases
}

// These are authored MathML notation fixtures, not claims that the frozen D2
// configuration registers the optional public \\enclose TeX extension.
func TestEnclosureBBoxNotationsFrozenOriginal(t *testing.T) {
	for i, c := range readEnclosureFixtures(t, "enclosure_notations_mathjax_3_2_2.json", 176) {
		t.Run(fmt.Sprintf("%03d", i), func(t *testing.T) {
			root, err := tex.NewCompiler().Compile(c.TeX, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			for _, n := range root.Find("menclose") {
				n.Attributes.Set("notation", c.Notation)
			}
			options := pipeline.DefaultOptions()
			options.Display = c.Display
			got, err := NewTypesetter().Typeset(root, options)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.Original.SVG {
				t.Fatalf("%s notation=%s display=%v\ngot %s\nwant %s", c.TeX, c.Notation, c.Display, got, c.Original.SVG)
			}
		})
	}
}

// The original observer saves an unmodified SVG first, then explicitly calls
// getBBox(false), invalidateBBox(), and saved/temporary measurements. These
// controlled method observations distinguish a new zero box from a retained
// cache; they are not a claim that normal output repeats this whole sequence.
func TestEnclosureBBoxLifecycleFrozenOriginal(t *testing.T) {
	for i, c := range readEnclosureFixtures(t, "enclosure_lifecycle_mathjax_3_2_2.json", 58) {
		t.Run(fmt.Sprintf("%03d", i), func(t *testing.T) {
			root, err := tex.NewCompiler().Compile(c.TeX, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			options := pipeline.DefaultOptions()
			options.Display = c.Display
			got, err := NewTypesetter().Typeset(root, options)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.Original.SVG {
				t.Fatalf("initial SVG differs for %s display=%v\ngot %s\nwant %s", c.TeX, c.Display, got, c.Original.SVG)
			}
			r := &renderer{options: options, params: layout.TeXParameters, pxPerEm: options.Ex / layout.TeXParameters.XHeight}
			prepareTeXClasses(root)
			wrapped := r.wrap(root, nil, 0, c.Display)
			wrapped.outerBBox()
			wrapped.toSVG(NewElement("g"))
			var targets []*wrapper
			var visit func(*wrapper)
			visit = func(w *wrapper) {
				if w.node.Kind == "menclose" {
					targets = append(targets, w)
				}
				for _, child := range w.children {
					visit(child)
				}
			}
			visit(wrapped)
			if len(targets) != len(c.Lifecycle) {
				t.Fatalf("got %d enclosures, want %d", len(targets), len(c.Lifecycle))
			}
			for j, w := range targets {
				for _, state := range c.Lifecycle[j].States {
					var result *layout.BBox
					switch state.Event {
					case "cached-temporary", "temporary":
						result = w.getBBoxWithSave(false)
					case "invalidate", "second-invalidate":
						w.invalidateBBox()
					case "saved", "second-saved":
						result = w.getBBox()
					case "initial":
					default:
						t.Fatalf("unknown original event %s", state.Event)
					}
					if w.bboxComputed != state.Cached {
						t.Errorf("%s cached=%v, want %v", state.Event, w.bboxComputed, state.Cached)
					}
					checkEnclosureObservedBox(t, state.Event+" owned", w.bbox, &state.Owned)
					checkEnclosureObservedBox(t, state.Event+" result", result, state.Result)
					if state.ResultIsOwned != nil && (result == w.bbox) != *state.ResultIsOwned {
						t.Errorf("%s result ownership mismatch", state.Event)
					}
				}
			}
		})
	}
}

func checkEnclosureObservedBox(t *testing.T, event string, got *layout.BBox, want *enclosureObservedBox) {
	t.Helper()
	if got == nil || want == nil {
		if (got == nil) != (want == nil) {
			t.Errorf("%s nil mismatch", event)
		}
		return
	}
	actual := []float64{got.W, got.H, got.D, got.Scale, got.RScale, got.L, got.R}
	expected := []float64{want.W, want.H, want.D, want.Scale, want.RScale, want.L, want.R}
	for i, a := range actual {
		if math.Abs(a-expected[i]) > 1e-12 {
			t.Errorf("%s field %d got %.17g, want %.17g", event, i, a, expected[i])
		}
	}
	if got.PWidth != want.PWidth {
		t.Errorf("%s pwidth=%q, want %q", event, got.PWidth, want.PWidth)
	}
}
