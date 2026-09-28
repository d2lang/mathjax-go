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

func TestMoveEqLeftOriginalReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/move_eq_left_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Counts           struct{ Total, Strict int }
		Cases            []struct {
			Name, TeX, SVG string
			Display        bool
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || fixture.Counts.Total != 1246 || fixture.Counts.Strict != 1064 || len(fixture.Cases) != 1064 {
		t.Fatal("unbound MoveEqLeft references")
	}
	seen := map[string]bool{}
	seenInputs := map[struct {
		tex     string
		display bool
	}]bool{}
	for _, c := range fixture.Cases {
		key := struct {
			tex     string
			display bool
		}{c.TeX, c.Display}
		if c.Name == "" || c.SVG == "" || seen[c.Name] || seenInputs[key] {
			t.Fatal("invalid reference", c.Name)
		}
		seen[c.Name], seenInputs[key] = true, true
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

func TestMoveEqLeftSafeErrorAttribute(t *testing.T) {
	data, err := os.ReadFile("testdata/move_eq_left_residuals.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, Qualification string
			Display                  bool
			Original                 struct{ SVG string }
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" {
		t.Fatal("unbound original error controls")
	}
	count := 0
	seen := map[string]bool{}
	seenInputs := map[struct {
		tex     string
		display bool
	}]bool{}
	for _, c := range fixture.Cases {
		key := struct {
			tex     string
			display bool
		}{c.TeX, c.Display}
		if c.Name == "" || seen[c.Name] || seenInputs[key] {
			t.Fatal("invalid raw inventory", c.Name)
		}
		seen[c.Name], seenInputs[key] = true, true
		if c.Qualification != "safe-XML error attribute" {
			continue
		}
		count++
		t.Run(c.Name, func(t *testing.T) {
			// Preserve the complete original while requiring safe XML serialization.
			const bare = `data-mjx-error="Misplaced &"`
			if strings.Count(c.Original.SVG, bare) != 1 {
				t.Fatal("missing unique original error attribute")
			}
			want := strings.Replace(c.Original.SVG, bare, `data-mjx-error="Misplaced &amp;"`, 1)
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil || got != want {
				t.Fatalf("safe error attribute differs: %v\ngot: %s\nwant: %s", err, got, want)
			}
			decoder := xml.NewDecoder(strings.NewReader(got))
			for {
				_, err := decoder.Token()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatalf("safe error SVG is not well-formed XML: %v", err)
				}
			}
		})
	}
	if count != 168 {
		t.Fatalf("safe XML controls: got %d, want 168", count)
	}
}
