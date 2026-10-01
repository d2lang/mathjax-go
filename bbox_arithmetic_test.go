// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"os"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestBBoxArithmeticOriginalSVGs(t *testing.T) {
	f, err := os.Open("testdata/bbox_arithmetic_mathjax_3_2_2.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var fixture struct {
		MathjaxGitCommit string
		SVGCases         []struct {
			Name, TeX, SVG string
			Display        bool
		}
	}
	if err := json.NewDecoder(z).Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.SVGCases) != 96 {
		t.Fatal("unbound original BBox SVGs")
	}
	type input struct {
		tex     string
		display bool
	}
	names, inputs := make(map[string]bool), make(map[input]bool)
	for _, c := range fixture.SVGCases {
		key := input{c.TeX, c.Display}
		if c.Name == "" || names[c.Name] || inputs[key] || c.SVG == "" {
			t.Fatal("invalid original BBox SVG input", c.Name)
		}
		names[c.Name], inputs[key] = true, true
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.SVG {
				t.Fatalf("complete original differs: got %x want %x", sha256.Sum256([]byte(got)), sha256.Sum256([]byte(c.SVG)))
			}
		})
	}
}
