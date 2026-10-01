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

func TestGetDelimiterWhitespaceOriginalReferences(t *testing.T) {
	file, err := os.Open("testdata/getdelimiter_whitespace_mathjax_3_2_2.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	type originalCase struct {
		Name, TeX string
		Display   bool
		Original  struct{ SVG, Error string }
	}
	var fixture struct {
		MathjaxGitCommit    string
		Methods             map[string]string
		Cases, RuntimeCases []originalCase
	}
	if err := json.NewDecoder(reader).Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 1288 || len(fixture.RuntimeCases) != 2 ||
		!strings.Contains(fixture.Methods["GetDelimiter"], "GetNext") || !strings.Contains(fixture.Methods["GetDelimiter"], ".trim()") ||
		!strings.Contains(fixture.Methods["GetNext"], "nextIsSpace") || !strings.Contains(fixture.Methods["nextIsSpace"], `/\s/`) {
		t.Fatal("unbound original GetDelimiter whitespace observations")
	}
	type input struct {
		tex     string
		display bool
	}
	names, inputs := make(map[string]bool), make(map[input]bool)
	valid, diagnostics := 0, 0
	for _, c := range fixture.Cases {
		key := input{c.TeX, c.Display}
		if c.Name == "" || names[c.Name] || inputs[key] || c.Original.SVG == "" || c.Original.Error != "" {
			t.Fatal("invalid original SVG", c.Name)
		}
		names[c.Name], inputs[key] = true, true
		if strings.Contains(c.Original.SVG, "data-mjx-error") {
			diagnostics++
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
	if valid != 902 || diagnostics != 386 {
		t.Fatalf("original inventory = %d valid, %d rendered diagnostics; want 902, 386", valid, diagnostics)
	}
	// These original runtime failures have no reference SVG. Bind their exact
	// boundary while verifying Go's specific runtime error separately.
	for _, c := range fixture.RuntimeCases {
		key := input{c.TeX, c.Display}
		if names[c.Name] || inputs[key] || c.TeX != "\\left(\u0085x\\middle\u0085|y\\right)" || c.Original.SVG != "" ||
			!strings.HasPrefix(c.Original.Error, "TypeError: Cannot read properties of null (reading '4')\n") ||
			!strings.Contains(c.Original.Error, "at Object.y [as item] (mathjax.js:4:328295)") {
			t.Fatal("unbound original runtime boundary", c.Name)
		}
		names[c.Name], inputs[key] = true, true
		options := mathjax.DefaultOptions()
		options.Display = c.Display
		got, err := mathjax.RenderWithOptions(c.TeX, options)
		if err == nil || err.Error() != "no Unicode range for character U+0085" || got != "" {
			t.Fatalf("unexpected Go runtime control: SVG=%q error=%v", got, err)
		}
	}
}
