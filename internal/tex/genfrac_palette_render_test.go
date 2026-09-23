// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
)

type genfracRegistration struct {
	Name, Body string
	Arguments  int
}
type genfracReference struct {
	Name, TeX, Group, SVG string
	Display               bool
	OuterMacroCount       int
	Registrations         []genfracRegistration
}
type genfracReferenceState struct {
	Source             string
	Cursor, MacroCount int
}
type genfracFixture struct {
	MathjaxGitCommit string
	Cases            []genfracReference
	Readers          []struct {
		Name, Command, PrimaryCommand string
		Before, After                 genfracReferenceState
		Key                           *string
		Error                         *Error
		Origins                       []struct {
			Name    string
			Ordinal int
		}
	}
	Fences []struct {
		Name, TeX   string
		Open, Close *string
	}
	ExcludedDiagnostics    []struct{ Name, Finding string }
	KnownContractResiduals []struct{ Name, Finding, Disposition string }
}

func genfracReferences(t *testing.T) genfracFixture {
	t.Helper()
	data, err := os.ReadFile("../../testdata/genfrac_palette_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var f genfracFixture
	if err = json.Unmarshal(data, &f); err != nil {
		t.Fatal(err)
	}
	if f.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(f.Cases) != 48 || len(f.Readers) != 45 || len(f.Fences) != 5 || len(f.ExcludedDiagnostics) != 6 || len(f.KnownContractResiduals) != 1 {
		t.Fatal("unbound genfrac references")
	}
	for _, c := range f.ExcludedDiagnostics {
		if c.Finding != "D120" {
			t.Fatal("unexpected excluded diagnostic")
		}
	}
	if f.KnownContractResiduals[0].Finding != "D121" {
		t.Fatal("missing StarMacro residual")
	}
	return f
}

func genfracExpectError(t *testing.T, err error, want *Error) {
	t.Helper()
	if want == nil {
		if err != nil {
			t.Fatal(err)
		}
		return
	}
	var got *Error
	if !errors.As(err, &got) || got.ID != want.ID || got.Message != want.Message {
		t.Fatalf("error %v; want %+v", err, want)
	}
}

// These twelve actual registered fixtures compare complete untouched primary
// SVGs. They do not assert full raw-model or StarMacro total-count parity.
func TestGenfracRegisteredWholeSVG(t *testing.T) {
	count := 0
	for _, c := range genfracReferences(t).Cases {
		if c.Group != "registered" {
			continue
		}
		count++
		t.Run(c.Name, func(t *testing.T) {
			root, err := genfracFixtureCompile(c)
			if err != nil {
				t.Fatal(err)
			}
			options := pipeline.DefaultOptions()
			options.Display = c.Display
			got, err := svg.NewTypesetter().Typeset(root, options)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.SVG {
				t.Fatalf("complete primary SVG differs\ngot: %s\nwant: %s", got, c.SVG)
			}
		})
	}
	if count != 12 {
		t.Fatalf("private references %d; want 12", count)
	}
}

// This test-only seam retains Compiler.Compile verbatim after installing the
// actual registered fixture macros and the observed outer counter.
func genfracFixtureCompile(c genfracReference) (*mml.Node, error) {
	source, display := c.TeX, c.Display
	state := newParseState()
	state.augmentedPackages = false
	state.macroCount = c.OuterMacroCount
	for _, registration := range c.Registrations {
		state.macros[registration.Name] = macroDefinition{body: registration.Body, arguments: registration.Arguments}
	}
	p := &parser{source: source, state: state, display: display}
	children, stop, err := p.parseRow(0, false)
	if err != nil {
		var parseError *Error
		if !errors.As(err, &parseError) {
			return nil, err
		}
		return mathError(parseError.Message, display), nil
	}
	if stop != "" {
		return nil, texError("ExtraCloseMissingOpen", "Extra close brace or missing open brace")
	}
	children, err = p.amsTagFinalize(children)
	if err != nil {
		var parseError *Error
		if !errors.As(err, &parseError) {
			return nil, err
		}
		return mathError(parseError.Message, display), nil
	}
	// MmlMath has an inferred-mrow child even when the TeX stack reduces to a
	// single node. Keep that wrapper explicit in the shared Go tree so layout
	// sees the same structure as the JavaScript MML factory.
	root := node("math", children...)
	if display {
		root.Attributes.Set("display", "block")
	}
	root.Walk(func(n *mml.Node) bool {
		n.RemoveProperty(resolvedFontScope)
		n.RemoveProperty(ambientFontSource)
		n.RemoveProperty(vectorFactoryToken)
		n.RemoveProperty(vectorFactoryDone)
		n.RemoveProperty(limitsScriptOrigin)
		return true
	})
	setMathMLInheritance(root, display)
	root = moveMathLimits(root)
	cleanMathMLAttributes(root)
	return root, nil
}
