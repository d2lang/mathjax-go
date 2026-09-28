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

func TestCDEntryReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/cd_entries_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVG                     string
			Display, EscapeAmpersandAttributes bool
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 1060 {
		t.Fatal("unbound CD Entry references")
	}
	seen := make(map[string]bool)
	exact, escaped := 0, 0
	for _, c := range fixture.Cases {
		if c.Name == "" || seen[c.Name] || c.SVG == "" {
			t.Fatal("invalid CD Entry inventory", c.Name)
		}
		seen[c.Name] = true
		if c.EscapeAmpersandAttributes {
			escaped++
		} else {
			exact++
		}
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			want := c.SVG
			if c.EscapeAmpersandAttributes {
				// The original writes raw ampersands into these attributes. Keep its
				// complete reference unchanged, but assert Go's bounded XML safety fix.
				// All rendered text and every other SVG byte must still agree.
				want = strings.NewReplacer(
					`data-mjx-error="Misplaced &"`, `data-mjx-error="Misplaced &amp;"`,
					`fill="&"`, `fill="&amp;"`,
					`stroke="&"`, `stroke="&amp;"`,
					`id="mjx-eqn:&"`, `id="mjx-eqn:&amp;"`,
				).Replace(want)
				if want == c.SVG {
					t.Fatal("escaping case does not exercise raw original attributes")
				}
				decoder := xml.NewDecoder(strings.NewReader(got))
				for {
					if _, err := decoder.Token(); err == io.EOF {
						break
					} else if err != nil {
						t.Fatal("invalid candidate SVG", err)
					}
				}
			}
			if got != want {
				t.Fatalf("complete primary SVG differs\ngot: %s\nwant: %s", got, want)
			}
		})
	}
	if exact != 937 || escaped != 123 {
		t.Fatal("unexpected comparison inventory", exact, escaped)
	}
}
