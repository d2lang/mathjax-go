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

func TestFlalignReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/flalign_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVG, Comparison string
			Display                    bool
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 3005 {
		t.Fatal("unbound Flalign references")
	}
	seen := make(map[string]bool)
	exact, safe := 0, 0
	for _, c := range fixture.Cases {
		if c.Name == "" || seen[c.Name] || c.SVG == "" {
			t.Fatal("invalid Flalign inventory", c.Name)
		}
		seen[c.Name] = true

		want := c.SVG
		switch c.Comparison {
		case "byte":
			exact++
		case "safeXML":
			safe++
			found := false
			for _, name := range []string{"xalignat", "xalignat*", "xxalignat"} {
				from := `data-mjx-error="Extra & in row of ` + name + `"`
				if strings.Count(want, from) == 1 {
					want = strings.Replace(want, from, strings.Replace(from, " & ", " &amp; ", 1), 1)
					found = true
					break
				}
			}
			if !found {
				t.Fatal("unbound original XML diagnostic", c.Name)
			}
		default:
			t.Fatal("unknown comparison", c.Comparison)
		}
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if c.Comparison == "safeXML" {
				decoder := xml.NewDecoder(strings.NewReader(got))
				for {
					_, err := decoder.Token()
					if err == io.EOF {
						break
					}
					if err != nil {
						t.Fatal(err)
					}
				}
			}
			if got != want {
				t.Fatalf("complete primary SVG differs\ngot: %s\nwant: %s", got, want)
			}
		})
	}
	if exact != 2931 || safe != 74 {
		t.Fatal("unexpected primary comparison inventory", exact, safe)
	}
}
