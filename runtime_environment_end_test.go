// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestRuntimeEnvironmentEndOriginalReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/runtime_environment_end_mathjax_3_2_2.json")
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
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 670 {
		t.Fatal("unbound runtime environment-end references")
	}
	seenNames := make(map[string]bool)
	seenInputs := make(map[struct {
		tex     string
		display bool
	}]bool)
	for _, c := range fixture.Cases {
		key := struct {
			tex     string
			display bool
		}{c.TeX, c.Display}
		if c.Name == "" || c.SVG == "" || seenNames[c.Name] || seenInputs[key] {
			t.Fatal("invalid or duplicate original reference", c.Name)
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

// Original JavaScript runtime exceptions remain raw, not SVG goldens. Require
// a bounded failure at the documented source phase instead of a process panic
// or a silently successful rendering. This is an explicit API deviation.
func TestRuntimeEnvironmentEndBoundedFailures(t *testing.T) {
	data, err := os.ReadFile("testdata/runtime_environment_end_residuals.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name, TeX       string
			Display         bool
			Original        struct{ Error string }
			BoundedAPIError *struct{ Phase, Message string }
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 132 {
		t.Fatal("unbound runtime-end raw inventory")
	}
	apiCount, diagnosticCount := 0, 0
	for _, c := range fixture.Cases {
		if c.Original.Error == "" {
			continue
		}
		if c.BoundedAPIError != nil {
			apiCount++
		} else {
			diagnosticCount++
		}
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if c.BoundedAPIError != nil {
				if c.BoundedAPIError.Phase == "" || err == nil || err.Error() != c.BoundedAPIError.Message || got != "" {
					t.Fatalf("expected documented API failure, got %q, %v", got, err)
				}
			} else if err != nil || !strings.Contains(got, "data-mjx-error=") {
				t.Fatalf("expected inherited bounded diagnostic, got %q, %v", got, err)
			}
		})
	}
	if apiCount != 48 || diagnosticCount != 6 {
		t.Fatalf("unbound runtime boundary counts: %d API, %d diagnostic", apiCount, diagnosticCount)
	}
}
