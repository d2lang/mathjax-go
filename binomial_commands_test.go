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

func TestBinomialCommandPinnedReferences(t *testing.T) {
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
		Baseline           string
		MetadataBoundaries map[string]struct {
			TeX, PrimarySHA256, BaselineSHA256 string
			Display                            bool
			CandidateTree                      *limitsTree
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
	read("testdata/binomial_commands_mathjax_3_2_2.json", &fixture)
	read("testdata/binomial_commands_boundaries.json", &boundaries)
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 48 || boundaries.Baseline != "46e9024262e80131193abaa119a965789ec1c9ac" || len(boundaries.MetadataBoundaries) != 4 {
		t.Fatal("unbound fixed-fence reference corpus")
	}
	metadataCount := 0
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			wantSVG, wantTree := c.SVGSHA256, c.PropertiesTree
			if b, ok := boundaries.MetadataBoundaries[c.Name]; ok {
				metadataCount++
				if !(strings.HasPrefix(c.Name, "prime-") || strings.HasPrefix(c.Name, "genfrac-control-")) || b.TeX != c.TeX || b.Display != c.Display || b.PrimarySHA256 != c.SVGSHA256 {
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
				if !reflect.DeepEqual(wantTree, b.CandidateTree) {
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
		})
	}
	if metadataCount != 4 {
		t.Fatal("reference classification changed", metadataCount)
	}
}
