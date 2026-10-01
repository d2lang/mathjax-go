// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

type prescriptPostslotsCase struct {
	Name, TeX string
	Display   bool
	Original  struct{ SVG, Error string }
}

func TestPrescriptPostslotsOriginalReferences(t *testing.T) {
	file, err := os.Open("testdata/prescript_postslots_mathjax_3_2_2.json.gz")
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
		Cases            []prescriptPostslotsCase
	}
	if err := json.NewDecoder(reader).Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 1170 {
		t.Fatal("unbound prescript postslot originals", len(fixture.Cases))
	}
	type input struct {
		tex     string
		display bool
	}
	names, inputs := make(map[string]bool), make(map[input]bool)
	for _, c := range fixture.Cases {
		key := input{c.TeX, c.Display}
		if c.Name == "" || names[c.Name] || inputs[key] || c.Original.SVG == "" || c.Original.Error != "" {
			t.Fatal("invalid prescript postslot original", c.Name)
		}
		names[c.Name], inputs[key] = true, true
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
}

// Original runtime failures are retained separately and receive no SVG parity
// credit. Incomplete converted families return a bounded Go error.
func TestPrescriptPostslotsOriginalRuntimeBoundary(t *testing.T) {
	data, err := os.ReadFile("testdata/prescript_postslots_runtime_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []prescriptPostslotsCase
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 2 {
		t.Fatal("unbound source runtime boundaries")
	}
	for _, c := range fixture.Cases {
		if c.TeX != `\prescript{a}{b}{\sum}\limits` || c.Original.Error == "" || c.Original.SVG != "" {
			t.Fatal("invalid original runtime", c.Name)
		}
		options := mathjax.DefaultOptions()
		options.Display = c.Display
		got, err := mathjax.RenderWithOptions(c.TeX, options)
		if err == nil || got != "" {
			t.Fatal("incomplete family must return a bounded error", c.Name)
		}
	}
}
