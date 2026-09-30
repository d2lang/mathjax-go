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

type ordinaryArrayOwnerCase struct {
	Name, TeX, Partition, Reason, QualifiedChange string
	SourceIndex                                   int
	Display                                       bool
	Original                                      json.RawMessage
	SourceMemberships                             []json.RawMessage
	PublicationProvenance                         struct{ Category string }
	Contract                                      struct{ Kind, ID, Error, Diagnostic, OriginalFirstLine string }
}

func ordinaryArrayOwnerReferences(t *testing.T, name string) []ordinaryArrayOwnerCase {
	t.Helper()
	f, err := os.Open("testdata/" + name)
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
		MathjaxGitCommit, SourceUnionSHA256 string
		Cases                               []ordinaryArrayOwnerCase
	}
	if err := json.NewDecoder(z).Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || fixture.SourceUnionSHA256 != "816903dc9f2beeece895f5a668c58bf4b8128fe163c71be8bccf7b7e5af2155e" {
		t.Fatal("unbound ordinary-array original references")
	}
	return fixture.Cases
}

func ordinaryArrayOwnerSVG(s string) bool {
	var root struct{ XMLName xml.Name }
	return xml.Unmarshal([]byte(s), &root) == nil && root.XMLName.Local == "svg" && root.XMLName.Space == "http://www.w3.org/2000/svg"
}

func TestOrdinaryArrayOwnerOriginalReferences(t *testing.T) {
	matched := ordinaryArrayOwnerReferences(t, "ordinary_array_owner_mathjax_3_2_2.json.gz")
	residuals := ordinaryArrayOwnerReferences(t, "ordinary_array_owner_residuals.json.gz")
	if len(matched) != 5852 || len(residuals) != 272 {
		t.Fatal("ordinary-array container counts changed")
	}
	type input struct {
		tex     string
		display bool
	}
	names, inputs, indices := map[string]bool{}, map[input]bool{}, map[int]bool{}
	partitions, reasons, provenance, effects, runtime := map[string]int{}, map[string]int{}, map[string]int{}, map[string]int{}, map[string]int{}
	for _, c := range append(matched, residuals...) {
		key := input{c.TeX, c.Display}
		if c.Name == "" || names[c.Name] || inputs[key] || indices[c.SourceIndex] || c.SourceIndex < 0 || c.SourceIndex >= 6124 || len(c.SourceMemberships) == 0 {
			t.Fatal("duplicate or unbound ordinary-array original", c.Name)
		}
		names[c.Name], inputs[key], indices[c.SourceIndex] = true, true, true
		partitions[c.Partition]++
		provenance[c.PublicationProvenance.Category]++
		var original map[string]string
		if err := json.Unmarshal(c.Original, &original); err != nil || len(original) != 1 {
			t.Fatal("unexpected original API object", c.Name, err)
		}
		if c.Partition == "held-svg" {
			// Keep every held source output, but do not make its current
			// mismatch a requirement or substitute a Go-rendered golden.
			if original["error"] != "" || !ordinaryArrayOwnerSVG(original["svg"]) || c.Reason == "" || c.QualifiedChange != "" {
				t.Fatal("invalid held original", c.Name)
			}
			reasons[c.Reason]++
			continue
		}
		if c.Partition == "runtime" {
			runtime[c.Contract.ID]++
		} else {
			effects[c.QualifiedChange]++
		}
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if c.Partition == "runtime" {
				ordinaryArrayOwnerRuntime(t, c, original, got, err)
				return
			}
			want := original["svg"]
			if want == "" || original["error"] != "" || err != nil {
				t.Fatalf("missing original SVG or conversion error: %v", err)
			}
			switch c.Partition {
			case "strict-svg":
				if !ordinaryArrayOwnerSVG(want) {
					t.Fatal("strict original is not SVG XML")
				}
			case "safe-xml":
				const raw = `data-mjx-error="Misplaced &"`
				if ordinaryArrayOwnerSVG(want) || strings.Count(want, raw) != 1 {
					t.Fatal("unbound sole-attribute XML exception")
				}
				// The stored original stays raw. Only this one source-invalid
				// attribute is escaped for full-byte comparison.
				want = strings.Replace(want, raw, `data-mjx-error="Misplaced &amp;"`, 1)
				if !ordinaryArrayOwnerSVG(want) {
					t.Fatal("attribute escape did not produce SVG XML")
				}
			default:
				t.Fatal("unknown original partition", c.Partition)
			}
			if got != want {
				t.Fatalf("complete original differs: got %x want %x", sha256.Sum256([]byte(got)), sha256.Sum256([]byte(want)))
			}
		})
	}
	check := func(label string, got, want map[string]int) {
		t.Helper()
		if len(got) != len(want) {
			t.Fatalf("%s categories changed: %v", label, got)
		}
		for name, n := range want {
			if got[name] != n {
				t.Fatalf("%s %s count: got %d want %d", label, name, got[name], n)
			}
		}
	}
	check("partitions", partitions, map[string]int{"strict-svg": 5608, "safe-xml": 244, "held-svg": 124, "runtime": 148})
	check("held originals", reasons, map[string]int{"AlignedArray optional setup (explicitly separate)": 48, "dcases/drcases existing displaystyle metadata": 4, "Existing NumCases/SubNumCases direct Pop finalization": 8, "Captured displaylines missing explicit center alignment": 64})
	check("publication provenance", provenance, map[string]int{"firstSVGInput": 5898, "priorStrictCompleteSVG": 70, "priorPreservedRawCandidatePromotion": 6, "priorMMLAssertionNewCompleteSVG": 2, "firstRuntimeInput": 148})
	check("qualified changes", effects, map[string]int{"fix": 3174, "control": 2678})
	check("runtime contracts", runtime, map[string]int{"incomplete-script": 14, "unicode-nel": 4, "missing-spread-dimension": 124, "empheq-missing-row": 4, "extra-spreadlines-end": 2})
}

