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

type mathtoolsBigtimesCase struct {
	Name, TeX, Group, ExpansionControl, Qualification string
	Display                                           bool
	Original                                          struct{ SVG, Error string }
}

func mathtoolsBigtimesReferences(t *testing.T) ([]mathtoolsBigtimesCase, []mathtoolsBigtimesCase) {
	t.Helper()
	file, err := os.Open("testdata/mathtools_bigtimes_mathjax_3_2_2.json.gz")
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
		MathjaxGitCommit                                                          string
		OriginalValidCases, OriginalDiagnosticBoundaries, OriginalRuntimeFailures int
		Registration                                                              struct {
			Name              string
			MethodIsBaseMacro bool
			MethodArguments   []string
		}
		Cases                                   []mathtoolsBigtimesCase
		OriginalDiagnosticGoExtensionBoundaries []mathtoolsBigtimesCase
	}
	if err = json.NewDecoder(reader).Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" ||
		fixture.OriginalValidCases != 358 || fixture.OriginalDiagnosticBoundaries != 10 || fixture.OriginalRuntimeFailures != 0 ||
		len(fixture.Cases) != 358 || len(fixture.OriginalDiagnosticGoExtensionBoundaries) != 10 ||
		fixture.Registration.Name != "bigtimes" || !fixture.Registration.MethodIsBaseMacro ||
		len(fixture.Registration.MethodArguments) != 1 || fixture.Registration.MethodArguments[0] != `\mathop{\Large\kern-.1em\boldsymbol{\times}\kern-.1em}` {
		t.Fatal("unbound Mathtools bigtimes original references")
	}
	names, inputs := map[string]bool{}, map[struct {
		tex     string
		display bool
	}]bool{}
	for _, c := range append(append([]mathtoolsBigtimesCase(nil), fixture.Cases...), fixture.OriginalDiagnosticGoExtensionBoundaries...) {
		key := struct {
			tex     string
			display bool
		}{c.TeX, c.Display}
		if c.Name == "" || names[c.Name] || inputs[key] || c.Original.SVG == "" || c.Original.Error != "" {
			t.Fatal("invalid bigtimes original", c.Name)
		}
		names[c.Name], inputs[key] = true, true
	}
	return fixture.Cases, fixture.OriginalDiagnosticGoExtensionBoundaries
}

func TestMathtoolsBigtimesOriginalReferences(t *testing.T) {
	cases, _ := mathtoolsBigtimesReferences(t)
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.Group != "valid" || strings.Contains(c.Original.SVG, `data-mml-node="merror"`) {
				t.Fatal("original diagnostic credited as valid", c.Name)
			}
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.Original.SVG {
				t.Fatalf("complete original SVG differs\ngot: %s\nwant: %s", got, c.Original.SVG)
			}
		})
	}
}

// The Go-only boldsymbol extension intentionally predates this registration.
// Preserve complete original diagnostics without treating them as parity wins.
func TestMathtoolsBigtimesOriginalDiagnosticExtensionBoundary(t *testing.T) {
	_, boundaries := mathtoolsBigtimesReferences(t)
	for _, c := range boundaries {
		t.Run(c.Name, func(t *testing.T) {
			if c.Group != "original-diagnostic-go-extension-boundary" || c.Qualification == "" || c.ExpansionControl == "" ||
				!strings.Contains(c.Original.SVG, `data-mml-node="merror"`) ||
				!strings.Contains(c.Original.SVG, `data-mjx-error="Undefined control sequence \boldsymbol"`) {
				t.Fatal("lost original unavailable-dependency boundary")
			}
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			control, err := mathjax.RenderWithOptions(c.ExpansionControl, options)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(got, `data-mml-node="merror"`) || got != control {
				t.Fatal("bigtimes no longer uses the existing Go-only formatter normally")
			}
		})
	}
}
