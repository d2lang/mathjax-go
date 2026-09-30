// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

type displaylinesArrayOwnerReference struct {
	Name, TeX, Partition, QualifiedChange, PublicationCategory string
	Index                                                      int
	Display                                                    *bool
	Original                                                   map[string]string
	SourceMemberships                                          []struct {
		Document, Profile string
		Index             int
	}
	Contract struct {
		Kind, Codepoint, Error, OriginalFirstLine, Raw, Escaped string
		SourcePublic110Index                                    int
	}
}

func displaylinesArrayOwnerSource(t *testing.T, name, hash string) []json.RawMessage {
	t.Helper()
	file, err := os.Open("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	compressed, err := io.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(compressed)) != hash {
		t.Fatal("changed complete original archive", name)
	}
	reader, err := gzip.NewReader(strings.NewReader(string(compressed)))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	var source struct{ Cases []json.RawMessage }
	if err := json.NewDecoder(reader).Decode(&source); err != nil {
		t.Fatal(err)
	}
	return source.Cases
}

func displaylinesArrayOwnerSVG(t *testing.T, s string) {
	t.Helper()
	var root struct{ XMLName xml.Name }
	if err := xml.Unmarshal([]byte(s), &root); err != nil || root.XMLName.Local != "svg" || root.XMLName.Space != "http://www.w3.org/2000/svg" {
		t.Fatal("invalid SVG XML", err)
	}
}

func TestDisplaylinesArrayOwnerOriginalReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/displaylines_array_owner_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit, ActualMergedBase string
		SourceDocuments                    map[string]struct{ SHA256 string }
		Cases                              []displaylinesArrayOwnerReference
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || fixture.ActualMergedBase != "b745ce2b28400d18cd98fb7d14cd4f4280e936de" || len(fixture.Cases) != 116 {
		t.Fatal("unbound displaylines originals")
	}
	sourceHashes := map[string]string{
		"displaylines_array_owner_original_public110.json.gz":    "900ba6bf121f85288572e79fbabeeecfb90548c468b36b584d36b46f0b2c000c",
		"displaylines_array_owner_direct24_observations.json.gz": "b2e031e9b13436d18b65b9a8effc2a2e7d2ff1eee60ac0ef682257e9dd42a923",
		"displaylines_native_owner_observations.json.gz":         "eaf2e8d9ae60dfc720b9dc0ff45b1e8069007022e940abb84612bbdc555cb36a",
	}
	if len(fixture.SourceDocuments) != len(sourceHashes) {
		t.Fatal("changed source document set")
	}
	for name, hash := range sourceHashes {
		if fixture.SourceDocuments[name].SHA256 != hash {
			t.Fatal("unbound source document", name)
		}
	}
	public := displaylinesArrayOwnerSource(t, "displaylines_array_owner_original_public110.json.gz", sourceHashes["displaylines_array_owner_original_public110.json.gz"])
	direct := displaylinesArrayOwnerSource(t, "displaylines_array_owner_direct24_observations.json.gz", sourceHashes["displaylines_array_owner_direct24_observations.json.gz"])
	if len(public) != 110 || len(direct) != 24 {
		t.Fatal("changed preserved source observation counts")
	}
	inputs, indices, memberships := map[string]bool{}, map[int]bool{}, map[string]bool{}
	partitions, effects, publication := map[string]int{}, map[string]int{}, map[string]int{}
	for _, c := range fixture.Cases {
		if c.Display == nil || c.Name == "" || c.Index < 0 || c.Index >= 116 || len(c.Original) != 1 {
			t.Fatal("unbound input", c.Name)
		}
		input := fmt.Sprintf("%t:%s", *c.Display, c.TeX)
		if inputs[input] || indices[c.Index] || len(c.SourceMemberships) == 0 {
			t.Fatal("duplicate or missing source input", c.Name)
		}
		inputs[input], indices[c.Index] = true, true
		for _, m := range c.SourceMemberships {
			identity := fmt.Sprintf("%s:%d", m.Document, m.Index)
			if memberships[identity] {
				t.Fatal("duplicate source membership", identity)
			}
			memberships[identity] = true
			var tex string
			var display *bool
			var original map[string]string
			switch m.Document {
			case "displaylines_array_owner_original_public110.json.gz":
				if m.Index < 0 || m.Index >= len(public) || m.Profile != "public-default" {
					t.Fatal("unbound public source index")
				}
				var r struct {
					TeX         string
					Display     *bool
					PublicIndex int
					Original    map[string]string
				}
				if err := json.Unmarshal(public[m.Index], &r); err != nil || r.PublicIndex != m.Index {
					t.Fatal("changed public source", err)
				}
				tex, display, original = r.TeX, r.Display, r.Original
			case "displaylines_array_owner_direct24_observations.json.gz":
				if m.Index < 14 || m.Index >= len(direct) || m.Profile != "public-default-observed-full-input" {
					t.Fatal("constructed methods are not public inputs")
				}
				var r struct {
					Profile string
					Request struct {
						TeX, Profile  string
						Display       *bool
						Registrations []json.RawMessage
						Method        json.RawMessage
					}
					Plain struct{ OriginalAPI map[string]string }
				}
				if err := json.Unmarshal(direct[m.Index], &r); err != nil {
					t.Fatal(err)
				}
				if r.Profile != m.Profile || r.Request.Profile != m.Profile || len(r.Request.Registrations) != 0 || string(r.Request.Method) != "null" {
					t.Fatal("changed public observation profile")
				}
				tex, display, original = r.Request.TeX, r.Request.Display, r.Plain.OriginalAPI
			default:
				t.Fatal("unknown source membership", m.Document)
			}
			if display == nil || *display != *c.Display || tex != c.TeX || !reflect.DeepEqual(original, c.Original) {
				t.Fatal("source original changed", identity)
			}
		}
		partitions[c.Partition]++
		effects[c.Partition+":"+c.QualifiedChange]++
		publication[c.PublicationCategory]++
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = *c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if c.Partition == "bounded-original-runtime" {
				if c.Index != 62 && c.Index != 63 {
					t.Fatal("unbound runtime input")
				}
				if c.Contract.Kind != "bounded-null-range" || c.Contract.Codepoint != "U+0085" || c.Contract.Error != "no Unicode range for character U+0085" || c.Contract.OriginalFirstLine != "TypeError: Cannot read properties of null (reading '4')" {
					t.Fatal("changed runtime contract")
				}
				lines := strings.Split(c.Original["error"], "\n")
				if len(lines) != 11 || lines[0] != c.Contract.OriginalFirstLine {
					t.Fatal("incomplete original runtime")
				}
				for _, frame := range lines[1:] {
					if !strings.Contains(frame, "mathjax.js:") {
						t.Fatal("unbound original runtime frame")
					}
				}
				if got != "" || err == nil || err.Error() != c.Contract.Error {
					t.Fatalf("bounded API=%q/%v", got, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			want := c.Original["svg"]
			if want == "" {
				t.Fatal("missing full original SVG")
			}
			switch c.Partition {
			case "strict-svg":
				displaylinesArrayOwnerSVG(t, want)
			case "safe-error-attribute":
				// Only these four full source objects authorize this one attribute
				// escape. The mtext and every other byte remain source exact.
				if c.Index != 16 && c.Index != 17 && c.Index != 74 && c.Index != 75 {
					t.Fatal("unbound raw-invalid original")
				}
				if c.Contract.Kind != "single-error-attribute-escape" || c.Contract.SourcePublic110Index != c.Index || c.Contract.Raw != `data-mjx-error="Misplaced &"` || c.Contract.Escaped != `data-mjx-error="Misplaced &amp;"` || strings.Count(want, c.Contract.Raw) != 1 {
					t.Fatal("changed narrow XML contract")
				}
				var root struct{ XMLName xml.Name }
				if xml.Unmarshal([]byte(want), &root) == nil {
					t.Fatal("original invalid-XML boundary changed")
				}
				want = strings.Replace(want, c.Contract.Raw, c.Contract.Escaped, 1)
			default:
				t.Fatal("unknown source partition", c.Partition)
			}
			displaylinesArrayOwnerSVG(t, got)
			if got != want {
				t.Fatalf("original SVG contract differs: got %x want %x", sha256.Sum256([]byte(got)), sha256.Sum256([]byte(want)))
			}
		})
	}
	checks := []struct {
		Label     string
		Got, Want map[string]int
	}{
		{"partitions", partitions, map[string]int{"strict-svg": 110, "safe-error-attribute": 4, "bounded-original-runtime": 2}},
		{"effects", effects, map[string]int{"strict-svg:fix": 62, "strict-svg:control": 48, "safe-error-attribute:fix": 2, "safe-error-attribute:control": 2, "bounded-original-runtime:control": 2}},
		{"publication", publication, map[string]int{"firstStrictSVGInput": 102, "priorHeldCompleteOriginal": 8, "firstSafeXMLInput": 4, "firstOriginalRuntime": 2}},
	}
	for _, check := range checks {
		if !reflect.DeepEqual(check.Got, check.Want) {
			t.Fatal("changed", check.Label, check.Got, check.Want)
		}
	}
	if len(inputs) != 116 || len(indices) != 116 || len(memberships) != 120 {
		t.Fatal("changed complete source coverage")
	}
}
