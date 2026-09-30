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
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

type casesAliasLifetimeCase struct {
	Name, TeX, SourceKind, OriginalSVGHash, OriginalErrorHash string
	SourceUnionIndex, PartitionIndex                        int
	Display                                                *bool
	Original                                               map[string]string
	SourceMembership                                       struct {
		Family, RequestPath, RequestSHA256 string
		RequestIndex                      int
	}
}

type casesAliasLifetimeFixture struct {
	OriginalPin, ActualMergedBase, SourceOriginalUnionSHA256, Partition string
	OriginalAPIOnly, OriginalTreeExpectations                           bool
	SourceCaptureCount                                                  int
	Cases                                                               []casesAliasLifetimeCase
}

func casesAliasLifetimeJSON(t *testing.T, name, digest string) casesAliasLifetimeFixture {
	t.Helper()
	directory := "testdata"
	if scratch := os.Getenv("MATHJAX_GO_CASES_ALIAS_FIXTURE_DIR"); scratch != "" {
		directory = scratch
	}
	b, err := os.ReadFile(filepath.Join(directory, name))
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(b)
	if hex.EncodeToString(sum[:]) != digest {
		t.Fatal("unbound frozen original archive", name)
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
	var fixture casesAliasLifetimeFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func casesAliasLifetimeHeader(t *testing.T, fixture casesAliasLifetimeFixture, partition string, count int) {
	t.Helper()
	if fixture.OriginalPin != "ad8f5c21cb810236551da8c6512ba733e67357ee" || fixture.ActualMergedBase != "2a19a9626f9136931a5866ca14288d2d64998f0c" || fixture.SourceOriginalUnionSHA256 != "ce5a75b9828df72f980ae761c8100a40dc0b115984053e0d23f3ab23884317e8" || fixture.Partition != partition || fixture.SourceCaptureCount != 252 || !fixture.OriginalAPIOnly || fixture.OriginalTreeExpectations || len(fixture.Cases) != count {
		t.Fatal("unbound Cases alias original references")
	}
}

func casesAliasLifetimeInput(t *testing.T, c casesAliasLifetimeCase) {
	t.Helper()
	if c.Name == "" || c.TeX == "" || c.Display == nil || c.SourceUnionIndex < 0 || c.SourceUnionIndex >= 252 || c.SourceMembership.Family == "" || c.SourceMembership.RequestIndex < 0 || c.SourceMembership.RequestPath == "" || len(c.SourceMembership.RequestSHA256) != 64 {
		t.Fatal("missing explicit frozen input membership", c.Name)
	}
}

// All 196 full SVG expectations come only from the frozen original API.
// There are no candidate trees, bounded Go errors or normalized SVG goldens.
func TestCasesAliasLifetimeOriginalReferences(t *testing.T) {
	fixture := casesAliasLifetimeJSON(t, "cases_alias_lifetime_mathjax_3_2_2.json.gz", "00acc38cd1467cc4e8695a3593b774ebfa0bf69360811ef70603abae435fa8aa")
	casesAliasLifetimeHeader(t, fixture, "strict-svg", 196)
	type input struct {
		tex     string
		display bool
	}
	keys, names, indices := map[input]bool{}, map[string]bool{}, map[int]bool{}
	normal, diagnostics := 0, 0
	for index, c := range fixture.Cases {
		casesAliasLifetimeInput(t, c)
		key := input{c.TeX, *c.Display}
		if c.PartitionIndex != index || keys[key] || names[c.Name] || indices[c.SourceUnionIndex] {
			t.Fatal("duplicate original input", c.Name)
		}
		keys[key], names[c.Name], indices[c.SourceUnionIndex] = true, true, true
		svg := c.Original["svg"]
		if len(c.Original) != 1 || svg == "" {
			t.Fatal("unexpected whole original API", c.Name)
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
			if c.SourceKind != "diagnostic-svg" {
				t.Fatal("unbound diagnostic SVG", c.Name)
			}
			diagnostics++
		} else {
			if c.SourceKind != "normal-svg" {
				t.Fatal("unbound normal SVG", c.Name)
			}
			normal++
		}
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = *c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if got != svg {
				t.Fatalf("whole original SVG differs: got %x want %x", sha256.Sum256([]byte(got)), sha256.Sum256([]byte(svg)))
			}
		})
	}
	if normal != 32 || diagnostics != 164 {
		t.Fatal("frozen original partitions changed", normal, diagnostics)
	}
}

// These 56 original APIs contain no SVG. Preserve their complete raw errors
// without rendering or asserting any Go substitute for an original failure.
func TestCasesAliasLifetimeOriginalRuntimePreservation(t *testing.T) {
	fixture := casesAliasLifetimeJSON(t, "cases_alias_lifetime_original_runtime_mathjax_3_2_2.json.gz", "04977ed195a66881b5f98be3c6b9c19e5c4b903f57e9cd0b7dc564a2067e4765")
	casesAliasLifetimeHeader(t, fixture, "original-runtime-preservation-only", 56)
	names, indices := map[string]bool{}, map[int]bool{}
	copyFailures, matchFailures, appendFailures := 0, 0, 0
	for index, c := range fixture.Cases {
		casesAliasLifetimeInput(t, c)
		full := c.Original["error"]
		if c.PartitionIndex != index || names[c.Name] || indices[c.SourceUnionIndex] || c.SourceKind != "original-runtime" || len(c.Original) != 1 || full == "" {
			t.Fatal("unbound whole original runtime", c.Name)
		}
		names[c.Name], indices[c.SourceUnionIndex] = true, true
		sum := sha256.Sum256([]byte(full))
		if hex.EncodeToString(sum[:]) != c.OriginalErrorHash {
			t.Fatal("complete original runtime differs", c.Name)
		}
		if strings.HasPrefix(full, "TypeError: Cannot read properties of undefined (reading 'copy')\n") {
			copyFailures++
		} else if strings.HasPrefix(full, "TypeError: Cannot read properties of undefined (reading 'match')\n") {
			matchFailures++
		} else if strings.HasPrefix(full, "TypeError: Cannot read properties of undefined (reading 'appendChild')\n") {
			appendFailures++
		} else {
			t.Fatal("unexpected original runtime boundary", c.Name)
		}
	}
	if copyFailures != 44 || matchFailures != 4 || appendFailures != 8 {
		t.Fatal("original runtime partitions changed", copyFailures, matchFailures, appendFailures)
	}
}
