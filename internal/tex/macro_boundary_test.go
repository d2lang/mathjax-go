// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/ordered"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
	"os"
	"reflect"
	"testing"
	"unicode/utf16"
)

type macroReferenceRegistration struct {
	Name, Body, Handler string
	Arguments           int
	OptionalDefault     *string
	Parameters          []string
}
type macroReferenceOutput struct {
	Tree      json.RawMessage
	SVGSHA256 string
	Error     *Error
	Reason    string
}
type macroReferenceCase struct {
	Name, TeX, Profile string
	Display            bool
	InitialCount       int
	Registrations      []macroReferenceRegistration
	Primary            macroReferenceOutput
	Boundary           *macroReferenceOutput
}
type macroReferenceState struct {
	Source                  string
	CursorUTF16, MacroCount int
	Remaining               string
}
type macroReferenceFixture struct {
	MathjaxGitCommit string
	ReuseCompiler    []struct {
		Name, TeX string
		Primary   macroReferenceOutput
	}
	Public  []macroReferenceCase
	Helpers []struct {
		Name, S1, S2 string
		Body         *string
		Args         []string
		Limit        int
		Result       *string
		Error        *Error
	}
	Invocations []struct {
		Name          string
		Registration  macroReferenceRegistration
		Before, After macroReferenceState
		Error         *Error
	}
}

