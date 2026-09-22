// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	mathjax "github.com/d2lang/mathjax-go"
	"github.com/d2lang/mathjax-go/internal/tex"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestPendingPrimePinnedReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/prime_attachment_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name, TeX, SVGSHA256 string
			Display              bool
			Width, Height        int
			Tree                 *stackTree
			PropertiesTree       *limitsTree
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile("testdata/prime_attachment_boundaries.json")
	if err != nil {
		t.Fatal(err)
	}
	var boundaries struct {
		Baseline string
		Unicode  map[string]struct {
			TeX, SVGSHA256, BaselineTreeSHA256 string
			Display                            bool
			Tree                               *limitsTree
		}
		InheritedPrimeMetadata map[string][]struct {
			Path         []int
			PrimaryValue bool
			PrimaryText  string
		}
	}
	if err = json.Unmarshal(data, &boundaries); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 96 || len(boundaries.Unicode) != 4 || len(boundaries.InheritedPrimeMetadata) != 68 || boundaries.Baseline != "7c6a5cf0ec1aaa7a5dcab78c3d68c8720c1c5abc" {
		t.Fatal("unbound prime matrix")
	}
	primaryErrors := 0
	for _, c := range fixture.Cases {
		if c.Tree.Children[0].Children[0].Kind == "merror" {
			primaryErrors++
		}
		t.Run(c.Name, func(t *testing.T) {
			wantSVG, wantTree := c.SVGSHA256, c.PropertiesTree
			unicode, qualified := boundaries.Unicode[c.Name]
			if qualified {
				if unicode.TeX != c.TeX || unicode.Display != c.Display || len(unicode.BaselineTreeSHA256) != 64 || unicode.SVGSHA256 == c.SVGSHA256 {
					t.Fatal("changed D061 boundary")
				}
				wantSVG, wantTree = unicode.SVGSHA256, unicode.Tree
			} else {
				for _, b := range boundaries.InheritedPrimeMetadata[c.Name] {
					n := limitsNodeAt(wantTree, b.Path)
					if n == nil || n.Kind != "mo" || n.Properties["pseudoscript"] != b.PrimaryValue || len(n.Children) != 1 || n.Children[0].Text == nil || *n.Children[0].Text != b.PrimaryText {
						t.Fatal("changed exact inherited metadata path")
					}
					// The accepted compiler omits this MmlMo inheritance property. This
					// bound path/value is separate from the pending-prime parser fix.
					delete(n.Properties, "pseudoscript")
				}
			}
			root, err := tex.NewCompiler().Compile(c.TeX, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(limitsProjection(root))
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(raw), "_texLimitsScriptOrigin") {
				t.Fatal("transient state escaped parse boundary")
			}
			var got *limitsTree
			if err = json.Unmarshal(raw, &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, wantTree) {
				t.Error("complete explicit/own-property tree differs from exact bound reference")
			}
			if !qualified {
				raw, _ = json.Marshal(projectStackTree(root))
				var explicit *stackTree
				_ = json.Unmarshal(raw, &explicit)
				if !reflect.DeepEqual(explicit, c.Tree) {
					t.Error("explicit MathML must be raw primary")
				}
			}
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			svg, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256([]byte(svg))); got != wantSVG {
				t.Errorf("whole SVG=%s want=%s", got, wantSVG)
			}
			repeated, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil || repeated != svg {
				t.Fatal("repeat render changed")
			}
			if c.Display && !qualified {
				w, h, err := mathjax.Measure(c.TeX)
				if err != nil || w != c.Width || h != c.Height {
					t.Fatalf("measurement=%dx%d,%v want=%dx%d", w, h, err, c.Width, c.Height)
				}
			}
		})
	}
	if primaryErrors != 20 {
		t.Fatal("required primary errors changed")
	}
}

func TestPendingPrimeRowLifetimes(t *testing.T) {
	data, err := os.ReadFile("testdata/prime_lifetime_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name, TeX, SVGSHA256 string
			Display              bool
			Tree                 *stackTree
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 66 {
		t.Fatal("missing command/group/font lifetime controls")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			root, err := tex.NewCompiler().Compile(c.TeX, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			data, _ := json.Marshal(projectStackTree(root))
			var got *stackTree
			_ = json.Unmarshal(data, &got)
			if !reflect.DeepEqual(got, c.Tree) {
				t.Fatal("complete explicit lifetime tree differs")
			}
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			svg, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256([]byte(svg))); got != c.SVGSHA256 {
				t.Fatalf("complete lifetime SVG=%s want=%s", got, c.SVGSHA256)
			}
		})
	}
}
