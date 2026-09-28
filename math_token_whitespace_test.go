// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestMathTokenWhitespaceReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/math_token_whitespace_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVG string
			Display        bool
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 3016 {
		t.Fatal("unbound original math-token references")
	}
	seenNames, seenInputs := make(map[string]bool), make(map[struct {
		tex     string
		display bool
	}]bool)
	for _, c := range fixture.Cases {
		key := struct {
			tex     string
			display bool
		}{c.TeX, c.Display}
		if c.Name == "" || c.SVG == "" || seenNames[c.Name] || seenInputs[key] {
			t.Fatal("invalid or duplicate original math-token reference", c.Name)
		}
		seenNames[c.Name], seenInputs[key] = true, true
		t.Run(c.Name, func(t *testing.T) {
			// Exact equality must not accidentally promote source-returned but
			// XML-invalid literal caller data into the strict SVG partition.
			d := xml.NewDecoder(strings.NewReader(c.SVG))
			depth, roots := 0, 0
			for {
				token, err := d.Token()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatal("original strict SVG is not well formed", err)
				}
				switch v := token.(type) {
				case xml.StartElement:
					if depth == 0 {
						roots++
						if v.Name.Local != "svg" {
							t.Fatal("original root is not SVG")
						}
					}
					depth++
				case xml.EndElement:
					depth--
				case xml.CharData:
					if depth == 0 && strings.TrimSpace(string(v)) != "" {
						t.Fatal("text outside original SVG")
					}
				}
			}
			if depth != 0 || roots != 1 {
				t.Fatal("original strict SVG has an incomplete or multiple root")
			}
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

func TestMathTokenNullRangeReturnsBoundedError(t *testing.T) {
	data, err := os.ReadFile("testdata/math_token_whitespace_residuals.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, OriginalOutcome string
			Display, GuardedNullRange  bool
			Codepoint                 int
			Original                  struct{ SVG, Error string }
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 424 {
		t.Fatal("unbound original math-token residuals")
	}
	checked := 0
	for _, c := range fixture.Cases {
		if !c.GuardedNullRange {
			continue
		}
		checked++
		if c.OriginalOutcome != "runtime" || c.Original.SVG != "" || !strings.Contains(c.Original.Error, "Cannot read properties of null (reading '4')") {
			t.Fatal("guard case is not an original null-range runtime failure", c.Name)
		}
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if err == nil || got != "" {
				t.Fatalf("original runtime failure must return a bounded error, not an SVG: %q, %v", got, err)
			}
			if !strings.Contains(err.Error(), "no Unicode range for character U+") || !strings.Contains(err.Error(), fmt.Sprintf("U+%04X", c.Codepoint)) {
				t.Fatal("missing safe codepoint diagnostic", err)
			}
			for _, r := range err.Error() {
				if r < 0x20 {
					t.Fatal("raw control character in diagnostic")
				}
			}
		})
	}
	if checked != 204 {
		t.Fatal("incomplete source-null-range guard inventory", checked)
	}
}
