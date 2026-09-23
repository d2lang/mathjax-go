// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
)

type flatfracRegistration struct {
	Name, Body string
	Arguments  int
}
type flatfracCase struct {
	Name, TeX, Group, SVG string
	Display               bool
	OuterMacroCount       int
	Registrations         []flatfracRegistration
}
type flatfracState struct {
	Source             string
	Cursor, MacroCount int
}
type flatfracInvocation struct {
	Name, Body           string
	Arity                int
	Before, After        flatfracState
	Arguments            []string
	ArgumentError, Error *Error
}
type flatfracFixture struct {
	MathjaxGitCommit string
	Cases            []flatfracCase
	MacroCases       []struct {
		Name, TeX       string
		Display         bool
		OuterMacroCount int
		Registrations   []flatfracRegistration
		Invocations     []flatfracInvocation
	}
	ExcludedDiagnostics      []struct{ Name, Finding string }
	UnclosedCursorBoundaries []struct {
		Name        string
		Primary, Go int
	}
}

func flatfracReferences(t *testing.T) flatfracFixture {
	t.Helper()
	data, err := os.ReadFile("testdata/flatfrac_registration_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var f flatfracFixture
	if err = json.Unmarshal(data, &f); err != nil {
		t.Fatal(err)
	}
	if f.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(f.Cases) != 28 || len(f.MacroCases) != 8 || len(f.ExcludedDiagnostics) != 4 || len(f.UnclosedCursorBoundaries) != 2 {
		t.Fatal("unbound flatfrac primary references")
	}
	return f
}
func flatfracExpectError(t *testing.T, got error, want *Error) {
	t.Helper()
	if want == nil {
		if got != nil {
			t.Fatal(got)
		}
		return
	}
	var typed *Error
	if !errors.As(got, &typed) || typed.ID != want.ID || typed.Message != want.Message {
		t.Fatalf("error %v; want %+v", got, want)
	}
}
func flatfracExpectState(t *testing.T, p *parser, want flatfracState) {
	t.Helper()
	got := flatfracState{Source: p.source, Cursor: p.pos, MacroCount: p.state.macroCount}
	if got != want {
		t.Fatalf("parser state %+v; want %+v", got, want)
	}
}

// Compare complete untouched primary SVG strings. The four retained D119/D095
// diagnostics are deliberately absent from this success corpus. No raw-tree or
// ordered-property equivalence is inferred from matching SVG.
func TestFlatfracWholeSVG(t *testing.T) {
	for _, c := range flatfracReferences(t).Cases {
		t.Run(c.Name, func(t *testing.T) {
			var root *mml.Node
			var err error
			if c.Group == "public" {
				if c.OuterMacroCount != 0 || len(c.Registrations) != 0 {
					t.Fatal("public fixture overrides state")
				}
				root, err = NewCompiler().Compile(c.TeX, c.Display)
			} else {
				// Only private fixtures set actual internal state; there is no
				// invented public registration/count option.
				root, err = flatfracFixtureCompile(c)
			}
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
}

// Invoke actual registered commands, including both steps of ff -> flatfrac.
// The separate readArgument scanner checks the recorded primary arguments; it
// is not an interception hook or a claim to observe invokeMacro's local slice.
// Actual source installation, consumed tail, errors and counts are then checked
// on the real command parser against independent primary before/after records.
func TestFlatfracMacroContracts(t *testing.T) {
	f := flatfracReferences(t)
	invocations := 0
	for _, c := range f.MacroCases {
		invocations += len(c.Invocations)
		t.Run(c.Name, func(t *testing.T) {
			state := newParseState()
			state.macroCount = c.OuterMacroCount
			for _, r := range c.Registrations {
				state.macros[r.Name] = macroDefinition{body: r.Body, arguments: r.Arguments}
			}
			p := &parser{source: c.TeX, state: state, display: c.Display}
			for index, want := range c.Invocations {
				p.skipSpaces()
				if p.pos >= len(p.source) || p.source[p.pos] != '\\' {
					t.Fatal("missing actual command at continuation")
				}
				p.pos++
				name := p.readControlSequence()
				if name != want.Name {
					t.Fatalf("command %q; want %q", name, want.Name)
				}
				flatfracExpectState(t, p, want.Before)
				definition, ok := p.state.macros[name]
				if !ok || definition.body != want.Body || definition.arguments != want.Arity || definition.optionalDefault != nil || definition.prefix != "" || len(definition.delimiters) != 0 {
					t.Fatalf("actual registration %q: %+v present=%v", name, definition, ok)
				}
				scanner := &parser{source: p.source, pos: p.pos, state: newParseState()}
				arguments := []string{}
				var argumentError error
				for i := 0; i < definition.arguments; i++ {
					arg, _, err := scanner.readArgument(name, false)
					if err != nil {
						argumentError = err
						break
					}
					arguments = append(arguments, arg)
				}
				if !reflect.DeepEqual(arguments, want.Arguments) {
					t.Fatalf("actual reader arguments %q; want %q", arguments, want.Arguments)
				}
				flatfracExpectError(t, argumentError, want.ArgumentError)
				result, err := p.commandEvent(name)
				flatfracExpectError(t, err, want.Error)
				flatfracExpectState(t, p, want.After)
				if p.state != state || len(result.nodes) != 0 || result.afterNode != nil || result.namedFunction {
					t.Fatal("Macro changed state ownership or returned an immediate node/action")
				}
				if err != nil && index != len(c.Invocations)-1 {
					t.Fatal("fixture continued after an actual command error")
				}
			}
		})
	}
	if invocations != 9 {
		t.Fatal("expected nine actual command invocations")
	}
}

// PRIVATE_COMPILER_BODY: generated from the exact current Compiler.Compile,
// with only the explicit fixture source/display/count/registration setup added.
func flatfracFixtureCompile(c flatfracCase) (*mml.Node, error) {
	source, display := c.TeX, c.Display
	state := newParseState()
	state.augmentedPackages = false
	state.macroCount = c.OuterMacroCount
	for _, r := range c.Registrations {
		state.macros[r.Name] = macroDefinition{body: r.Body, arguments: r.Arguments}
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
