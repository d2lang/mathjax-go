// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

type unicodeCharacterOwnerCase struct {
	Name, TeX, Partition, QualifiedChange string
	SourceIndex                           int
	Display                               *bool
	Original                              json.RawMessage
	SourceRecord                          json.RawMessage
	SourceMemberships                     []json.RawMessage
	PublicationProvenance                 struct{ Category string }
	Contract                              struct{ Kind, Codepoint, Error, OriginalFirstLine string }
}

func unicodeCharacterOwnerReferences(t *testing.T, name string) []unicodeCharacterOwnerCase {
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
		MathjaxGitCommit, SourceUnionSHA256, ActualMergedBase string
		Cases                                                 []unicodeCharacterOwnerCase
	}
	if err := json.NewDecoder(z).Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || fixture.SourceUnionSHA256 != "4a5d99ff13db6c0ca8dcd349ce42c6d8638db0598536a270b8b058092c097ac1" || fixture.ActualMergedBase != "6d88373dd7479949265afdeae982b1bd5fcd5613" {
		t.Fatal("unbound Unicode character originals")
	}
	return fixture.Cases
}

func unicodeCharacterOwnerSVG(t *testing.T, name, svg string) {
	t.Helper()
	var root struct{ XMLName xml.Name }
	if err := xml.Unmarshal([]byte(svg), &root); err != nil || root.XMLName.Local != "svg" || root.XMLName.Space != "http://www.w3.org/2000/svg" || strings.Contains(svg, "data-mjx-error=") {
		t.Fatal("expected a valid non-error SVG", name, err)
	}
}

