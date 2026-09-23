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

func TestPodCommandsPinnedReferences(t *testing.T) {
	type output struct {
		SVGSHA256      string
		PropertiesTree *limitsTree
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX     string
			Display       bool
			Width, Height int
			output
		}
	}
	var boundaries struct {
		Baseline  string
		Unchanged map[string]struct {
			TeX     string
			Display bool
			output
		}
	}
	for name, target := range map[string]any{
		"testdata/pod_commands_mathjax_3_2_2.json": &fixture,
		"testdata/pod_commands_boundaries.json":    &boundaries,
	} {
		b, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(b, target); err != nil {
			t.Fatal(err)
		}
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 64 || boundaries.Baseline != "91e9c36681e39a4e8b1482208cd905643bcfcbe5" || len(boundaries.Unchanged) != 8 {
		t.Fatal("unbound pod/pmod references")
	}
	compiler := tex.NewCompiler()
	primary, unchanged := 0, 0
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			want := c.output
			if b, ok := boundaries.Unchanged[c.Name]; ok {
				if !(strings.HasPrefix(c.Name, "mod-") || strings.HasPrefix(c.Name, "bmod-")) || b.TeX != c.TeX || b.Display != c.Display {
					t.Fatal("unexpected separate-command boundary")
				}
				want = b.output
				unchanged++
			} else {
				primary++
			}
			o := mathjax.DefaultOptions()
			o.Display = c.Display
			svg, err := mathjax.RenderWithOptions(c.TeX, o)
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256([]byte(svg))); got != want.SVGSHA256 {
				t.Errorf("whole SVG %s; want %s", got, want.SVGSHA256)
			}
			n, err := compiler.Compile(c.TeX, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(limitsProjection(n))
			if err != nil {
				t.Fatal(err)
			}
			var got *limitsTree
			if err = json.Unmarshal(data, &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want.PropertiesTree) {
				t.Error("complete explicit-attribute and own-property tree differs")
			}
			if c.Display && want.SVGSHA256 == c.SVGSHA256 {
				width, height, err := mathjax.Measure(c.TeX)
				if err != nil || width != c.Width || height != c.Height {
					t.Errorf("measure %dx%d, %v; want %dx%d", width, height, err, c.Width, c.Height)
				}
			}
		})
	}
	if primary != 56 || unchanged != 8 {
		t.Fatal("changed reference classification", primary, unchanged)
	}
}