func ordinaryArrayOwnerRuntime(t *testing.T, c ordinaryArrayOwnerCase, original map[string]string, got string, err error) {
	t.Helper()
	contracts := map[string]struct{ first, message string }{
		"incomplete-script":        {"TypeError: Cannot read properties of undefined (reading 'isInferred')", "incomplete msubsup has no script child"},
		"unicode-nel":              {"TypeError: Cannot read properties of null (reading '4')", "no Unicode range for character U+0085"},
		"missing-spread-dimension": {"TypeError: Cannot read properties of undefined (reading 'match')", "spreadlines popped a table without a spread dimension"},
		"empheq-missing-row":       {"TypeError: Cannot read properties of undefined (reading 'appendChild')", "Empheq left block requires a table row"},
		"extra-spreadlines-end":    {"TypeError: Cannot read properties of null (reading 'isInferred')", `Extra \end{spreadlines}`},
	}
	want, ok := contracts[c.Contract.ID]
	if !ok || c.Contract.OriginalFirstLine != want.first || original["svg"] != "" || !strings.HasPrefix(original["error"], want.first+"\n") {
		t.Fatal("unbound complete original runtime")
	}
	if c.Contract.ID == "unicode-nel" && !strings.Contains(c.TeX, "\u0085") {
		t.Fatal("missing source NEL condition")
	}
	if c.Contract.ID != "extra-spreadlines-end" {
		if c.Contract.Kind != "bounded-api" || c.Contract.Error != want.message || got != "" || err == nil || err.Error() != want.message {
			t.Fatalf("bounded API substitute: SVG=%q error=%v", got, err)
		}
		return
	}
	// This inherited control has no original SVG. Assert only successful
	// XML parsing and the reviewed diagnostic, never a Go SVG golden.
	if c.Contract.Kind != "inherited-rendered-diagnostic" || c.Contract.Diagnostic != want.message || err != nil || !ordinaryArrayOwnerSVG(got) {
		t.Fatalf("inherited diagnostic control: error=%v", err)
	}
	decoder, found := xml.NewDecoder(strings.NewReader(got)), 0
	for {
		token, e := decoder.Token()
		if e == io.EOF {
			break
		}
		if e != nil {
			t.Fatal(e)
		}
		if start, ok := token.(xml.StartElement); ok {
			for _, a := range start.Attr {
				if a.Name.Local == "data-mjx-error" {
					if a.Value != want.message {
						t.Fatal("unexpected inherited diagnostic", a.Value)
					}
					found++
				}
			}
		}
	}
	if found != 1 {
		t.Fatal("missing unique inherited diagnostic", found)
	}
}
