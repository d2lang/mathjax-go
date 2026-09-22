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

func TestStackrelPinnedReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/stackrel_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVGSHA256 string
			Display, IsError     bool
			Width, Height        int
			Tree                 *limitsTree
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 72 || fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" {
		t.Fatal("unbound stackrel primary matrix")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			root, err := tex.NewCompiler().Compile(c.TeX, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			// Preserve every node's explicit attributes and own properties, ordered
			// children, kind and text; compare JSON numeric types consistently.
			encoded, err := json.Marshal(limitsProjection(root))
			if err != nil {
				t.Fatal(err)
			}
			var actual *limitsTree
			if err = json.Unmarshal(encoded, &actual); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(actual, c.Tree) {
				t.Error("complete explicit/own-property tree differs from primary")
			}
			opts := mathjax.DefaultOptions()
			opts.Display = c.Display
			svg, err := mathjax.RenderWithOptions(c.TeX, opts)
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256([]byte(svg))); got != c.SVGSHA256 {
				t.Errorf("complete SVG %s, want %s", got, c.SVGSHA256)
			}
			if strings.Contains(svg, `data-mml-node="merror"`) != c.IsError {
				t.Error("primary error disposition changed")
			}
			if c.Display {
				w, h, err := mathjax.Measure(c.TeX)
				if err != nil || w != c.Width || h != c.Height {
					t.Errorf("measure %dx%d, %v; want %dx%d", w, h, err, c.Width, c.Height)
				}
			}
		})
	}
}
