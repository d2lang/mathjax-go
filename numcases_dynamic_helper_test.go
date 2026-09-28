// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestNumCasesDynamicHelperOriginalReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/numcases_dynamic_helper_mathjax_3_2_2.json")
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
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 902 {
		t.Fatal("unbound NumCases original references")
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
			t.Fatal("invalid or duplicate NumCases reference", c.Name)
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

// The original fails before it can construct the required first cell. Preserve
// the raw exception, but enforce the public API's bounded conversion-error
// contract rather than allowing an unchecked nil node to panic in Go.
func TestNumCasesEmptyTableConversionBoundary(t *testing.T) {
	data, err := os.ReadFile("testdata/numcases_dynamic_helper_residuals.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name, TeX string
			Display   bool
			Original  struct{ Error string }
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, c := range fixture.Cases {
		if !strings.HasPrefix(c.Original.Error, "TypeError: Cannot read properties of undefined (reading 'appendChild')") {
			continue
		}
		count++
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if err == nil || got != "" {
				t.Fatalf("expected bounded conversion failure, got %q, %v", got, err)
			}
		})
	}
	if count != 46 {
		t.Fatalf("unbound empty-table original failures: %d", count)
	}
}