func macroReferences(t *testing.T) macroReferenceFixture {
	t.Helper()
	data, err := os.ReadFile("testdata/macro_boundary_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var f macroReferenceFixture
	if err = json.Unmarshal(data, &f); err != nil {
		t.Fatal(err)
	}
	if f.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(f.Public) != 85 || len(f.Helpers) != 28 || len(f.Invocations) != 22 || len(f.ReuseCompiler) != 3 {
		t.Fatal("unbound actual primary references")
	}
	return f
}
func macroReferenceDefinition(r macroReferenceRegistration) macroDefinition {
	d := macroDefinition{body: r.Body, arguments: r.Arguments, optionalDefault: r.OptionalDefault}
	if r.Handler == "MacroWithTemplate" && len(r.Parameters) > 0 {
		d.prefix = r.Parameters[0]
		d.delimiters = r.Parameters[1:]
	}
	return d
}
func macroReferenceError(t *testing.T, err error, want *Error) {
	t.Helper()
	if want == nil {
		if err != nil {
			t.Fatal(err)
		}
		return
	}
	var actual *Error
	if !errors.As(err, &actual) || actual.ID != want.ID || actual.Message != want.Message {
		t.Fatalf("error %v; want %+v", err, want)
	}
}
func macroReferencePairs(m *ordered.Map[mml.Property]) []map[string]any {
	r := []map[string]any{}
	m.Range(func(k string, v any) bool {
		if mml.IsInherit(v) {
			v = "_inherit_"
		}
		r = append(r, map[string]any{"name": k, "value": v})
		return true
	})
	return r
}
func macroReferenceTree(n *mml.Node) map[string]any {
	children := []any{}
	for _, c := range n.Children {
		if c.Parent != n {
			panic("child ownership changed")
		}
		children = append(children, macroReferenceTree(c))
	}
	kind := n.Kind
	if kind == "mrow" && n.Flags.Inferred {
		kind = "inferredMrow"
	}
	var text any
	if n.Kind == "text" {
		text = n.Text
	}
	return map[string]any{"kind": kind, "text": text, "attributes": map[string]any{"explicit": macroReferencePairs(n.Attributes.Explicit()), "inherited": macroReferencePairs(n.Attributes.Inherited()), "defaults": macroReferencePairs(n.Attributes.Defaults()), "global": macroReferencePairs(n.Attributes.Globals())}, "properties": macroReferencePairs(n.Properties), "children": children}
}
func TestMacroBoundaryPublicReferences(t *testing.T) {
	f := macroReferences(t)
	promotedHashes := 0
	for _, c := range f.Public {
		t.Run(c.Name, func(t *testing.T) {
			state := newParseState()
			state.augmentedPackages = c.Profile == "official-newcommand-augmented"
			state.macroCount = c.InitialCount
			for _, r := range c.Registrations {
				state.macros[r.Name] = macroReferenceDefinition(r)
			}
			p := &parser{source: c.TeX, state: state, display: c.Display}
			children, stop, err := p.parseRow(0, false)
			if stop != "" {
				t.Fatal("unexpected stop", stop)
			}
			want := c.Primary
			macroReferenceBoundary(t, c)
			// Retain historical hash boundaries in the original fixture, but
			// require their complete untouched primary outputs after D098.
			promotedHash := c.Boundary != nil && c.Boundary.Reason == "D098 diagnostic quotation marks"
			if promotedHash {
				promotedHashes++
				if c.Primary.Error == nil || c.Primary.Error.ID != "CantUseHash1" || c.Primary.Error.Message != "You can't use 'macro parameter character #' in math mode" {
					t.Fatal("unbound hash promotion")
				}
			} else if c.Boundary != nil {
				want = *c.Boundary
			}
			macroReferenceError(t, err, want.Error)
			var root *mml.Node
			if err != nil {
				var typed *Error
				if !errors.As(err, &typed) {
					t.Fatal(err)
				}
				root = mathError(typed.Message, c.Display)
			} else {
				children, err = p.amsTagFinalize(children)
				if err != nil {
					t.Fatal(err)
				}
				root = node("math", children...)
				if c.Display {
					root.Attributes.Set("display", "block")
				}
				root.Walk(func(n *mml.Node) bool {
					for _, k := range []string{resolvedFontScope, ambientFontSource, vectorFactoryToken, vectorFactoryDone, limitsScriptOrigin} {
						n.RemoveProperty(k)
					}
					return true
				})
				setMathMLInheritance(root, c.Display)
				root = moveMathLimits(root)
				cleanMathMLAttributes(root)
			}
			encoded, err := json.Marshal(macroReferenceTree(root))
			if err != nil {
				t.Fatal(err)
			}
			var actual, expected any
			if err = json.Unmarshal(encoded, &actual); err != nil {
				t.Fatal(err)
			}
			if err = json.Unmarshal(want.Tree, &expected); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("complete explicit/own/inherited tree differs: %s", encoded)
			}
			opts := pipeline.DefaultOptions()
			opts.Display = c.Display
			rendered, err := svg.NewTypesetter().Typeset(root, opts)
			if err != nil {
				t.Fatal(err)
			}
			if fmt.Sprintf("%x", sha256.Sum256([]byte(rendered))) != want.SVGSHA256 {
				t.Fatalf("complete SVG differs: %s", rendered)
			}
		})
	}
	if promotedHashes != 4 {
		t.Fatal("changed hash promotion count", promotedHashes)
	}
}
func TestMacroBoundaryHelpers(t *testing.T) {
	for _, c := range macroReferences(t).Helpers {
		t.Run(c.Name, func(t *testing.T) {
			var result string
			var err error
			if c.Body != nil {
				result, err = substituteMacroArgumentsWithLimit(*c.Body, c.Args, c.Limit)
			} else {
				result, err = macroAddArgs(c.S1, c.S2, c.Limit)
			}
			macroReferenceError(t, err, c.Error)
			if c.Result != nil && result != *c.Result {
				t.Fatalf("result %q; want %q", result, *c.Result)
			}
		})
	}
}

// These are actual registered Macro/Template entry/return/error receipts. Custom
// primary maxBuffer cases are not represented as a configurable Go API; fresh
// default-5120 records exercise the real production handler instead.
func TestMacroBoundaryInvocationState(t *testing.T) {
	for _, c := range macroReferences(t).Invocations {
		t.Run(c.Name, func(t *testing.T) {
			pos, units := 0, 0
			for _, r := range c.Before.Source {
				if units == c.Before.CursorUTF16 {
					break
				}
				units += len(utf16.Encode([]rune{r}))
				pos += len(string(r))
			}
			if units != c.Before.CursorUTF16 {
				t.Fatal("invalid primary cursor")
			}
			state := newParseState()
			state.macroCount = c.Before.MacroCount
			state.macros["kept"] = macroDefinition{body: "q"}
			p := &parser{source: c.Before.Source, pos: pos, state: state}
			definition := macroReferenceDefinition(c.Registration)
			nodes, err := p.invokeMacro(c.Registration.Name, definition)
			macroReferenceError(t, err, c.Error)
			if len(nodes) != 0 || p.state != state || state.macros["kept"].body != "q" {
				t.Fatal("macro reset or replaced shared parse state/returned an item")
			}
			if p.pos < 0 || p.pos > len(p.source) {
				t.Fatal("invalid byte cursor")
			}
			actual := macroReferenceState{Source: p.source, CursorUTF16: len(utf16.Encode([]rune(p.source[:p.pos]))), Remaining: p.source[p.pos:], MacroCount: p.state.macroCount}
			if actual != c.After {
				t.Fatalf("state %+v; want %+v", actual, c.After)
			}
		})
	}
}

