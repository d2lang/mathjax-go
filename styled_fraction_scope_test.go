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

func TestStyledFractionScopeReferences(t *testing.T) {
	file, err := os.Open("testdata/styled_fraction_scope_mathjax_3_2_2.json.gz")
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
		PrimaryAssets    map[string]string
		Cases            []struct {
			Name, TeX, SVG string
			Display        bool
		}
	}
	if err := json.NewDecoder(reader).Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 170 ||
		fixture.PrimaryAssets["mathjax.js"] != "cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869" ||
		fixture.PrimaryAssets["polyfills.js"] != "7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01" ||
		fixture.PrimaryAssets["setup.js"] != "a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881" {
		t.Fatal("unbound styled-fraction references")
	}
	seen := make(map[string]bool)
	errors := 0
	for _, c := range fixture.Cases {
		if c.Name == "" || seen[c.Name] || c.SVG == "" {
			t.Fatal("invalid styled-fraction inventory", c.Name)
		}
		seen[c.Name] = true
		if strings.Contains(c.SVG, "data-mjx-error=") {
			errors++
		}
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
	if errors != 16 {
		t.Fatal("styled-fraction diagnostic inventory differs", errors)
	}
}
