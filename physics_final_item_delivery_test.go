// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"os"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestPhysicsFinalItemDeliveryOriginals(t *testing.T) {
	f, err := os.Open("testdata/physics_final_item_delivery_mathjax_3_2_2.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var fixture struct {
		MathjaxGitCommit   string
		InitialRegressions []struct {
			Name, TeX string
			Display   bool
		}
		Cases []struct {
			Name, TeX string
			Display   bool
			Original  struct{ SVG, Error string }
		}
	}
	if err := json.NewDecoder(z).Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 272 || len(fixture.InitialRegressions) != 10 {
		t.Fatal("unbound original final-item corpus")
	}
	type pair struct {
		tex     string
		display bool
	}
	names := map[string]pair{}
	pairs := map[pair]bool{}
	strict, runtime := 0, 0
	for _, c := range fixture.Cases {
		key := pair{c.TeX, c.Display}
		if _, duplicate := names[c.Name]; duplicate || c.Name == "" || pairs[key] {
			t.Fatal("duplicate final-item control", c.Name)
		}
		names[c.Name], pairs[key] = key, true
		if c.Original.Error != "" {
			if c.Original.SVG != "" {
				t.Fatal("ambiguous original outcome", c.Name)
			}
			// Keep complete source-runtime observations outside strict SVG
			// assertions, but require a bounded failure rather than success.
			runtime++
			t.Run(c.Name, func(t *testing.T) {
				options := mathjax.DefaultOptions()
				options.Display = c.Display
				got, err := mathjax.RenderWithOptions(c.TeX, options)
				if err == nil && !strings.Contains(got, `data-mjx-error=`) {
					t.Fatal("original runtime failure became a successful render")
				}
			})
			continue
		}
		if c.Original.SVG == "" {
			t.Fatal("missing complete original", c.Name)
		}
		strict++
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.Original.SVG {
				t.Fatalf("complete original SVG differs for %q: got sha256=%x, want=%x", c.TeX, sha256.Sum256([]byte(got)), sha256.Sum256([]byte(c.Original.SVG)))
			}
		})
	}
	if strict != 270 || runtime != 2 {
		t.Fatal("original partition changed", strict, runtime)
	}
	for _, c := range fixture.InitialRegressions {
		if key, ok := names[c.Name]; !ok || key != (pair{c.TeX, c.Display}) {
			t.Fatal("lost initial regression", c.Name)
		}
	}
}
