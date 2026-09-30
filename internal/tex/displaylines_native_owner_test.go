// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
)

type displaylinesNativeRegistration struct {
	Name, Body string
	Arguments  int
}

type displaylinesNativeOutput struct {
	Name        string
	OriginalAPI map[string]string
	Output      struct {
		Tree            json.RawMessage
		Error           json.RawMessage
		FormattedErrors []struct{ ID, Message, Type string }
	}
	SourceBindings       json.RawMessage
	FormattedErrorStacks []string
	Trace                json.RawMessage
}

// This is Compiler.Compile's complete pipeline at actual PR182, with only
// native Macro registrations injected before parsing and diagnostics exposed.
// It does not load newcommand, seed counters, pre-expand arguments, or replace
// the Matrix/Array handlers. The literal MathJax profile is retained in full.
func displaylinesNativeCompile(source string, display bool, registrations []displaylinesNativeRegistration) (*mml.Node, []*Error, error) {
	state := newParseState()
	for _, r := range registrations {
		state.macros[r.Name] = macroDefinition{body: r.Body, arguments: r.Arguments}
	}
	p := &parser{source: source, state: state, display: display}
	children, stop, err := p.parseRow(0, false)
	if err != nil {
		var parseError *Error
		if !errors.As(err, &parseError) {
			return nil, nil, err
		}
		return mathError(parseError.Message, display), []*Error{parseError}, nil
	}
	if stop != "" {
		return nil, nil, texError("ExtraCloseMissingOpen", "Extra close brace or missing open brace")
	}
	children, err = p.amsTagFinalize(children)
	if err != nil {
		var parseError *Error
		if !errors.As(err, &parseError) {
			return nil, nil, err
		}
		return mathError(parseError.Message, display), []*Error{parseError}, nil
	}
	root := node("math", children...)
	if display {
		root.Attributes.Set("display", "block")
	}
	root, err = cleanPoppedScripts(root, state.poppedScripts)
	if err != nil {
		return nil, nil, err
	}
	root.Walk(func(n *mml.Node) bool {
		n.RemoveProperty(poppedScriptOrigin)
		n.RemoveProperty(resolvedFontScope)
		n.RemoveProperty(ambientFontSource)
		n.RemoveProperty(vectorFactoryToken)
		n.RemoveProperty(vectorFactoryDone)
		n.RemoveProperty(limitsScriptOrigin)
		return true
	})
	stretchy := stretchyCleanupOperators(state.operators)
	setMathMLInheritance(root, display)
	if err := filterNonscript(state.nonscriptSpaces); err != nil {
		return nil, nil, err
	}
	root = moveMathLimits(root)
	cleanStretchy(root, stretchy)
	cleanMathMLAttributes(root)
	state.operators = combineRelations(root, state.operators)
	return root, nil, nil
}

func displaylinesNativeJSON(t *testing.T, raw json.RawMessage) any {
	t.Helper()
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	return v
}

func displaylinesNativeValidSVG(s string) bool {
	decoder := xml.NewDecoder(strings.NewReader(s))
	depth, roots := 0, 0
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return depth == 0 && roots == 1
		}
		if err != nil {
			return false
		}
		switch token := token.(type) {
		case xml.StartElement:
			if depth == 0 {
				roots++
				if roots != 1 || token.Name.Local != "svg" || token.Name.Space != "http://www.w3.org/2000/svg" {
					return false
				}
			}
			depth++
		case xml.EndElement:
			depth--
		case xml.CharData:
			if depth == 0 && strings.TrimSpace(string(token)) != "" {
				return false
			}
		}
	}
}

