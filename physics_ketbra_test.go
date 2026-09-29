// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"encoding/xml"
	"os"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestPhysicsKetBraOriginalReferences(t *testing.T) {
	f, err := os.Open("testdata/physics_ketbra_mathjax_3_2_2.json.gz")
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
			Name, TeX, Partition string
			Display              bool
			Original             struct{ SVG, Error string }
		}
	}
	if err := json.NewDecoder(z).Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 1796 {
		t.Fatal("unbound KetBra originals")
	}
	type input struct {
		tex     string
		display bool
	}
	names, inputs := make(map[string]bool), make(map[input]bool)
	strict, runtime := 0, 0
	for _, c := range fixture.Cases {
		key := input{c.TeX, c.Display}
		if c.Name == "" || names[c.Name] || inputs[key] {
			t.Fatal("duplicate original KetBra input", c.Name)
		}
		names[c.Name], inputs[key] = true, true
		switch c.Partition {
		case "strict-svg":
			strict++
		case "runtime-bounded-nel":
			runtime++
		default:
			t.Fatal("unknown KetBra partition", c.Partition)
		}
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if c.Partition == "runtime-bounded-nel" {
				const firstLine = "TypeError: Cannot read properties of null (reading '4')"
				if c.Original.SVG != "" || !strings.HasPrefix(c.Original.Error, firstLine+"\n") || !strings.Contains(c.TeX, "\u0085") {
					t.Fatal("unbound original null-range runtime")
				}
				// Source Other dereferences a missing NEL range. This remains
				// an explicit bounded API substitute, not original SVG parity.
				if err == nil || err.Error() != "no Unicode range for character U+0085" || got != "" {
					t.Fatalf("bounded runtime: SVG=%q error=%v", got, err)
				}
				return
			}
			if c.Original.SVG == "" || c.Original.Error != "" || err != nil {
				t.Fatalf("missing original SVG or conversion error: %v", err)
			}
			var root struct{ XMLName xml.Name }
			if err := xml.Unmarshal([]byte(c.Original.SVG), &root); err != nil || root.XMLName.Local != "svg" {
				t.Fatalf("invalid original SVG: %v", err)
			}
			if got != c.Original.SVG {
				t.Fatalf("complete original differs: got %x want %x", sha256.Sum256([]byte(got)), sha256.Sum256([]byte(c.Original.SVG)))
			}
		})
	}
	if strict != 1764 || runtime != 32 {
		t.Fatal("KetBra partition changed", strict, runtime)
	}
}
