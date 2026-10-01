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

type baseCommandRegistrationCase struct {
	Name, TeX string
	Display   bool
	Original  struct{ SVG, Error string }
}

func TestBaseCommandRegistrationOriginalReferences(t *testing.T) {
	file, err := os.Open("testdata/base_command_registration_mathjax_3_2_2.json.gz")
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
		Cases            []baseCommandRegistrationCase
	}
	if err := json.NewDecoder(reader).Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 992 {
		t.Fatal("unbound Base/AMS command registration originals", len(fixture.Cases))
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
			t.Fatal("invalid Base/AMS command registration original", c.Name)
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
	if valid != 824 || renderedErrors != 168 {
		t.Fatalf("original inventory = %d valid, %d rendered errors; want 824, 168", valid, renderedErrors)
	}
}
