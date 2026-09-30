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
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

type numcasesDirectPopCase struct {
	Name, TeX, Origin, SourceKind, OriginalSVGHash, OriginalErrorHash, OriginalFirstLine string
	PublicIndex, RuntimeIndex, SourceUnionIndex                                         int
	Display                                                                           *bool
	SourceXMLValid                                                                    bool
	Original                                                                          map[string]string
	InputMetadata                                                                     struct {
		Name, TeX string
		Display   *bool
	}
	PublicationProvenance struct {
		Category   string
		FirstInput bool
	}
	Provenance struct {
		CompleteCurrentPublishedResidual json.RawMessage
		SourceIndexInResidual            int
		CurrentSource                    struct{ Commit, Path, SHA256 string }
		NewnessReceipt                   *struct{ Path, SHA256 string }
	}
}

type numcasesDirectPopFixture struct {
	OriginalPin, ActualMergedBase, SourceOriginalUnionSHA256, Partition string
	OriginalAPIOnly, OriginalTreeExpectations                           bool
	SourceCaptureCount                                                  int
	Cases                                                               []numcasesDirectPopCase
}

func numcasesDirectPopJSON(t *testing.T, name, hash string, target any) {
	t.Helper()
	fixturePath := filepath.Join("testdata", name)
	// Only these two new draft archives can be supplied from scratch. The
	// historical residual archive always comes from the tested checkout.
	if name == "numcases_direct_pop_mathjax_3_2_2.json.gz" || name == "numcases_direct_pop_original_runtime_mathjax_3_2_2.json.gz" {
		if directory := os.Getenv("MATHJAX_GO_NUMCASES_DIRECT_POP_FIXTURE_DIR"); directory != "" {
			fixturePath = filepath.Join(directory, name)
		}
	}
	b, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(b)
	if hex.EncodeToString(sum[:]) != hash {
		t.Fatal("unbound original archive", name)
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

func numcasesDirectPopHeader(t *testing.T, fixture numcasesDirectPopFixture, partition string, count int) {
	t.Helper()
	wantUnion, wantCaptures := "a4fda78dd875df1e1ffbdd452a48dd98d2602c3f5ac865d4bb39ec964c8650e3", 112
	if partition == "original-runtime-preservation-only" {
		// The twelve original runtime records keep the initial56 archive byte for byte.
		wantUnion, wantCaptures = "e09328fc33b63a604c95c93a9b29db3d8fe72854e425ed11a8d83ea06b1d5655", 56
	}
	if fixture.OriginalPin != "ad8f5c21cb810236551da8c6512ba733e67357ee" || fixture.ActualMergedBase != "c24058161bb5df6af3f269ce65030c94079b952a" || fixture.SourceOriginalUnionSHA256 != wantUnion || fixture.Partition != partition || fixture.SourceCaptureCount != wantCaptures || !fixture.OriginalAPIOnly || fixture.OriginalTreeExpectations || len(fixture.Cases) != count {
		t.Fatal("unbound direct-Pop original references")
	}
}

func numcasesDirectPopInput(t *testing.T, c numcasesDirectPopCase) {
	t.Helper()
	if c.Name == "" || c.Display == nil || c.InputMetadata.Display == nil || c.Name != c.InputMetadata.Name || c.TeX != c.InputMetadata.TeX || *c.Display != *c.InputMetadata.Display || c.SourceUnionIndex < 0 || c.SourceUnionIndex >= 112 {
		t.Fatal("missing explicit original input", c.Name)
	}
}

// These 100 full SVG expectations come exclusively from the frozen original.
// The eight historical held records retain their original fields and reason;
// their previous Go mismatch is not an expectation.
func TestNumCasesDirectPopOriginalReferences(t *testing.T) {
	var fixture numcasesDirectPopFixture
	numcasesDirectPopJSON(t, "numcases_direct_pop_mathjax_3_2_2.json.gz", "83dc943dffb364b534b363dd08ef6bde5452cdab6f953ae8c7ec25a1c59dffae", &fixture)
	numcasesDirectPopHeader(t, fixture, "strict-svg", 100)
	var published struct{ Cases []json.RawMessage }
	numcasesDirectPopJSON(t, "ordinary_array_owner_residuals.json.gz", "7877905982204984a0a19b69b3f6129d0df23a4d97524d8d77dbe9f5323dd209", &published)
	type input struct {
		tex     string
		display bool
	}
	seen, names, indices := map[input]bool{}, map[string]bool{}, map[int]bool{}
	valid, errors, reused, fresh := 0, 0, 0, 0
	for index, c := range fixture.Cases {
		numcasesDirectPopInput(t, c)
		key := input{c.TeX, *c.Display}
		if c.PublicIndex != index || seen[key] || names[c.Name] || indices[c.SourceUnionIndex] {
			t.Fatal("duplicate original input", c.Name)
		}
		seen[key], names[c.Name], indices[c.SourceUnionIndex] = true, true, true
		svg := c.Original["svg"]
		if len(c.Original) != 1 || svg == "" || !c.SourceXMLValid {
			t.Fatal("unexpected complete original API", c.Name)
		}
		sum := sha256.Sum256([]byte(svg))
		if hex.EncodeToString(sum[:]) != c.OriginalSVGHash {
			t.Fatal("original SVG hash differs", c.Name)
		}
		var root struct{ XMLName xml.Name }
		if err := xml.Unmarshal([]byte(svg), &root); err != nil || root.XMLName.Local != "svg" || root.XMLName.Space != "http://www.w3.org/2000/svg" {
			t.Fatal("original is not SVG XML", c.Name, err)
		}
		if strings.Contains(svg, "data-mjx-error=") {
			if c.SourceKind != "error-svg" {
				t.Fatal("unbound original error SVG", c.Name)
			}
			errors++
		} else {
			if c.SourceKind != "svg" {
				t.Fatal("unbound original SVG", c.Name)
			}
			valid++
		}
		if c.Origin == "exact-retained-original-reuse" {
			r := c.Provenance
			if c.PublicationProvenance.Category != "prior-held-svg-assertion-promotion" || c.PublicationProvenance.FirstInput || r.NewnessReceipt != nil || r.CurrentSource.Commit != fixture.ActualMergedBase || r.CurrentSource.Path != "testdata/ordinary_array_owner_residuals.json.gz" || r.CurrentSource.SHA256 != "7877905982204984a0a19b69b3f6129d0df23a4d97524d8d77dbe9f5323dd209" || r.SourceIndexInResidual < 0 || r.SourceIndexInResidual >= len(published.Cases) {
				t.Fatal("unbound historical original", c.Name)
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
				Name, TeX, Partition, Reason string
				Display                     bool
				Original                    map[string]string
			}
			if err := json.Unmarshal(r.CompleteCurrentPublishedResidual, &held); err != nil {
				t.Fatal(err)
			}
			if held.Name != c.Name || held.TeX != c.TeX || held.Display != *c.Display || held.Partition != "held-svg" || held.Reason != "Existing NumCases/SubNumCases direct Pop finalization" || !reflect.DeepEqual(held.Original, c.Original) {
				t.Fatal("historical original contract changed", c.Name)
			}
			reused++
		} else if c.Origin == "fresh-frozen-default-original" && c.PublicationProvenance.Category == "first-svg-input" && c.PublicationProvenance.FirstInput && len(c.Provenance.CompleteCurrentPublishedResidual) == 0 && c.Provenance.NewnessReceipt != nil {
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
	if valid != 48 || errors != 52 || reused != 8 || fresh != 92 {
		t.Fatal("original source counts changed", valid, errors, reused, fresh)
	}
}

// The original produced no SVG for these 12 inputs. Preserve and validate its
// complete error objects, without defining or rendering a Go substitute.
func TestNumCasesDirectPopOriginalRuntimePreservation(t *testing.T) {
	var fixture numcasesDirectPopFixture
	numcasesDirectPopJSON(t, "numcases_direct_pop_original_runtime_mathjax_3_2_2.json.gz", "ff61cca875471472028b067b9c854efb6780fc72365c83f4f1ce1fc110a1f195", &fixture)
	numcasesDirectPopHeader(t, fixture, "original-runtime-preservation-only", 12)
	names, indices := map[string]bool{}, map[int]bool{}
	copyFailures, spreadFailures := 0, 0
	for index, c := range fixture.Cases {
		numcasesDirectPopInput(t, c)
		if c.RuntimeIndex != index || names[c.Name] || indices[c.SourceUnionIndex] || c.SourceKind != "runtime" || c.SourceXMLValid || c.Origin != "fresh-frozen-default-original" || c.PublicationProvenance.Category != "first-runtime-input" || !c.PublicationProvenance.FirstInput || c.Provenance.NewnessReceipt == nil || len(c.Original) != 1 || c.Original["error"] == "" || c.Original["svg"] != "" {
			t.Fatal("unbound complete original runtime", c.Name)
		}
		names[c.Name], indices[c.SourceUnionIndex] = true, true
		full := c.Original["error"]
		sum := sha256.Sum256([]byte(full))
		if hex.EncodeToString(sum[:]) != c.OriginalErrorHash || !strings.HasPrefix(full, c.OriginalFirstLine+"\n") || strings.Count(full, "\n") < 9 {
			t.Fatal("complete original runtime changed", c.Name)
		}
		switch c.OriginalFirstLine {
		case "TypeError: Cannot read properties of undefined (reading 'copy')":
			copyFailures++
		case "TypeError: Cannot read properties of undefined (reading 'match')":
			spreadFailures++
		default:
			t.Fatal("unexpected original runtime boundary", c.Name)
		}
	}
	if copyFailures != 4 || spreadFailures != 8 {
		t.Fatal("original runtime counts changed", copyFailures, spreadFailures)
	}
}
