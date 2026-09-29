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

func TestPhysicsQuantityOriginals(t *testing.T) {
	testPhysicsQuantityFixture(t, "testdata/physics_quantity_mathjax_3_2_2.json.gz", 2760, 16)
}

func TestFinalItemSharedRecipients(t *testing.T) {
	testPhysicsQuantityFixture(t, "testdata/physics_final_item_siblings_mathjax_3_2_2.json.gz", 312, 0)
}

func testPhysicsQuantityFixture(t *testing.T, path string, wantStrict, wantRuntime int) {
	t.Helper()
	f, err := os.Open(path)
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
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, Family string
			Display           bool
			Original          struct{ SVG, Error string }
		}
	}
	if err := json.NewDecoder(z).Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != wantStrict+wantRuntime {
		t.Fatal("unbound Quantity references")
	}
	type pair struct {
		tex     string
		display bool
	}
	names, pairs := map[string]bool{}, map[pair]bool{}
	strict, runtime := 0, 0
	for _, c := range fixture.Cases {
		key := pair{c.TeX, c.Display}
		if c.Name == "" || names[c.Name] || pairs[key] {
			t.Fatal("duplicate Quantity original", c.Name)
		}
		names[c.Name], pairs[key] = true, true
		if c.Original.Error == "" {
			if c.Original.SVG == "" {
				t.Fatal("missing original SVG", c.Name)
			}
			strict++
		} else {
			if c.Original.SVG != "" || !strings.Contains(c.TeX, "\u0085") || !strings.Contains(c.Original.Error, "TypeError: Cannot read properties of null (reading '4')") {
				t.Fatal("unbound original null-range failure", c.Name)
			}
			runtime++
		}
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if c.Original.Error != "" {
				// Source Other dereferences the missing NEL range. The existing
				// bounded conversion-error contract preserves this failure.
				if err == nil || err.Error() != "no Unicode range for character U+0085" || got != "" {
					t.Fatalf("unbounded original runtime outcome: SVG=%q, error=%v", got, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != c.Original.SVG {
				t.Fatalf("complete original SVG differs for %q: got sha256=%x, want=%x", c.TeX, sha256.Sum256([]byte(got)), sha256.Sum256([]byte(c.Original.SVG)))
			}
		})
	}
	if strict != wantStrict || runtime != wantRuntime {
		t.Fatal("Quantity original partition changed", strict, runtime)
	}
}
