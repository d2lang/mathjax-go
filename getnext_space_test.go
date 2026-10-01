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

func TestGetNextWhitespaceOriginalReferences(t *testing.T) {
	file, err := os.Open("testdata/getnext_space_mathjax_3_2_2.json.gz")
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
		Name, TeX, Family string
		Display           bool
		Original          struct{ SVG, Error string }
	}
	var fixture struct {
		MathjaxGitCommit string
		SourceBindings   struct {
			Methods  map[string]string
			Commands map[string]struct {
				Binding []string
				Source  string
			}
		}
		Cases, DiagnosticCases, RuntimeCases []originalCase
	}
	if err := json.NewDecoder(reader).Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" ||
		len(fixture.Cases) != 2256 || len(fixture.DiagnosticCases) != 0 || len(fixture.RuntimeCases) != 104 ||
		len(fixture.SourceBindings.Commands) != 36 ||
		!strings.Contains(fixture.SourceBindings.Methods["GetNext"], "nextIsSpace") ||
		!strings.Contains(fixture.SourceBindings.Methods["nextIsSpace"], `\s`) {
		t.Fatal("unbound active GetNext originals")
	}
	families := map[string]int{"Base.Matrix": 0, "Braket.Braket": 0, "Physics.Derivative": 0,
		"Physics.Differential": 0, "Physics.Expression": 0, "Physics.DiagonalMatrix": 0}
	for name, binding := range fixture.SourceBindings.Commands {
		if len(binding.Binding) != 1 || binding.Source == "" {
			t.Fatal("unbound original active handler", name)
		}
		if _, ok := families[binding.Binding[0]]; !ok {
			t.Fatal("unexpected original handler", name, binding.Binding)
		}
	}
	names, inputs := make(map[string]bool), make(map[string]bool)
	for _, c := range fixture.Cases {
		key := c.TeX + "\x00" + map[bool]string{false: "inline", true: "display"}[c.Display]
		if c.Name == "" || names[c.Name] || inputs[key] || c.Original.SVG == "" ||
			c.Original.Error != "" || strings.Contains(c.Original.SVG, "data-mjx-error") {
			t.Fatal("invalid valid-original GetNext case", c.Name)
		}
		names[c.Name], inputs[key] = true, true
		if _, ok := families[c.Family]; !ok {
			t.Fatal("unknown original family", c.Name, c.Family)
		}
		families[c.Family]++
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
	for family, count := range families {
		if count == 0 {
			t.Fatal("missing original active family", family)
		}
	}
	for _, c := range fixture.RuntimeCases {
		if c.Name == "" || names[c.Name] || c.Original.SVG != "" || !strings.Contains(c.TeX, "\u0085") ||
			!strings.HasPrefix(c.Original.Error, "TypeError: Cannot read properties of null (reading '4')\n") ||
			!strings.Contains(c.Original.Error, "at Object.y [as item] (mathjax.js:4:328295)") {
			t.Fatal("unbound original GetNext runtime exception", c.Name)
		}
		names[c.Name] = true
	}
}
