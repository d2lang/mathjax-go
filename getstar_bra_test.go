// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"encoding/xml"
	"io"
	"os"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestGetStarBraOriginalReferences(t *testing.T) {
	type reference struct {
		Name, TeX, Partition, Qualification string
		Display                             bool
		Original                            struct{ SVG, Error string }
	}
	var cases []reference
	for _, path := range []string{"testdata/getstar_bra_mathjax_3_2_2.json.gz", "testdata/getstar_bra_residuals.json.gz"} {
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		z, err := gzip.NewReader(f)
		if err != nil {
			f.Close()
			t.Fatal(err)
		}
		var fixture struct {
			MathjaxGitCommit string
			Cases            []reference
		}
		err = json.NewDecoder(z).Decode(&fixture)
		z.Close()
		f.Close()
		if err != nil {
			t.Fatal(err)
		}
		if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" {
			t.Fatal("unbound GetStar/Bra source")
		}
		cases = append(cases, fixture.Cases...)
	}
	if len(cases) != 5170 {
		t.Fatal("unbound GetStar/Bra inventory", len(cases))
	}
	type input struct {
		tex     string
		display bool
	}
	names, inputs := make(map[string]bool), make(map[input]bool)
	counts := make(map[string]int)
	for _, c := range cases {
		key := input{c.TeX, c.Display}
		if c.Name == "" || names[c.Name] || inputs[key] {
			t.Fatal("invalid or duplicate original", c.Name)
		}
		names[c.Name], inputs[key] = true, true
		counts[c.Partition]++
		if c.Partition == "inherited-valid-raw" {
			if c.Qualification == "" || c.Original.SVG == "" || c.Original.Error != "" || strings.Contains(c.Original.SVG, "data-mjx-error=") {
				t.Fatal("unqualified valid raw original", c.Name)
			}
			getStarBraWellFormed(t, c.Original.SVG)
			continue // Preserved source originals; no passing SVG assertion claimed.
		}
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if c.Partition == "runtime-bounded-nel" {
				const sourceFailure = "TypeError: Cannot read properties of null (reading '4')"
				if c.Original.SVG != "" || !strings.HasPrefix(c.Original.Error, sourceFailure+"\n") || !strings.Contains(c.TeX, "\u0085") {
					t.Fatal("unbound source null-range failure")
				}
				// Source Other dereferences a missing NEL range. This is a
				// bounded API substitute, not original SVG or message parity.
				if err == nil || err.Error() != "no Unicode range for character U+0085" || got != "" {
					t.Fatalf("bounded NEL outcome: SVG=%q error=%v", got, err)
				}
				return
			}
			if c.Original.SVG == "" || c.Original.Error != "" || err != nil {
				t.Fatalf("missing SVG or conversion error: %v", err)
			}
			want := c.Original.SVG
			switch c.Partition {
			case "strict-svg":
				getStarBraWellFormed(t, want)
			case "safe-xml-misplaced-ampersand":
				const from = `data-mjx-error="Misplaced &"`
				const to = `data-mjx-error="Misplaced &amp;"`
				if strings.Count(want, from) != 1 {
					t.Fatal("unbound original error attribute")
				}
				want = strings.Replace(want, from, to, 1)
				getStarBraWellFormed(t, want)
			default:
				t.Fatal("unknown reference partition", c.Partition)
			}
			if got != want {
				t.Fatalf("complete source SVG differs: got %x want %x", sha256.Sum256([]byte(got)), sha256.Sum256([]byte(want)))
			}
		})
	}
	if len(counts) != 4 || counts["strict-svg"] != 4394 || counts["safe-xml-misplaced-ampersand"] != 30 || counts["runtime-bounded-nel"] != 682 || counts["inherited-valid-raw"] != 64 {
		t.Fatal("GetStar/Bra reference partition changed", counts)
	}
}

func getStarBraWellFormed(t *testing.T, svg string) {
	t.Helper()
	d := xml.NewDecoder(strings.NewReader(svg))
	depth, roots := 0, 0
	for {
		token, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal("invalid original XML", err)
		}
		switch v := token.(type) {
		case xml.StartElement:
			if depth == 0 {
				roots++
				if v.Name.Local != "svg" {
					t.Fatal("non-SVG root")
				}
			}
			depth++
		case xml.EndElement:
			depth--
		case xml.CharData:
			if depth == 0 && strings.TrimSpace(string(v)) != "" {
				t.Fatal("text outside SVG root")
			}
		}
	}
	if roots != 1 || depth != 0 {
		t.Fatal("incomplete SVG document")
	}
}
