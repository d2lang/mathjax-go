// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
)

type infixScopeRegistration struct {
	Name, Body string
	Arguments  int
}
type infixScopeCursor struct {
	Source     string
	ByteCursor int
	Remaining  string
}

func infixScopeProjection(n map[string]any) map[string]any {
	children := []any{}
	for _, c := range n["children"].([]any) {
		children = append(children, infixScopeProjection(c.(map[string]any)))
	}
	return map[string]any{"kind": n["kind"], "text": n["text"], "attributes": n["attributes"].(map[string]any)["explicit"], "properties": n["properties"], "children": children}
}
func infixScopeValue(t *testing.T, value any) any {
	t.Helper()
	b, e := json.Marshal(value)
	if e != nil {
		t.Fatal(e)
	}
	var result any
	if e = json.Unmarshal(b, &result); e != nil {
		t.Fatal(e)
	}
	return result
}
func TestInfixScopePinnedReferences(t *testing.T) {
	type output struct {
		SVGSHA256 string
		Tree      any
		Error     *Error
	}
	var fixture struct {
		Cases []struct {
			Name, TeX    string
			Display      bool
			Registration *infixScopeRegistration
			Cursor       *infixScopeCursor
			output
		}
	}
	var boundaries struct {
		Baseline  string
		Unchanged map[string]struct {
			TeX          string
			Display      bool
			Registration *infixScopeRegistration
			FullTree     any
			Parser       struct {
				Source          string
				ByteCursor      int
				Remaining, Stop string
				MacroCount      int
			}
			output
		}
	}
	for name, target := range map[string]any{"infix_scope_mathjax_3_2_2.json": &fixture, "infix_scope_boundaries.json": &boundaries} {
		b, e := os.ReadFile(filepath.Join("../../testdata", name))
		if e != nil {
			t.Fatal(e)
		}
		if e = json.Unmarshal(b, target); e != nil {
			t.Fatal(e)
		}
	}
	allowed := map[string]bool{"separate-args-inline": true, "unclosed-nested-inline": true, "unclosed-nested-display": true, "unsupported-delims-inline": true, "unsupported-delims-display": true, "registered-alias-inline": true, "registered-alias-display": true, "registered-above-inline": true, "registered-above-display": true, "nested-style-group-inline": true, "nested-style-group-display": true}
	if len(fixture.Cases) != 108 || len(boundaries.Unchanged) != len(allowed) || boundaries.Baseline != "2b829a798d95a4e93b6d8ff6d71b3b348b3d26fe" {
		t.Fatal("unbound infix scope corpus")
	}
	primary, unchanged := 0, 0
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			want := c.output
			b, bounded := boundaries.Unchanged[c.Name]
			if bounded {
				if !allowed[c.Name] || b.TeX != c.TeX || b.Display != c.Display || !reflect.DeepEqual(b.Registration, c.Registration) {
					t.Fatal("unexpected baseline qualification")
				}
				want = b.output
				unchanged++
			} else {
				primary++
			}
			state := newParseState()
			if c.Registration != nil {
				r := c.Registration
				state.macros[r.Name] = macroDefinition{body: r.Body, arguments: r.Arguments}
			}
			p := &parser{source: c.TeX, state: state, display: c.Display}
			children, stop, parseErr := p.parseRow(0, false)
			var actualError *Error
			if parseErr != nil && !errors.As(parseErr, &actualError) {
				t.Fatal(parseErr)
			}
			if !reflect.DeepEqual(actualError, want.Error) {
				t.Errorf("structured error %#v; want %#v", actualError, want.Error)
			}
			if stop != "" {
				t.Fatal("unexpected root stop", stop)
			}
			if c.Cursor != nil && !bounded {
				got := infixScopeCursor{p.source, p.pos, p.source[p.pos:]}
				if got != *c.Cursor {
					t.Errorf("source/cursor/tail %#v; want %#v", got, *c.Cursor)
				}
			}
			if bounded {
				if p.source != b.Parser.Source || p.pos != b.Parser.ByteCursor || p.source[p.pos:] != b.Parser.Remaining || stop != b.Parser.Stop || p.state.macroCount != b.Parser.MacroCount {
					t.Fatal("baseline parser state changed")
				}
			}
			var root *mml.Node
			if parseErr != nil {
				root = mathError(actualError.Message, c.Display)
			} else {
				var e error
				children, e = p.amsTagFinalize(children)
				if e != nil {
					t.Fatal(e)
				}
				root = node("math", children...)
				if c.Display {
					root.Attributes.Set("display", "block")
				}
				root.Walk(func(n *mml.Node) bool {
					for _, key := range []string{resolvedFontScope, ambientFontSource, vectorFactoryToken, vectorFactoryDone, limitsScriptOrigin} {
						n.RemoveProperty(key)
					}
					return true
				})
				setMathMLInheritance(root, c.Display)
				root = moveMathLimits(root)
				cleanMathMLAttributes(root)
			}
			full := infixMacroTree(t, root)
			if c.Registration == nil {
				compiled, e := NewCompiler().Compile(c.TeX, c.Display)
				if e != nil {
					t.Fatal(e)
				}
				if !reflect.DeepEqual(infixScopeValue(t, full), infixScopeValue(t, infixMacroTree(t, compiled))) {
					t.Fatal("registered capture finalization differs from public Compiler")
				}
			}
			if !reflect.DeepEqual(infixScopeValue(t, infixScopeProjection(full)), want.Tree) {
				t.Error("complete ordered explicit/own tree differs")
			}
			if bounded && !reflect.DeepEqual(infixScopeValue(t, full), b.FullTree) {
				t.Error("whole accepted tree boundary changed")
			}
			options := pipeline.DefaultOptions()
			options.Display = c.Display
			s, e := svg.NewTypesetter().Typeset(root, options)
			if e != nil {
				t.Fatal(e)
			}
			hash := fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
			svgMatches := hash == want.SVGSHA256
			if c.Name == "separate-args-inline" {
				// Both complete states predate D070: local darwin/arm64 retains
				// the saved baseline; darwin/amd64 and hosted linux/amd64 match
				// untouched primary. Keep tree, error and parser-state checks above.
				if !bounded || c.TeX != `\frac{x\choose y}{a\choose b}` || c.Display || c.Registration != nil ||
					b.SVGSHA256 != "ad36e1589c18d9616d0448892abd5e6452f91572f4e7b46ca065cdddaf000d7c" ||
					c.SVGSHA256 != "97b2abbcee31d8223582aa62735790c3b9a4eca6e9b3d4b20b2464d9a6217490" {
					t.Fatal("unbound nested-fraction output states")
				}
				svgMatches = svgMatches || hash == c.SVGSHA256
			}
			if !svgMatches {
				t.Errorf("whole SVG %s; want %s", hash, want.SVGSHA256)
			}
		})
	}
	if primary != 97 || unchanged != 11 {
		t.Fatal("changed qualification counts", primary, unchanged)
	}
}

