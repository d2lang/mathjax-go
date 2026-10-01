// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"os"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestPairedArgumentCountOriginalSVGs(t *testing.T) {
	f, err := os.Open("testdata/paired_argument_count_mathjax_3_2_2.json.gz")
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
		MathjaxGitCommit, GetArgCountSource, PairedDelimitersSource string
		Cases                                                       []struct {
			Name, TeX, SVG, CountType string
			Display, RenderedError    bool
			Registration              []json.RawMessage
		}
	}
	if err := json.NewDecoder(z).Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || !strings.Contains(fixture.GetArgCountSource, "trimSpaces") || !strings.Contains(fixture.GetArgCountSource, "^[0-9]+$") || fixture.PairedDelimitersSource == "" || len(fixture.Cases) != 494 {
		t.Fatal("unbound original paired count observations")
	}
	type input struct {
		tex     string
		display bool
	}
	names, inputs := make(map[string]bool), make(map[input]bool)
	errors := 0
	for _, c := range fixture.Cases {
		key := input{c.TeX, c.Display}
		if c.Name == "" || names[c.Name] || inputs[key] || c.SVG == "" || strings.Contains(c.SVG, `data-mml-node="merror"`) != c.RenderedError {
			t.Fatal("invalid original paired count input", c.Name)
		}
		names[c.Name], inputs[key] = true, true
		if c.RenderedError {
			errors++
		}
		if c.Name == "DeclarePairedDelimiterX-absent-inline" && (c.CountType != "undefined" || len(c.Registration) != 4 || string(c.Registration[3]) != "null") {
			t.Fatal("unbound original absent count")
		}
		if c.Name == "DeclarePairedDelimiterX-empty-inline" && (c.CountType != "string" || len(c.Registration) != 4 || string(c.Registration[3]) != `""`) {
			t.Fatal("unbound original empty count")
		}
		if c.Name == "DeclarePairedDelimiterX-zero-inline" && (c.CountType != "string" || len(c.Registration) != 4 || string(c.Registration[3]) != `"0"`) {
			t.Fatal("unbound original zero count")
		}
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.SVG {
				t.Fatalf("complete original differs: got %x want %x", sha256.Sum256([]byte(got)), sha256.Sum256([]byte(c.SVG)))
			}
		})
	}
	if errors != 104 {
		t.Fatal("unexpected rendered-error partition", errors)
	}
}
