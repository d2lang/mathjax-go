// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestLegacyMatrixFenceSpacingOriginalReferences(t *testing.T) {
	file, err := os.Open("testdata/matrix_fence_spacing_mathjax_3_2_2.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	var fixture struct {
		MathjaxGitCommit     string
		FreshOriginalPerCase bool
		Counts               struct {
			Total, FullSVGAssertions, Valid, Diagnostics, QualifiedDiagnostics, OriginalRuntimeErrors int
		}
		Cases []struct {
			Name, TeX, SVG         string
			Display, OriginalValid bool
		}
		QualifiedDiagnostics []struct {
			Original struct{ SVG string }
		}
		OriginalRuntimeErrors []json.RawMessage
	}
	if err := json.NewDecoder(reader).Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || !fixture.FreshOriginalPerCase ||
		fixture.Counts.Total != 1236 || fixture.Counts.FullSVGAssertions != 1220 || fixture.Counts.Valid != 1156 ||
		fixture.Counts.Diagnostics != 64 || fixture.Counts.QualifiedDiagnostics != 16 || fixture.Counts.OriginalRuntimeErrors != 0 ||
		len(fixture.Cases) != 1220 || len(fixture.QualifiedDiagnostics) != 16 || len(fixture.OriginalRuntimeErrors) != 0 {
		t.Fatal("unbound legacy Matrix fence references")
	}
	for _, c := range fixture.QualifiedDiagnostics {
		if !strings.Contains(c.Original.SVG, `data-mml-node="merror"`) || !strings.Contains(c.Original.SVG, `data-mjx-error="Undefined control sequence \boldsymbol"`) {
			t.Fatal("qualified original is not the retained unsupported-command diagnostic")
		}
	}
	seen := map[string]bool{}
	valid := 0
	for _, c := range fixture.Cases {
		if c.Name == "" || c.SVG == "" || seen[c.Name] || c.OriginalValid == strings.Contains(c.SVG, `data-mml-node="merror"`) {
			t.Fatal("invalid original reference", c.Name)
		}
		seen[c.Name] = true
		if c.OriginalValid {
			valid++
		}
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.SVG {
				t.Fatalf("complete original SVG differs\ngot: %s\nwant: %s", got, c.SVG)
			}
		})
	}
	if valid != fixture.Counts.Valid {
		t.Fatalf("valid originals: got %d, want %d", valid, fixture.Counts.Valid)
	}
}
