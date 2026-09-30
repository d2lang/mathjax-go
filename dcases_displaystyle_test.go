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

type dcasesDisplaystyleCase struct {
	Name, TeX, Partition, Reason, QualifiedChange string
	SourceIndex int
	Display *bool
	Original json.RawMessage
	SourceMemberships []json.RawMessage
	PublicationProvenance struct{ Category string }
}

func dcasesDisplaystyleReferences(t *testing.T, name string) []dcasesDisplaystyleCase {
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
		Cases []dcasesDisplaystyleCase
	}
	if err := json.NewDecoder(z).Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || fixture.SourceUnionSHA256 != "6e2be76864e694bbd8756e27eebe0a9abdf02c4a0cea6c79add36c97dd46d4ae" {
		t.Fatal("unbound dcases displaystyle references")
	}
	return fixture.Cases
}

func TestDcasesDisplaystyleOriginalReferences(t *testing.T) {
	matched := dcasesDisplaystyleReferences(t, "dcases_displaystyle_mathjax_3_2_2.json.gz")
	held := dcasesDisplaystyleReferences(t, "dcases_displaystyle_residuals.json.gz")
	if len(matched) != 500 || len(held) != 8 {
		t.Fatal("dcases displaystyle container counts changed")
	}
	type input struct { tex string; display bool }
	inputs, names, indices := map[input]bool{}, map[string]bool{}, map[int]bool{}
	partitions, provenance, matchedProvenance, effects := map[string]int{}, map[string]int{}, map[string]int{}, map[string]int{}
	valid, errors, memberships := 0, 0, 0
	for _, c := range append(matched, held...) {
		if c.Display == nil || c.Name == "" || len(c.SourceMemberships) == 0 || c.SourceIndex < 0 || c.SourceIndex >= 508 {
			t.Fatal("unbound original input", c.Name)
		}
		key := input{c.TeX, *c.Display}
		if inputs[key] || names[c.Name] || indices[c.SourceIndex] {
			t.Fatal("duplicate original input", c.Name)
		}
		inputs[key], names[c.Name], indices[c.SourceIndex] = true, true, true
		memberships += len(c.SourceMemberships)
		partitions[c.Partition]++
		provenance[c.PublicationProvenance.Category]++
		var original map[string]string
		if err := json.Unmarshal(c.Original, &original); err != nil || len(original) != 1 || original["svg"] == "" {
			t.Fatal("unexpected original API object", c.Name, err)
		}
		var root struct{ XMLName xml.Name }
		if err := xml.Unmarshal([]byte(original["svg"]), &root); err != nil || root.XMLName.Local != "svg" || root.XMLName.Space != "http://www.w3.org/2000/svg" {
			t.Fatal("invalid original SVG", c.Name, err)
		}
		if strings.Contains(original["svg"], "data-mjx-error=") { errors++ } else { valid++ }
		if c.Partition == "held-svg" {
			// Preserve the original without making a current mismatch a requirement.
			if c.Reason != "Inherited displaylines missing explicit center alignment" || c.QualifiedChange != "" || !strings.Contains(c.TeX, `\displaylines`) || strings.Contains(original["svg"], "data-mjx-error=") {
				t.Fatal("unbound held displaylines original", c.Name)
			}
			continue
		}
		if c.Partition != "strict-svg" || c.Reason != "" {
			t.Fatal("unexpected original partition", c.Name)
		}
		matchedProvenance[c.PublicationProvenance.Category]++
		effects[c.QualifiedChange]++
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = *c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil { t.Fatal(err) }
			if got != original["svg"] {
				t.Fatalf("complete original differs: got %x want %x", sha256.Sum256([]byte(got)), sha256.Sum256([]byte(original["svg"])))
			}
		})
	}
	check := func(label string, got, want map[string]int) {
		t.Helper()
		if len(got) != len(want) { t.Fatalf("%s categories changed: %v", label, got) }
		for k, n := range want { if got[k] != n { t.Fatalf("%s %s: got %d want %d", label, k, got[k], n) } }
	}
	check("partitions", partitions, map[string]int{"strict-svg":500, "held-svg":8})
	check("publication", provenance, map[string]int{"firstSVGInput":496, "priorStrictCompleteSVG":4, "priorPreservedRawPromotion":8})
	check("matched publication", matchedProvenance, map[string]int{"firstSVGInput":488, "priorStrictCompleteSVG":4, "priorPreservedRawPromotion":8})
	check("qualified effects", effects, map[string]int{"fix":230, "control":270})
	if valid != 476 || errors != 32 || memberships != 512 { t.Fatal("original kind or source membership count changed", valid, errors, memberships) }
}
