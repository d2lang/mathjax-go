// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"
	"unicode/utf16"

	"github.com/d2lang/mathjax-go/internal/mml"
)

type ordinaryArrayAugmentedOutput struct {
	Name            string
	SVG             string
	Tree            json.RawMessage
	Error           *Error
	FormattedErrors []Error
	Trace           *struct {
		FinalParsers []struct {
			Source, Remaining       string
			CursorUTF16, MacroCount int
		}
	}
}

type ordinaryArrayAugmentedCase struct {
	Request struct {
		Name, TeX, Profile, Environment, Declaration, Invocation string
		Display                                                  bool
		CountBeforeBegin                                         int
		InitialCount                                             *int
		Registrations                                            []macroReferenceRegistration
	}
	Plain, Observed ordinaryArrayAugmentedOutput
}

func TestOrdinaryArrayLiteralBeginAugmentedReferences(t *testing.T) {
	file, err := os.Open("../../testdata/ordinary_array_owner_augmented_methods.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Profile                         string
		Pins                            map[string]string
		SupplementSHA256, PrimarySHA256 string
		Cases                           []ordinaryArrayAugmentedCase
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Profile != "official-newcommand-augmented; method state seeded using existing repository primary harness; not frozen public coverage" ||
		fixture.SupplementSHA256 != "a41669a4ae924ab83cbc3d08f95ae90490a33cd649ed865e83675011b8708ab9" ||
		fixture.PrimarySHA256 != "e59040261f8024d42d1c1c1e5d893ab807f8d0625790d5242d215fc285a59432" || len(fixture.Cases) != 48 {
		t.Fatal("unbound augmented source profile")
	}
	for name, want := range map[string]string{
		"polyfills.js": "7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01",
		"mathjax.js":   "cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869",
		"setup.js":     "a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881",
	} {
		if fixture.Pins[name] != want {
			t.Fatal("unbound frozen asset", name)
		}
	}
	beginFailures, successfulOutputs, heldEndFailures := 0, 0, 0
	names, methodKeys := make(map[string]bool), make(map[string]bool)
	for _, c := range fixture.Cases {
		r := c.Request
		methodKey, err := json.Marshal(struct {
			TeX          string
			Display      bool
			InitialCount *int
		}{r.TeX, r.Display, r.InitialCount})
		if err != nil {
			t.Fatal(err)
		}
		if r.Name == "" || names[r.Name] || methodKeys[string(methodKey)] ||
			c.Plain.Name != r.Name || c.Observed.Name != r.Name ||
			c.Plain.SVG == "" || !json.Valid(c.Plain.Tree) || string(c.Plain.Tree) == "null" {
			t.Fatal("missing, duplicate, or unbound augmented method observation", r.Name)
		}
		names[r.Name], methodKeys[string(methodKey)] = true, true
		if r.Profile != "official-newcommand-augmented" || r.Display || c.Plain.Error != nil ||
			c.Observed.Trace == nil || len(c.Observed.Trace.FinalParsers) != 1 {
			t.Fatal("changed augmented profile or original runtime boundary", r.Name)
		}
		originalFinal := c.Observed.Trace.FinalParsers[0]
		literal := r.Invocation == r.Environment
		if literal && r.CountBeforeBegin == 999 {
			// This complete original is retained, but the inherited user-defined
			// environment capture omits its End charge. Do not invent a matching
			// Go output or replay a synthetic End to conceal that separate gap.
			if len(c.Plain.FormattedErrors) != 1 || c.Plain.FormattedErrors[0].ID != "MaxMacroSub2" || originalFinal.MacroCount != 1001 {
				t.Fatal("changed held user-defined End boundary", r.Name)
			}
			heldEndFailures++
			continue
		}
		terminalBegin := !literal || r.CountBeforeBegin == 1000
		if terminalBegin {
			beginFailures++
		} else {
			successfulOutputs++
		}
		t.Run(r.Name, func(t *testing.T) {
			state := newParseState()
			state.augmentedPackages = true
			if r.InitialCount != nil {
				state.macroCount = *r.InitialCount
			}
			for _, registration := range r.Registrations {
				state.macros[registration.Name] = macroReferenceDefinition(registration)
			}
			p := &parser{source: r.TeX, state: state, display: r.Display}
			children, stop, err := p.parseRow(0, false)
			if stop != "" {
				t.Fatal("unexpected closing token", stop)
			}
			definition, registered := state.environments[r.Environment]
			if !registered || definition.begin != "(" || definition.end != ")" {
				t.Fatal("authored definition was not registered under its literal trimmed declaration name")
			}
			var root *mml.Node
			if terminalBegin {
				if len(c.Plain.FormattedErrors) != 1 || originalFinal.MacroCount != r.CountBeforeBegin+1 {
					t.Fatal("changed original Begin failure boundary")
				}
				macroReferenceError(t, err, &c.Plain.FormattedErrors[0])
				if p.state != state || state.macroCount != originalFinal.MacroCount ||
					p.environmentOwner != nil || p.environmentRow != nil || p.ordinaryArray != nil {
					t.Fatal("Begin failure changed the original charge or created an environment owner")
				}
				if p.pos < 0 || p.pos > len(p.source) || p.source != originalFinal.Source ||
					len(utf16.Encode([]rune(p.source[:p.pos]))) != originalFinal.CursorUTF16 ||
					p.source[p.pos:] != originalFinal.Remaining {
					t.Fatal("Begin failure consumed or rewrote the original remaining input")
				}
				root = mathError(c.Plain.FormattedErrors[0].Message, r.Display)
			} else {
				if len(c.Plain.FormattedErrors) != 0 || originalFinal.MacroCount != r.CountBeforeBegin+2 {
					t.Fatal("changed original successful environment boundary")
				}
				macroReferenceError(t, err, nil)
				children, err = p.amsTagFinalize(children)
				if err != nil {
					t.Fatal(err)
				}
				root = operatorNameFinalize(children, r.Display)
				// These twelve assertions cover full tree/SVG output only. The
				// same inherited missing-End-charge gap keeps their Go counter
				// one short; preserve the original count and do not bless that Go
				// count as a source expectation or claim lifecycle counter parity.
			}
			operatorNameReferenceOutput(t, root, r.Display, macroReferenceOutput{
				Tree: c.Plain.Tree, SVGSHA256: fmt.Sprintf("%x", sha256.Sum256([]byte(c.Plain.SVG))),
			})
		})
	}
	if beginFailures != 30 || successfulOutputs != 12 || heldEndFailures != 6 {
		t.Fatalf("changed source partitions: Begin failures=%d successful outputs=%d held End failures=%d", beginFailures, successfulOutputs, heldEndFailures)
	}
}