func TestUnicodeCharacterOwnerOriginalReferences(t *testing.T) {
	matched := unicodeCharacterOwnerReferences(t, "unicode_character_owner_mathjax_3_2_2.json.gz")
	held := unicodeCharacterOwnerReferences(t, "unicode_character_owner_residuals.json.gz")
	runtime := unicodeCharacterOwnerReferences(t, "unicode_character_owner_runtime_originals.json.gz")
	if len(matched) != 868 || len(held) != 0 || len(runtime) != 104 {
		t.Fatal("focused original partitions changed")
	}
	inputs, indices := map[string]bool{}, map[int]bool{}
	partitions, svgProvenance, runtimeProvenance, effects, contracts, runtimeEffects := map[string]int{}, map[string]int{}, map[string]int{}, map[string]int{}, map[string]int{}, map[string]int{}
	for _, c := range append(append(matched, held...), runtime...) {
		if c.Name == "" || c.Display == nil || c.SourceIndex < 0 || c.SourceIndex >= 972 || len(c.SourceMemberships) != 1 {
			t.Fatal("unbound focused input", c.Name)
		}
		input := fmt.Sprintf("%t:%s", *c.Display, c.TeX)
		if inputs[input] || indices[c.SourceIndex] {
			t.Fatal("duplicate original input", c.Name)
		}
		inputs[input], indices[c.SourceIndex] = true, true
		var original map[string]string
		if err := json.Unmarshal(c.Original, &original); err != nil || len(original) != 1 {
			t.Fatal("unexpected original API shape", c.Name, err)
		}
		var source struct {
			TeX      string
			Display  *bool
			Original map[string]string
		}
		if err := json.Unmarshal(c.SourceRecord, &source); err != nil || source.Display == nil || *source.Display != *c.Display || source.TeX != c.TeX || !reflect.DeepEqual(source.Original, original) {
			t.Fatal("source record changed", c.Name, err)
		}
		partitions[c.Partition]++
		if c.Partition == "strict-svg" {
			if original["svg"] == "" || c.Contract.Kind != "" {
				t.Fatal("unexpected strict original", c.Name)
			}
			unicodeCharacterOwnerSVG(t, c.Name, original["svg"])
			svgProvenance[c.PublicationProvenance.Category]++
			effects[c.QualifiedChange]++
			t.Run(c.Name, func(t *testing.T) {
				options := mathjax.DefaultOptions()
				options.Display = *c.Display
				got, err := mathjax.RenderWithOptions(c.TeX, options)
				if err != nil {
					t.Fatal(err)
				}
				if got != original["svg"] {
					t.Fatalf("complete original SVG differs: got %x want %x", sha256.Sum256([]byte(got)), sha256.Sum256([]byte(original["svg"])))
				}
			})
			continue
		}
		if c.Partition != "original-runtime" || original["error"] == "" || c.QualifiedChange == "" {
			t.Fatal("unexpected source runtime", c.Name)
		}
		lines := strings.Split(original["error"], "\n")
		if len(lines) != 11 || lines[0] != c.Contract.OriginalFirstLine {
			t.Fatal("incomplete historical original exception", c.Name)
		}
		for _, frame := range lines[1:] {
			if !strings.Contains(frame, "mathjax.js:") {
				t.Fatal("unbound source runtime frame", c.Name)
			}
		}
		runtimeProvenance[c.PublicationProvenance.Category]++
		contracts[c.Contract.Kind]++
		runtimeEffects[c.QualifiedChange]++
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = *c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			switch c.Contract.Kind {
			case "bounded-null-range":
				if c.Contract.OriginalFirstLine != "TypeError: Cannot read properties of null (reading '4')" || c.Contract.Codepoint == "" || c.Contract.Error != "no Unicode range for character "+c.Contract.Codepoint {
					t.Fatal("unbound null-range contract")
				}
				if got != "" || err == nil || err.Error() != c.Contract.Error {
					t.Fatalf("bounded API got %q/%v want empty/%s", got, err, c.Contract.Error)
				}
			case "successful-normnal-fallback":
				// Original MathJax failed before producing SVG. This checks a
				// successful Go fallback, and deliberately has no SVG golden.
				if c.Contract.OriginalFirstLine != "TypeError: Cannot read properties of undefined (reading 'chars')" || c.Contract.Error != "" {
					t.Fatal("unbound original renderer failure")
				}
				if err != nil {
					t.Fatal(err)
				}
				unicodeCharacterOwnerSVG(t, c.Name, got)
			default:
				t.Fatal("unknown source runtime contract", c.Contract.Kind)
			}
		})
	}
	check := func(label string, got, want map[string]int) {
		t.Helper()
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s: got %v want %v", label, got, want)
		}
	}
	check("partition", partitions, map[string]int{"strict-svg": 868, "original-runtime": 104})
	check("SVG provenance", svgProvenance, map[string]int{"firstSVGInput": 840, "priorStrictCompleteSVG": 28})
	check("runtime provenance", runtimeProvenance, map[string]int{"firstOriginalRuntimeInput": 98, "priorOriginalRuntimeWithBoundedGuard": 6})
	check("strict effects", effects, map[string]int{"fix": 522, "control": 346})
	check("runtime contracts", contracts, map[string]int{"bounded-null-range": 98, "successful-normnal-fallback": 6})
	check("runtime effects", runtimeEffects, map[string]int{"new-bounded": 58, "bounded-control": 40, "fallback-control": 6})
	if len(inputs) != 972 || len(indices) != 972 {
		t.Fatal("focused source inputs missing")
	}
}

