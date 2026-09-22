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
	"github.com/d2lang/mathjax-go/internal/tex"
)

func TestFixedInfixFencePinnedReferences(t *testing.T) {
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVGSHA256 string
			Display              bool
			Width, Height        int
			PropertiesTree       *limitsTree
		}
	}
	var boundaries struct {
		Baseline      string
		RawBoundaries map[string]struct {
			TeX, PrimarySHA256, BaselineSHA256 string
			Display                            bool
			BaselineTree                       *limitsTree
		}
		MetadataBoundaries map[string]struct {
			TeX, PrimarySHA256, BaselineSHA256 string
			Display                            bool
			BaselineTree                       *limitsTree
			Adjustments                        []struct {
				Path                            []int
				Kind, Field, Key                string
				PrimaryPresent, BaselinePresent bool
				Primary, Baseline               any
			}
		}
	}
	read := func(path string, target any) {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(data, target); err != nil {
			t.Fatal(err)
		}
	}
	read("testdata/fixed_infix_fences_mathjax_3_2_2.json", &fixture)
	read("testdata/fixed_infix_fences_boundaries.json", &boundaries)
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 48 || boundaries.Baseline != "06fcb3e2df8c5800d3b26cc7e1586c2ff59bd041" || len(boundaries.RawBoundaries) != 6 || len(boundaries.MetadataBoundaries) != 8 {
		t.Fatal("unbound fixed-fence reference corpus")
	}
	rawCount, metadataCount, fixedCount := 0, 0, 0
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			wantSVG, wantTree := c.SVGSHA256, c.PropertiesTree
			if b, ok := boundaries.RawBoundaries[c.Name]; ok {
				rawCount++
				if !(strings.HasPrefix(c.Name, "binom-boundary-") || strings.HasPrefix(c.Name, "dbinom-boundary-") || strings.HasPrefix(c.Name, "tbinom-boundary-")) || b.TeX != c.TeX || b.Display != c.Display || b.PrimarySHA256 != c.SVGSHA256 || b.BaselineSHA256 == c.SVGSHA256 {
					t.Fatal("unexpected D064 boundary")
				}
				wantSVG, wantTree = b.BaselineSHA256, b.BaselineTree
			}
			if b, ok := boundaries.MetadataBoundaries[c.Name]; ok {
				metadataCount++
				if !(strings.HasPrefix(c.Name, "atop-control-") || strings.HasPrefix(c.Name, "genfrac-")) || b.TeX != c.TeX || b.Display != c.Display || b.PrimarySHA256 != c.SVGSHA256 || b.BaselineSHA256 != c.SVGSHA256 {
					t.Fatal("unexpected preexisting metadata boundary")
				}
				for _, a := range b.Adjustments {
					n := limitsNodeAt(wantTree, a.Path)
					if n == nil || n.Kind != a.Kind {
						t.Fatal("metadata path changed")
					}
					fields := limitsField(n, a.Field)
					value, exists := fields[a.Key]
					if fields == nil || exists != a.PrimaryPresent || !reflect.DeepEqual(value, a.Primary) {
						t.Fatal("primary metadata changed")
					}
					if a.BaselinePresent {
						fields[a.Key] = a.Baseline
					} else {
						delete(fields, a.Key)
					}
				}
				if !reflect.DeepEqual(wantTree, b.BaselineTree) {
					t.Fatal("metadata receipts do not reconstruct exact accepted baseline")
				}
			}
			root, err := tex.NewCompiler().Compile(c.TeX, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(limitsProjection(root))
			if err != nil {
				t.Fatal(err)
			}
			var actual *limitsTree
			if err = json.Unmarshal(data, &actual); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(actual, wantTree) {
				t.Error("complete explicit-attribute and own-property tree differs")
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
				t.Fatal("repeat rendering changed")
			}
			if c.Display && wantSVG == c.SVGSHA256 {
				w, h, err := mathjax.Measure(c.TeX)
				if err != nil || w != c.Width || h != c.Height {
					t.Errorf("measure=%dx%d,%v want=%dx%d", w, h, err, c.Width, c.Height)
				}
			}
			if _, raw := boundaries.RawBoundaries[c.Name]; !raw {
				if _, metadata := boundaries.MetadataBoundaries[c.Name]; !metadata && (strings.Contains(c.TeX, "\\choose") || strings.Contains(c.TeX, "\\brace") || strings.Contains(c.TeX, "\\brack")) {
					fixedCount++
				}
			}
		})
	}
	if rawCount != 6 || metadataCount != 8 || fixedCount != 30 {
		t.Fatal("reference classification changed", rawCount, metadataCount, fixedCount)
	}
}
