// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
	"github.com/d2lang/mathjax-go/internal/tex"
)

func TestMathopLimitsFilterPinnedReferences(t *testing.T) {
	b, err := os.ReadFile("testdata/mathop_limits_filter_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVGSHA256 string
			Display              bool
			Width, Height        int
			Tree                 *limitsTree
		}
	}
	if err = json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Cases) != 28 || f.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" {
		t.Fatal("unbound mathop filter controls")
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			o := mathjax.DefaultOptions()
			o.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, o)
			if err != nil {
				t.Fatal(err)
			}
			if fmt.Sprintf("%x", sha256.Sum256([]byte(got))) != c.SVGSHA256 {
				t.Fatal("complete primary SVG mismatch")
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
			if !reflect.DeepEqual(actual, c.Tree) {
				t.Fatal("complete explicit attribute/own property tree mismatch")
			}
			if c.Display {
				w, h, err := mathjax.Measure(c.TeX)
				if err != nil || w != c.Width || h != c.Height {
					t.Fatal("primary dimensions changed")
				}
			}
		})
	}
}
