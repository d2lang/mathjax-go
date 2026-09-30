// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func arrayOptionalPositionJSON(t *testing.T, name, hash string, target any) {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	if hash != "" {
		sum := sha256.Sum256(b)
		if hex.EncodeToString(sum[:]) != hash {
			t.Fatal("unbound original archive", name)
		}
	}
	z, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, target); err != nil {
		t.Fatal(err)
	}
}

// All public expectations come from the frozen original. The 48 historical
// held records stay unchanged; their previous mismatch is not a requirement.
func TestArrayOptionalPositionOriginalReferences(t *testing.T) {
	var fixture struct {
		OriginalPin, ActualMergedBase string
		Cases                         []struct {
			Name, TeX, Origin, SourceKind, OriginalSVGHash string
			PublicIndex                                    int
			Display                                        *bool
			SourceXMLValid                                 bool
			Original                                       map[string]string
			InputMetadata                                  struct {
				TeX     string
				Display *bool
			}
			Provenance struct {
				CompleteReuseRecord *struct {
					TeX                              string
					Display                          *bool
					Original                         map[string]string
					SourceIndexInResidual            int
					CompleteCurrentPublishedResidual json.RawMessage
					CurrentSource                    struct{ Commit, Path string }
				}
			}
		}
	}
	arrayOptionalPositionJSON(t, "array_optional_position_mathjax_3_2_2.json.gz", "3a3360481e88ab9d7afa4da9a1fa3609e08c6296a5ed224d5bb5cbb79e0d14e4", &fixture)
	if fixture.OriginalPin != "ad8f5c21cb810236551da8c6512ba733e67357ee" || fixture.ActualMergedBase != "16801f1f0e8e5c0b1582224cce85140b1093ed38" || len(fixture.Cases) != 106 {
		t.Fatal("unbound optional-position originals")
	}
	var published struct{ Cases []json.RawMessage }
	arrayOptionalPositionJSON(t, "ordinary_array_owner_residuals.json.gz", "", &published)
	type input struct {
		tex     string
		display bool
	}
	seen, names := map[input]bool{}, map[string]bool{}
	valid, errors, reused, fresh := 0, 0, 0, 0
	for index, c := range fixture.Cases {
		if c.PublicIndex != index || c.Display == nil || c.InputMetadata.Display == nil || c.Name == "" || c.TeX != c.InputMetadata.TeX || *c.Display != *c.InputMetadata.Display {
			t.Fatal("missing explicit source input", c.Name)
		}
		key := input{c.TeX, *c.Display}
		if seen[key] || names[c.Name] {
			t.Fatal("duplicate original", c.Name)
		}
		seen[key], names[c.Name] = true, true
		svg := c.Original["svg"]
		if len(c.Original) != 1 || svg == "" || !c.SourceXMLValid {
			t.Fatal("unexpected original API object", c.Name)
		}
		sum := sha256.Sum256([]byte(svg))
		if hex.EncodeToString(sum[:]) != c.OriginalSVGHash {
			t.Fatal("original SVG hash differs", c.Name)
		}
		var root struct{ XMLName xml.Name }
		if err := xml.Unmarshal([]byte(svg), &root); err != nil || root.XMLName.Local != "svg" || root.XMLName.Space != "http://www.w3.org/2000/svg" {
			t.Fatal("invalid original XML", c.Name, err)
		}
		if strings.Contains(svg, "data-mjx-error=") {
			if c.SourceKind != "error-svg" {
				t.Fatal("unbound source error", c.Name)
			}
			errors++
		} else {
			if c.SourceKind != "svg" {
				t.Fatal("unbound source SVG", c.Name)
			}
			valid++
		}
		if c.Origin == "exact-retained-original-reuse" {
			r := c.Provenance.CompleteReuseRecord
			if r == nil || r.Display == nil || r.TeX != c.TeX || *r.Display != *c.Display || !reflect.DeepEqual(r.Original, c.Original) || r.CurrentSource.Commit != fixture.ActualMergedBase || r.CurrentSource.Path != "testdata/ordinary_array_owner_residuals.json.gz" || r.SourceIndexInResidual < 0 || r.SourceIndexInResidual >= len(published.Cases) {
				t.Fatal("unbound retained original", c.Name)
			}
			var historical, actual any
			if err := json.Unmarshal(r.CompleteCurrentPublishedResidual, &historical); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(published.Cases[r.SourceIndexInResidual], &actual); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(historical, actual) {
				t.Fatal("complete historical residual changed", c.Name)
			}
			var held struct {
				Partition, Reason string
				Original          map[string]string
			}
			if err := json.Unmarshal(r.CompleteCurrentPublishedResidual, &held); err != nil {
				t.Fatal(err)
			}
			if held.Partition != "held-svg" || held.Reason != "AlignedArray optional setup (explicitly separate)" || !reflect.DeepEqual(held.Original, c.Original) {
				t.Fatal("retained source contract changed", c.Name)
			}
			reused++
		} else if c.Origin == "fresh-frozen-default-original" && c.Provenance.CompleteReuseRecord == nil {
			fresh++
		} else {
			t.Fatal("unrecognized original provenance", c.Name)
		}
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = *c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if got != svg {
				t.Fatalf("complete original differs: got %x want %x", sha256.Sum256([]byte(got)), sha256.Sum256([]byte(svg)))
			}
		})
	}
	if valid != 96 || errors != 10 || reused != 48 || fresh != 58 {
		t.Fatal("source counts changed", valid, errors, reused, fresh)
	}
}
