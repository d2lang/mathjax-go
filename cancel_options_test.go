// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

type cancelOptionsCase struct {
	Name, TeX string
	Display   bool
	Original  struct{ SVG, Error string }
}

func TestCancelOptionsOriginalReferences(t *testing.T) {
	file, err := os.Open("testdata/cancel_options_mathjax_3_2_2.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	var fixture struct {
		MathjaxGitCommit string
		Cases            []cancelOptionsCase
		RuntimeCases     []cancelOptionsCase
	}
	if err := json.NewDecoder(reader).Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 1724 || len(fixture.RuntimeCases) != 32 {
		t.Fatal("unbound Cancel option originals", len(fixture.Cases))
	}
	type input struct {
		tex     string
		display bool
	}
	names, inputs := make(map[string]bool), make(map[input]bool)
	valid, renderedErrors := 0, 0
	for _, c := range fixture.Cases {
		key := input{c.TeX, c.Display}
		if c.Name == "" || names[c.Name] || inputs[key] || c.Original.SVG == "" || c.Original.Error != "" {
			t.Fatal("invalid Cancel option original", c.Name)
		}
		names[c.Name], inputs[key] = true, true
		if strings.Contains(c.Original.SVG, "data-mjx-error") {
			renderedErrors++
		} else {
			valid++
		}
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.Original.SVG {
				t.Fatalf("complete original SVG differs\ngot: %s\nwant: %s", got, c.Original.SVG)
			}
		})
	}
	for _, c := range fixture.RuntimeCases {
		key := input{c.TeX, c.Display}
		if c.Name == "" || names[c.Name] || inputs[key] || c.Original.SVG != "" ||
			!strings.HasPrefix(c.Original.Error, "TypeError: t.trim is not a function\n") ||
			!strings.Contains(c.Original.Error, "at e.getParameters (mathjax.js:4:506021)") {
			t.Fatal("unbound original Cancel runtime exception", c.Name)
		}
		names[c.Name], inputs[key] = true, true
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			svg, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil || !strings.HasPrefix(svg, "<svg ") || !strings.Contains(svg, `data-mml-node="menclose"`) || strings.Contains(svg, "data-mjx-error") {
				t.Fatalf("Go typed-attribute boundary: error=%v SVG=%s", err, svg)
			}
		})
	}
	if valid != 1724 || renderedErrors != 0 {
		t.Fatalf("original inventory = %d valid, %d rendered errors; want 1724, 0", valid, renderedErrors)
	}
}
