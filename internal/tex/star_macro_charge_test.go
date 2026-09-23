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

type starMacroReference struct {
	Name, TeX, SVG, ErrorID, ErrorMessage string
	Display, Public                       bool
	OuterMacroCount, FinalMacroCount      int
	Registrations                         []struct {
		Name, Body string
		Arguments  int
	}
}

// Whole SVG and caller counters come from the fourteen retained actual primary
// observations. Model/source/cursor equality is deliberately not asserted.
func TestStarMacroObservedPrimary(t *testing.T) {
	data, err := os.ReadFile("../../testdata/star_macro_charge_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []starMacroReference
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 14 || maxMacros != 1000 {
		t.Fatal("unbound StarMacro references/limit")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			state := newParseState()
			if c.Public {
				if c.OuterMacroCount != 0 || len(c.Registrations) != 0 || state.macroCount != 0 {
					t.Fatal("public initialization changed")
				}
			} else {
				state.macroCount = c.OuterMacroCount
				for _, r := range c.Registrations {
					state.macros[r.Name] = macroDefinition{body: r.Body, arguments: r.Arguments}
				}
			}
			p := &parser{source: c.TeX, state: state, display: c.Display}
			root, parseError, err := starMacroFixtureCompile(p)
			if err != nil {
				t.Fatal(err)
			}
			if p.state != state {
				t.Fatal("parser state replaced")
			}
			if c.ErrorID == "" {
				if parseError != nil {
					t.Fatalf("parse error %v; want success", parseError)
				}
			} else if parseError == nil || parseError.ID != c.ErrorID || parseError.Message != c.ErrorMessage {
				t.Fatalf("parse error %v; want %s: %s", parseError, c.ErrorID, c.ErrorMessage)
			}
			if state.macroCount != c.FinalMacroCount {
				t.Fatalf("caller macro count %d; want primary %d", state.macroCount, c.FinalMacroCount)
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

// Test-only Compiler.Compile seam: initialization is supplied by the caller,
// and the real parse error is retained alongside the compiler's rendered merror.
// All parsing, finalization, inheritance and cleanup operations stay unchanged.
func starMacroFixtureCompile(p *parser) (*mml.Node, *Error, error) {
	display := p.display
	children, stop, err := p.parseRow(0, false)
	if err != nil {
		var parseError *Error
		if !errors.As(err, &parseError) {
			return nil, nil, err
		}
		return mathError(parseError.Message, display), parseError, nil
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
		return mathError(parseError.Message, display), parseError, nil
	}
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
	return root, nil, nil
}
