// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestOperatorNameWhitespaceReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/operatorname_whitespace_mathjax_3_2_2.json")
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
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 962 {
		t.Fatal("unbound original operator-name references")
	}
	type input struct {
		tex     string
		display bool
	}
	seenNames, seenInputs := make(map[string]bool), make(map[input]bool)
	for _, c := range fixture.Cases {
		key := input{c.TeX, c.Display}
		if c.Name == "" || c.SVG == "" || seenNames[c.Name] || seenInputs[key] {
			t.Fatal("invalid or duplicate operator-name reference", c.Name)
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

// The original Other handler throws when a preserved NEL, VT, or FF has no
// Unicode range. Keep that observation and require the existing bounded Go API
// error, rather than treating a rendered replacement as successful parity.
func TestOperatorNameWhitespaceNullRange(t *testing.T) {
	data, err := os.ReadFile("testdata/operatorname_whitespace_residuals.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name, TeX          string
			Display            bool
			NullRangeCodepoint int
			Original           struct{ SVG, Error string }
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 142 {
		t.Fatal("unbound original null-range observations")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.Original.SVG != "" || !strings.HasPrefix(c.Original.Error, "TypeError: Cannot read properties of null (reading '4')") {
				t.Fatal("not an original null-range exception")
			}
			switch c.NullRangeCodepoint {
			case 0x0085, 0x000B, 0x000C:
			default:
				t.Fatal("unbound null-range codepoint")
			}
			if !strings.ContainsRune(c.TeX, rune(c.NullRangeCodepoint)) {
				t.Fatal("missing input codepoint")
			}
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			svg, err := mathjax.RenderWithOptions(c.TeX, options)
			if svg != "" || err == nil || !strings.Contains(err.Error(), "no Unicode range for character "+fmt.Sprintf("U+%04X", c.NullRangeCodepoint)) {
				t.Fatalf("expected bounded null-range failure, got %q, %v", svg, err)
			}
		})
	}
}