// The primary record reuses one MathDocument with registered handlers. This
// separate Go lifetime check reuses the actual Compiler with equivalent authored
// definitions in its declared newcommand profile; the input mechanisms differ.
func TestMacroBoundaryCompilerReuse(t *testing.T) {
	c := NewCompiler()
	c.augmentedPackages = true
	for _, r := range macroReferences(t).ReuseCompiler {
		t.Run(r.Name, func(t *testing.T) {
			root, err := c.Compile(r.TeX, false)
			if err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(macroReferenceTree(root))
			if err != nil {
				t.Fatal(err)
			}
			var got, want any
			if err = json.Unmarshal(encoded, &got); err != nil {
				t.Fatal(err)
			}
			if err = json.Unmarshal(r.Primary.Tree, &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatal("reused compiler full tree differs")
			}
			rendered, err := svg.NewTypesetter().Typeset(root, pipeline.DefaultOptions())
			if err != nil {
				t.Fatal(err)
			}
			if fmt.Sprintf("%x", sha256.Sum256([]byte(rendered))) != r.Primary.SVGSHA256 {
				t.Fatal("reused compiler complete SVG differs")
			}
		})
	}
}

func macroReferenceBoundary(t *testing.T, c macroReferenceCase) {
	t.Helper()
	names := map[string]string{}
	for _, mode := range []string{"inline", "display"} {
		for _, name := range []string{"literal-hash", "zero-arg-hash"} {
			names[name+"-"+mode] = "D098 diagnostic quotation marks"
		}
		names["environment-control-"+mode] = "augmented-only environment supplemental observation"
		names["vector-control-"+mode] = "exact existing extra texClass:0 on vector accent"
		for _, name := range []string{"prime-macro", "direct-prime"} {
			names[name+"-"+mode] = "D060 exact inherited missing pseudoscript:false; direct accepted-before tree identical, raw SVG primary"
		}
	}
	reason, qualified := names[c.Name]
	if (c.Boundary != nil) != qualified {
		t.Fatal("named qualification set changed")
	}
	if !qualified {
		return
	}
	if c.Boundary.Reason != reason {
		t.Fatal("qualification reason changed")
	}
	vector := c.Name == "vector-control-inline" || c.Name == "vector-control-display"
	prime := c.Name == "prime-macro-inline" || c.Name == "prime-macro-display" || c.Name == "direct-prime-inline" || c.Name == "direct-prime-display"
	if !vector && !prime {
		if c.Boundary.SVGSHA256 == c.Primary.SVGSHA256 {
			t.Fatal("obsolete whole-output boundary")
		}
		return
	}
	if c.Boundary.SVGSHA256 != c.Primary.SVGSHA256 || c.Boundary.Error != nil || c.Primary.Error != nil {
		t.Fatal("metadata-only boundary changed paint/error")
	}
	var primary, bounded map[string]any
	if err := json.Unmarshal(c.Primary.Tree, &primary); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(c.Boundary.Tree, &bounded); err != nil {
		t.Fatal(err)
	}
	path := []int{0, 0, 1}
	if vector {
		path = []int{0, 0, 0, 0, 1}
	}
	at := primary
	for _, i := range path {
		at = at["children"].([]any)[i].(map[string]any)
	}
	expected := []any{map[string]any{"name": "variantForm", "value": true}, map[string]any{"name": "pseudoscript", "value": false}}
	replacement := []any{map[string]any{"name": "variantForm", "value": true}}
	if vector {
		expected = []any{map[string]any{"name": "mathaccent", "value": true}}
		replacement = []any{map[string]any{"name": "texClass", "value": float64(0)}, map[string]any{"name": "mathaccent", "value": true}}
	}
	if !reflect.DeepEqual(at["properties"], expected) {
		t.Fatal("primary precise property path changed")
	}
	at["properties"] = replacement
	if !reflect.DeepEqual(primary, bounded) {
		t.Fatal("boundary changed another full-tree field")
	}
}