func TestUnicodeCharacterOwnerPrimeSupplement(t *testing.T) {
	f, err := os.Open("testdata/unicode_character_owner_prime_supplement.json.gz")
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
		MathjaxGitCommit, ActualMergedBase, Partition string
		Cases                                         []struct {
			Name, TeX, Profile, QualifiedChange, PublicationCategory, ExpectedField string
			Display                                                                 *bool
			Original                                                                map[string]string
			SourceMembership                                                        struct {
				Profile, SourcePath, JSONPointer, GitBlob, SourceSHA256 string
				Record                                                  json.RawMessage
			}
			SourceGitBinding struct{ Head, Path, GitBlob, SHA256 string }
		}
	}
	if err := json.NewDecoder(z).Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || fixture.ActualMergedBase != "6d88373dd7479949265afdeae982b1bd5fcd5613" || fixture.Partition != "strict-svg-supplement" || len(fixture.Cases) != 40 {
		t.Fatal("unbound prime supplement")
	}
	inputs, identities, names := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, name := range []string{"unicode_character_owner_mathjax_3_2_2.json.gz", "unicode_character_owner_residuals.json.gz", "unicode_character_owner_runtime_originals.json.gz"} {
		for _, c := range unicodeCharacterOwnerReferences(t, name) {
			if c.Display == nil {
				t.Fatal("missing focused display", c.Name)
			}
			input := fmt.Sprintf("%t:%s", *c.Display, c.TeX)
			if inputs[input] {
				t.Fatal("duplicate focused input", c.Name)
			}
			inputs[input] = true
		}
	}
	if len(inputs) != 972 {
		t.Fatal("focused original input count changed")
	}
	profiles := map[string]int{}
	for _, c := range fixture.Cases {
		if c.Name == "" || c.Display == nil || names[c.Name] {
			t.Fatal("unbound supplemental input", c.Name)
		}
		names[c.Name] = true
		input := fmt.Sprintf("%t:%s", *c.Display, c.TeX)
		member, binding := c.SourceMembership, c.SourceGitBinding
		identity := member.SourcePath + ":" + member.JSONPointer
		if inputs[input] || identities[identity] {
			t.Fatal("duplicate supplemental input or source identity", c.Name)
		}
		inputs[input], identities[identity] = true, true
		if member.Profile != c.Profile || binding.Head != fixture.ActualMergedBase || binding.Path != member.SourcePath || binding.GitBlob != member.GitBlob || binding.SHA256 != member.SourceSHA256 {
			t.Fatal("supplemental Git source binding changed", c.Name)
		}
		var source struct {
			TeX, SVG string
			Display  *bool
			Original map[string]string
		}
		if err := json.Unmarshal(member.Record, &source); err != nil || source.Display == nil || source.TeX != c.TeX || *source.Display != *c.Display {
			t.Fatal("supplemental source input changed", c.Name, err)
		}
		var index int
		if _, err := fmt.Sscanf(member.JSONPointer, "/cases/%d", &index); err != nil || member.JSONPointer != fmt.Sprintf("/cases/%d", index) {
			t.Fatal("invalid source pointer", c.Name)
		}
		var original map[string]string
		switch c.Profile {
		case "raw-prime-residual":
			if member.SourcePath != "testdata/operator_primes_residuals.json" || binding.GitBlob != "8933dac9387a943943ef6eeaa67a441a19ccfdfa" || binding.SHA256 != "4ff8888780f5f22057ce8d48b9a2e47a23b0890b15a9d1a3e68609fd7a2b3003" || !(index >= 121 && index <= 142 || index >= 189 && index <= 204) {
				t.Fatal("unbound original residual source", c.Name)
			}
			if c.QualifiedChange != "fix" || c.PublicationCategory != "firstStrictPromotionFromExistingRaw" || c.ExpectedField != "sourceMembership.record.original.svg" {
				t.Fatal("residual promotion classification changed", c.Name)
			}
			original = source.Original
		case "strict-phantom-control":
			if member.SourcePath != "testdata/operator_primes_mathjax_3_2_2.json" || binding.GitBlob != "e8d161e4e38aabe67fdd921c709bc316a7dd573b" || binding.SHA256 != "2261e9ddd241d11005c4706be360e756d8c8ed867036baee75655ec98ec39b17" || !(index == 3589 || index == 3590) {
				t.Fatal("unbound strict phantom source", c.Name)
			}
			if c.QualifiedChange != "control" || c.PublicationCategory != "priorStrictCompleteSVG" || c.ExpectedField != "sourceMembership.record.svg" {
				t.Fatal("strict phantom classification changed", c.Name)
			}
			original = map[string]string{"svg": source.SVG}
		default:
			t.Fatal("unknown supplemental source profile", c.Profile)
		}
		if len(original) != 1 || original["svg"] == "" || !reflect.DeepEqual(c.Original, original) {
			t.Fatal("original supplemental output changed", c.Name)
		}
		unicodeCharacterOwnerSVG(t, c.Name, original["svg"])
		profiles[c.Profile]++
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = *c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if got != original["svg"] {
				t.Fatalf("complete original SVG differs: got %x want %x", sha256.Sum256([]byte(got)), sha256.Sum256([]byte(original["svg"])))
			}
		})
	}
	if len(inputs) != 1012 || len(identities) != 40 || !reflect.DeepEqual(profiles, map[string]int{"raw-prime-residual": 38, "strict-phantom-control": 2}) {
		t.Fatal("supplemental source coverage changed", profiles)
	}
}
