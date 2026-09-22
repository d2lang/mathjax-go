// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/tex"
)

type limitsTree struct {
	Kind       string         `json:"kind"`
	Text       *string        `json:"text"`
	Attributes map[string]any `json:"attributes"`
	Properties map[string]any `json:"properties"`
	Children   []*limitsTree  `json:"children"`
}

func limitsProjection(n *mml.Node) *limitsTree {
	out := &limitsTree{Kind: n.Kind, Attributes: map[string]any{}, Properties: map[string]any{}, Children: []*limitsTree{}}
	if n.Kind == "mrow" && n.Flags.Inferred {
		out.Kind = "inferredMrow"
	}
	if n.Kind == "text" {
		text := n.Text
		out.Text = &text
	}
	for _, k := range n.Attributes.ExplicitNames() {
		out.Attributes[k], _ = n.Attributes.GetExplicit(k)
	}
	for _, k := range n.Properties.Keys() {
		out.Properties[k], _ = n.Property(k)
	}
	for _, child := range n.Children {
		out.Children = append(out.Children, limitsProjection(child))
	}
	return out
}
func limitsNodeAt(n *limitsTree, path []int) *limitsTree {
	for _, i := range path {
		if i < 0 || i >= len(n.Children) {
			return nil
		}
		n = n.Children[i]
	}
	return n
}
func limitsField(n *limitsTree, field string) map[string]any {
	if field == "attributes" {
		return n.Attributes
	}
	if field == "properties" {
		return n.Properties
	}
	return nil
}
func TestExplicitLimitsPinnedReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/explicit_limits_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVGSHA256 string
			Display              bool
			Width, Height        int
			Tree                 *limitsTree
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 98 || fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" {
		t.Fatal("unbound primary matrix")
	}
	data, err = os.ReadFile("testdata/explicit_limits_boundaries.json")
	if err != nil {
		t.Fatal(err)
	}
	var boundaries struct {
		InheritedPrimeMetadata map[string][]struct {
			Path         []int
			PrimaryValue bool
			PrimaryText  string
		}
		Baseline      string
		RawBoundaries map[string]struct {
			TeX                                   string
			Display                               bool
			PrimarySHA256, ExpectedSHA256, Reason string
			Tree                                  *limitsTree
			Width, Height                         int
		}
		MetadataBoundaries map[string]struct {
			TeX                           string
			Display                       bool
			PrimarySHA256, BaselineSHA256 string
			BaselineTree                  *limitsTree
			Adjustments                   []struct {
				Path                            []int
				Kind, Field, Key                string
				PrimaryPresent, BaselinePresent bool
				PrimaryValue, BaselineValue     any
			}
		}
	}
	if err = json.Unmarshal(data, &boundaries); err != nil {
		t.Fatal(err)
	}
	if boundaries.Baseline != "0daaf607d985dfdd636e1ff9c353355541425961" || len(boundaries.RawBoundaries) != 0 || len(boundaries.MetadataBoundaries) != 4 || len(boundaries.InheritedPrimeMetadata) != 6 {
		t.Fatal("unbound qualification matrix")
	}
	metadataCounts := map[string]int{"brace-no": 2, "underbrace-limits": 2}
	for stem, count := range metadataCounts {
		for _, mode := range []string{"inline", "display"} {
			if c, ok := boundaries.MetadataBoundaries[stem+"-"+mode]; !ok || len(c.Adjustments) != count {
				t.Fatal("changed exact metadata boundary set")
			}
		}
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			wantTree, wantSHA, wantWidth, wantHeight := c.Tree, c.SVGSHA256, c.Width, c.Height
			if b, ok := boundaries.RawBoundaries[c.Name]; ok {
				if b.TeX != c.TeX || b.Display != c.Display || b.PrimarySHA256 != c.SVGSHA256 || b.ExpectedSHA256 == c.SVGSHA256 || b.Reason == "" {
					t.Fatal("raw boundary source/primary changed")
				}
				wantTree, wantSHA, wantWidth, wantHeight = b.Tree, b.ExpectedSHA256, b.Width, b.Height
			}
			if b, ok := boundaries.MetadataBoundaries[c.Name]; ok {
				if b.TeX != c.TeX || b.Display != c.Display || b.PrimarySHA256 != c.SVGSHA256 || len(b.BaselineSHA256) != 64 {
					t.Fatal("metadata boundary source/primary changed")
				}
				for _, a := range b.Adjustments {
					primary, baseline := limitsNodeAt(wantTree, a.Path), limitsNodeAt(b.BaselineTree, a.Path)
					if primary == nil || baseline == nil || primary.Kind != a.Kind || baseline.Kind != a.Kind {
						t.Fatal("metadata boundary path/kind changed")
					}
					primaryFields, baselineFields := limitsField(primary, a.Field), limitsField(baseline, a.Field)
					if primaryFields == nil || baselineFields == nil {
						t.Fatal("unknown metadata field")
					}
					pv, pp := primaryFields[a.Key]
					bv, bp := baselineFields[a.Key]
					if pp != a.PrimaryPresent || bp != a.BaselinePresent || !reflect.DeepEqual(pv, a.PrimaryValue) || !reflect.DeepEqual(bv, a.BaselineValue) {
						t.Fatal("metadata boundary value no longer bound to primary/baseline")
					}
					// Only these exact accepted-parent fields differ. All other node kinds,
					// ordered children, text, explicit attributes and own properties stay strict.
					if a.BaselinePresent {
						primaryFields[a.Key] = a.BaselineValue
					} else {
						delete(primaryFields, a.Key)
					}
				}
			}
			for _, b := range boundaries.InheritedPrimeMetadata[c.Name] {
				n := limitsNodeAt(wantTree, b.Path)
				if n == nil || n.Kind != "mo" || n.Properties["pseudoscript"] != b.PrimaryValue || len(n.Children) != 1 || n.Children[0].Text == nil || *n.Children[0].Text != b.PrimaryText {
					t.Fatal("changed inherited prime metadata receipt")
				}
				delete(n.Properties, "pseudoscript")
			}
			root, err := tex.NewCompiler().Compile(c.TeX, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(limitsProjection(root))
			if err != nil {
				t.Fatal(err)
			}
			var actual *limitsTree
			if err = json.Unmarshal(raw, &actual); err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(raw), "_texLimitsScriptOrigin") {
				t.Fatal("internal representation marker leaked")
			}
			if !reflect.DeepEqual(actual, wantTree) {
				t.Error("complete explicit/own-property tree differs from exact bound reference")
			}
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			svg, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256([]byte(svg))); got != wantSHA {
				t.Errorf("complete SVG %s; want %s", got, wantSHA)
			}
			if c.Display {
				w, h, err := mathjax.Measure(c.TeX)
				if err != nil {
					t.Fatal(err)
				}
				if w != wantWidth || h != wantHeight {
					t.Errorf("dimensions %dx%d; want %dx%d", w, h, wantWidth, wantHeight)
				}
			}
		})
	}
}