func TestInfixScopeReentryAndUnwind(t *testing.T) {
	// An open group/left item closes its fraction before the following row. Reuse
	// the actual parser to detect scope leakage rather than inspecting a flag.
	for _, c := range []struct {
		name, source string
		term         byte
		right        bool
		stop         string
	}{
		{"brace", "x\\over y}a\\over b", '}', false, ""},
		{"right", "x\\over y\\right)a\\over b", 0, true, ")"},
	} {
		t.Run(c.name, func(t *testing.T) {
			p := &parser{source: c.source, state: newParseState()}
			first, stop, e := p.parseRow(c.term, c.right)
			if e != nil || stop != c.stop || len(first) != 1 || first[0].Kind != "mfrac" {
				t.Fatal("first logical row", stop, e)
			}
			second, stop, e := p.parseRow(0, false)
			if e != nil || stop != "" || len(second) != 1 || second[0].Kind != "mfrac" || p.pos != len(p.source) {
				t.Fatal("following independent row", stop, e)
			}
			if first[0] == second[0] || first[0].Parent != nil || second[0].Parent != nil {
				t.Fatal("independent fractions lost identity")
			}
		})
	}
	for _, source := range []string{"x\\over y\\atop z", "x\\over\\bf y\\atop z", "x\\over\\color{red}y\\atop z", "x\\over y\\above bad z"} {
		t.Run(source, func(t *testing.T) {
			p := &parser{source: source, state: newParseState(), activeFont: "italic"}
			_, _, err := p.parseRow(0, false)
			var typed *Error
			if !errors.As(err, &typed) {
				t.Fatal("missing first error", err)
			}
			if p.activeFont != "italic" {
				t.Fatal("font environment leaked through error")
			}
			p.source = "a\\over b"
			p.pos = 0
			nodes, stop, err := p.parseRow(0, false)
			if err != nil || stop != "" || len(nodes) != 1 || nodes[0].Kind != "mfrac" {
				t.Fatal("new row inherited failed fraction", stop, err)
			}
		})
	}
}
