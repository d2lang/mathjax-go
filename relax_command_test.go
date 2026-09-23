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
	"testing"
)

func TestRelaxCommandPinnedReferences(t *testing.T) {
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVGSHA256 string
			Display              bool
			Width, Height        int
			Tree                 *stackTree
			PropertiesTree       *limitsTree
		}
	}
	data, err := os.ReadFile("testdata/relax_command_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	var boundaries struct {
		Baseline  string
		Unchanged map[string]struct {
			TeX, SVGSHA256 string
			Display        bool
			Tree           *limitsTree
		}
	}
	data, err = os.ReadFile("testdata/relax_command_boundaries.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &boundaries); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 70 || fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || boundaries.Baseline != "03c34cde20d3c98f326a7cb3961290e0264375ab" || len(boundaries.Unchanged) != 0 {
		t.Fatal("unbound command references")
	}
	expectedBoundaries := map[string]string{}
	for name, c := range boundaries.Unchanged {
		if expectedBoundaries[name] != c.TeX {
			t.Fatal("unexpected non-dispatch boundary", name)
		}
	}
	primaryErrors := 0
	for _, c := range fixture.Cases {
		if c.Tree.Children[0].Children[0].Kind == "merror" {
			primaryErrors++
		}
		t.Run(c.Name, func(t *testing.T) {
			wantSVG, wantTree := c.SVGSHA256, c.PropertiesTree
			boundary, unchanged := boundaries.Unchanged[c.Name]
			if unchanged {
				if boundary.TeX != c.TeX || boundary.Display != c.Display || boundary.SVGSHA256 == c.SVGSHA256 {
					t.Fatal("invalid unchanged boundary")
				}
				wantSVG, wantTree = boundary.SVGSHA256, boundary.Tree
			}
			root, err := tex.NewCompiler().Compile(c.TeX, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			ownJSON, err := json.Marshal(limitsProjection(root))
			if err != nil {
				t.Fatal(err)
			}
			var gotOwn *limitsTree
			if err = json.Unmarshal(ownJSON, &gotOwn); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(gotOwn, wantTree) {
				t.Error("complete explicit/own-property tree differs")
			}
			explicitJSON, err := json.Marshal(projectStackTree(root))
			if err != nil {
				t.Fatal(err)
			}
			var gotExplicit *stackTree
			if err = json.Unmarshal(explicitJSON, &gotExplicit); err != nil {
				t.Fatal(err)
			}
			if !unchanged && !reflect.DeepEqual(gotExplicit, c.Tree) {
				t.Error("complete explicit tree differs from raw primary")
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
			repeat, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil || repeat != svg {
				t.Fatal("repeat render changed")
			}
			if c.Display && !unchanged {
				w, h, err := mathjax.Measure(c.TeX)
				if err != nil || w != c.Width || h != c.Height {
					t.Fatalf("measure=%dx%d,%v want=%dx%d", w, h, err, c.Width, c.Height)
				}
			}
		})
	}
	if primaryErrors != 40 {
		t.Fatal("required primary errors changed", primaryErrors)
	}
}
