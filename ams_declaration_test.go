// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"os"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestAMSOperatorDeclarationReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/ams_declaration_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases []struct {
			Name, TeX, SVG string
			Display bool
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 1104 {
		t.Fatal("unbound AMS declaration references")
	}
	seenNames, seenInputs := make(map[string]bool), make(map[struct{ tex string; display bool }]bool)
	for _, c := range fixture.Cases {
		key := struct{ tex string; display bool }{c.TeX, c.Display}
		if c.Name == "" || c.SVG == "" || seenNames[c.Name] || seenInputs[key] {
			t.Fatal("invalid or duplicate AMS declaration reference", c.Name)
		}
		seenNames[c.Name], seenInputs[key] = true, true
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.SVG {
				t.Fatalf("complete original SVG differs\ngot: %s\nwant: %s", got, c.SVG)
			}
		})
	}
}

func TestAMSOperatorDeclarationSafeXMLErrors(t *testing.T) {
	data, err := os.ReadFile("testdata/ams_declaration_residuals.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases []struct {
			Name, TeX, Qualification string
			Display bool
			Original struct{ SVG string }
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 20 {
		t.Fatal("unbound AMS declaration residuals")
	}
	seen, checked := make(map[struct{ tex string; display bool }]bool), 0
	for _, c := range fixture.Cases {
		key := struct{ tex string; display bool }{c.TeX, c.Display}
		if seen[key] {
			t.Fatal("duplicate AMS declaration residual", c.Name)
		}
		seen[key] = true
		if c.Qualification != "safe-xml-single-attribute" {
			continue // Raw GetStar and original runtime outcomes are not passing references.
		}
		checked++
		t.Run(c.Name, func(t *testing.T) {
			const originalAttribute = `data-mjx-error="Misplaced &"`
			if strings.Count(c.Original.SVG, originalAttribute) != 1 {
				t.Fatal("unexpected original XML error attribute")
			}
			want := strings.Replace(c.Original.SVG, originalAttribute, `data-mjx-error="Misplaced &amp;"`, 1)
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Fatalf("difference exceeds the single XML attribute escape\ngot: %s\nwant: %s", got, want)
			}
			decoder := xml.NewDecoder(strings.NewReader(got))
			for {
				if _, err := decoder.Token(); err == io.EOF {
					break
				} else if err != nil {
					t.Fatal("invalid complete error SVG", err)
				}
			}
		})
	}
	if checked != 4 {
		t.Fatal("unbound safe-XML control count", checked)
	}
}
