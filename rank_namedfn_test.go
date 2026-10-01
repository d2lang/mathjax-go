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

type rankNamedFnCase struct {
	Name, TeX string
	Display   bool
	Original  struct{ SVG, Error string }
}

func TestRankNamedFnOriginalReferences(t *testing.T) {
	file, err := os.Open("testdata/rank_namedfn_mathjax_3_2_2.json.gz")
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
		Cases            []rankNamedFnCase
		Deferred         []rankNamedFnCase
	}
	if err := json.NewDecoder(reader).Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 262 {
		t.Fatal("unbound rank NamedFn originals", len(fixture.Cases))
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
			t.Fatal("invalid rank NamedFn original", c.Name)
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
	if valid != 254 || renderedErrors != 8 || len(fixture.Deferred) != 2 {
		t.Fatalf("original inventory = %d valid, %d rendered errors, %d deferred; want 254, 8, 2", valid, renderedErrors, len(fixture.Deferred))
	}
	// Preserve both complete original controls for the existing FnItem/GetForm
	// invisible-ApplyFunction gap. They receive no equality or parity credit.
	seenDisplay := make(map[bool]bool)
	for _, c := range fixture.Deferred {
		if c.TeX != `\rank\times Z` || c.Name == "" || seenDisplay[c.Display] || c.Original.SVG == "" || c.Original.Error != "" || strings.Contains(c.Original.SVG, "data-mjx-error") {
			t.Fatal("invalid deferred original", c.Name)
		}
		seenDisplay[c.Display] = true
	}
}
