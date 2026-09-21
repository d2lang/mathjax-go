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

func TestOverUnderSetDirectBasePrimary(t *testing.T) {
	data, err := os.ReadFile("testdata/overunderset_direct_base_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name, TeX, SVGSHA256, ExpectedSHA256 string
			Display, Boundary                    bool
			Width, Height                        int
			Tree                                 *nestedFontTree
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 24 {
		t.Fatal("incomplete direct-base matrix")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			actual, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(actual, `data-mml-node="merror"`) {
				t.Fatal("unexpected error output")
			}
			if got := fmt.Sprintf("%x", sha256.Sum256([]byte(actual))); got != c.ExpectedSHA256 {
				t.Fatalf("whole SVG=%s want%s", got, c.ExpectedSHA256)
			}
			if c.Boundary {
				// Overset/Underset are still shadowed by existing macros. Freeze
				// their accepted baseline bytes without calling them primary parity.
				if !strings.HasPrefix(c.Name, "overset-") && !strings.HasPrefix(c.Name, "underset-") {
					t.Fatal("unexpected boundary exception")
				}
				if c.ExpectedSHA256 == c.SVGSHA256 {
					t.Fatal("boundary is no longer a known discrepancy")
				}
				return
			}
			if c.ExpectedSHA256 != c.SVGSHA256 {
				t.Fatal("direct-base case must use raw primary SVG")
			}
			root, err := tex.NewCompiler().Compile(c.TeX, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			got := projectNestedFont(root)
			encoded, _ := json.Marshal(got)
			if err = json.Unmarshal(encoded, &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, c.Tree) {
				t.Fatalf("compiled explicit-attribute tree differs from primary: %s", encoded)
			}
			if c.Display {
				w, h, err := mathjax.Measure(c.TeX)
				if err != nil || w != c.Width || h != c.Height {
					t.Fatalf("Measure=%dx%d,%v want%dx%d", w, h, err, c.Width, c.Height)
				}
			}
		})
	}
}
