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

func TestCapitalKetbraOriginalSVGs(t *testing.T) {
	f, err := os.Open("testdata/capital_ketbra_mathjax_3_2_2.json.gz")
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
		MathjaxGitCommit string
		Registration     struct {
			Symbol, Handler, LowerHandler, MacroSource string
			Args                                       []json.RawMessage
		}
		Cases []struct {
			Name, TeX, SVG         string
			Display, RenderedError bool
		}
	}
	if err := json.NewDecoder(z).Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	registration := fixture.Registration
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || registration.Symbol != "Ketbra" || registration.Handler != "BaseMethods.Macro" || registration.LowerHandler != "PhysicsMethods.KetBra" || registration.MacroSource == "" || len(registration.Args) != 2 || len(fixture.Cases) != 184 {
		t.Fatal("unbound original Braket registration")
	}
	var body string
	var arguments int
	if json.Unmarshal(registration.Args[0], &body) != nil || json.Unmarshal(registration.Args[1], &arguments) != nil || body != `{\left\vert {#1} \right\rangle\left\langle {#2} \right\vert}` || arguments != 2 {
		t.Fatal("unexpected original macro body")
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
			t.Fatal("invalid original Braket input", c.Name)
		}
		names[c.Name], inputs[key] = true, true
		if c.RenderedError {
			errors++
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
	if errors != 10 {
		t.Fatal("unexpected original rendered-error partition", errors)
	}
}
