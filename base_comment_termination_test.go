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

type baseCommentTerminationCase struct {
	Name, TeX string
	Display   bool
	Original  struct{ SVG, Error string }
}

func TestBaseCommentTerminationOriginalReferences(t *testing.T) {
	file, err := os.Open("testdata/base_comment_termination_mathjax_3_2_2.json.gz")
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
		Cases            []baseCommentTerminationCase
	}
	if err := json.NewDecoder(reader).Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 302 {
		t.Fatal("unbound Base comment termination originals", len(fixture.Cases))
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
			t.Fatal("invalid Base comment termination original", c.Name)
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
	if valid != 282 || renderedErrors != 20 {
		t.Fatalf("original inventory = %d valid, %d rendered errors; want 282, 20", valid, renderedErrors)
	}
}