func TestDisplaylinesNativeOwnerOriginalReferences(t *testing.T) {
	// This file is the unchanged complete original capture, including both
	// native observation lanes, all registration bindings, events and stacks.
	compressed, err := os.ReadFile("../../testdata/displaylines_native_owner_observations.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(compressed)) != "eaf2e8d9ae60dfc720b9dc0ff45b1e8069007022e940abb84612bbdc555cb36a" {
		t.Fatal("unbound native original capture")
	}
	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Request struct {
				Name, TeX     string
				Display       bool
				Registrations []displaylinesNativeRegistration
			}
			OriginalPin, PrimarySHA256, Profile string
			AssetSHA256                         map[string]string
			Plain, Observed                     displaylinesNativeOutput
			InfrastructureError                 json.RawMessage
			Equivalence                         struct{ FullAPIEqual, OutputEqual, SourceBindingsEqual, PassiveEquivalent, FormattedErrorStacksEqual bool }
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 28 {
		t.Fatal("changed native profile inventory")
	}
	seen, names := map[string]bool{}, map[string]bool{}
	strictSVG, rawInvalidSVG, formattedDiagnostics := 0, 0, 0
	for _, c := range fixture.Cases {
		r := c.Request
		identity, err := json.Marshal(struct {
			TeX           string
			Display       bool
			Registrations []displaylinesNativeRegistration
		}{r.TeX, r.Display, r.Registrations})
		if err != nil {
			t.Fatal(err)
		}
		if r.Name == "" || names[r.Name] || seen[string(identity)] || len(r.Registrations) != 1 ||
			c.OriginalPin != "ad8f5c21cb810236551da8c6512ba733e67357ee" ||
			c.PrimarySHA256 != "1b9a8c9eaf2fbbc196d4ef75ba6a0b7a3da298e11693fb8fef65cb819b7c541b" ||
			c.Profile != "seeded-Macro-native-full-input" {
			t.Fatal("changed source profile", r.Name)
		}
		seen[string(identity)], names[r.Name] = true, true
		for name, want := range map[string]string{
			"polyfills.js": "7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01",
			"mathjax.js":   "cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869",
			"setup.js":     "a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881",
		} {
			if c.AssetSHA256[name] != want {
				t.Fatal("unbound frozen asset", name)
			}
		}
		if c.Plain.Name != r.Name || c.Observed.Name != r.Name ||
			len(c.Plain.OriginalAPI) != 1 || c.Plain.OriginalAPI["svg"] == "" ||
			!reflect.DeepEqual(c.Plain.OriginalAPI, c.Observed.OriginalAPI) ||
			!reflect.DeepEqual(c.Plain.Output, c.Observed.Output) ||
			!reflect.DeepEqual(c.Plain.FormattedErrorStacks, c.Observed.FormattedErrorStacks) ||
			!reflect.DeepEqual(displaylinesNativeJSON(t, c.Plain.SourceBindings), displaylinesNativeJSON(t, c.Observed.SourceBindings)) ||
			displaylinesNativeJSON(t, c.Plain.Output.Error) != nil ||
			displaylinesNativeJSON(t, c.InfrastructureError) != nil ||
			displaylinesNativeJSON(t, c.Plain.Trace) != nil ||
			displaylinesNativeJSON(t, c.Observed.Trace) == nil ||
			!c.Equivalence.FullAPIEqual || !c.Equivalence.OutputEqual || !c.Equivalence.SourceBindingsEqual ||
			!c.Equivalence.PassiveEquivalent || !c.Equivalence.FormattedErrorStacksEqual {
			t.Fatal("changed complete native observation", r.Name)
		}
		wantSVG := c.Plain.OriginalAPI["svg"]
		validSVG := displaylinesNativeValidSVG(wantSVG)
		if validSVG {
			strictSVG++
		} else {
			rawInvalidSVG++
		}
		formattedDiagnostics += len(c.Plain.Output.FormattedErrors)
		t.Run(r.Name, func(t *testing.T) {
			root, diagnostics, err := displaylinesNativeCompile(r.TeX, r.Display, r.Registrations)
			if err != nil {
				t.Fatalf("native compile API error: %v", err)
			}
			if len(diagnostics) != len(c.Plain.Output.FormattedErrors) {
				t.Fatal("changed formatted diagnostic count")
			}
			for i, want := range c.Plain.Output.FormattedErrors {
				if diagnostics[i].ID != want.ID || diagnostics[i].Message != want.Message {
					t.Fatalf("diagnostic=%+v want %s/%s", diagnostics[i], want.ID, want.Message)
				}
			}
			encoded, err := json.Marshal(macroReferenceTree(root))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(displaylinesNativeJSON(t, encoded), displaylinesNativeJSON(t, c.Plain.Output.Tree)) {
				t.Fatalf("complete original MathML tree differs\ngot %s\nwant %s", encoded, c.Plain.Output.Tree)
			}
			options := pipeline.DefaultOptions()
			options.Display = r.Display
			rendered, err := svg.NewTypesetter().Typeset(root, options)
			if err != nil {
				t.Fatal(err)
			}
			if !displaylinesNativeValidSVG(rendered) {
				t.Fatal("native Go result is not valid SVG XML")
			}
			if validSVG {
				if rendered != wantSVG {
					t.Fatal("complete original native SVG differs")
				}
			} else {
				// The source SVG has an unescaped '&' in data-mjx-error. Preserve
				// its full bytes without inventing a matching Go SVG or silently
				// normalizing it. These two assertions cover original full MathML,
				// diagnostic ID/message and a successful XML-valid Go API only.
				// They are not counted as SVG parity or a safe-XML promotion.
				if !strings.HasPrefix(r.Name, "displaylines-override-") || len(diagnostics) != 1 ||
					diagnostics[0].ID != "Misplaced" || diagnostics[0].Message != "Misplaced &" {
					t.Fatal("changed raw-invalid native source boundary")
				}
			}
		})
	}
	if strictSVG != 26 || rawInvalidSVG != 2 || formattedDiagnostics != 6 {
		t.Fatal("changed source output partition", strictSVG, rawInvalidSVG, formattedDiagnostics)
	}
}
