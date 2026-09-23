// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
)

type mathTildeInput struct {
	Name, TeX    string
	Display      bool
	Registration *struct {
		Name, Body string
		Arguments  int
	}
}
type mathTildeField struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}
type mathTildeRawTree struct {
	Kind       string  `json:"kind"`
	Text       *string `json:"text"`
	Attributes struct {
		Explicit []mathTildeField `json:"explicit"`
	} `json:"attributes"`
	Properties []mathTildeField    `json:"properties"`
	Children   []*mathTildeRawTree `json:"children"`
}
type mathTildeTree struct {
	Kind                   string
	Text                   *string
	Attributes, Properties []mathTildeField
	Children               []*mathTildeTree
}
type mathTildeState struct {
	Source, Remaining, Font, MultiLetterFont, VectorFont string
	ByteCursor, CursorUTF16, MacroCount                  int
	VectorFactory, VectorStar                            bool
}
type mathTildeRecord struct {
	Input                 mathTildeInput
	Tree                  *mathTildeRawTree
	SVG                   string
	Error, FormattedError *Error
	Parser                mathTildeState
	Events                []struct{ Before, After mathTildeState }
}
type mathTildeStored struct {
	Name, Scope, SHA256, Raw string
	Input                    mathTildeInput
}
type mathTildeFixture struct {
	CaptureFreeze, MathjaxGitCommit, AcceptedMerge string
	Records                                        []mathTildeStored
}

func mathTildeFixtures(t *testing.T) (mathTildeFixture, map[string]mathTildeStored) {
	t.Helper()
	load := func(path string) mathTildeFixture {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var f mathTildeFixture
		if err = json.Unmarshal(b, &f); err != nil {
			t.Fatal(err)
		}
		if f.CaptureFreeze != "e1fae210aca9cbf022f24d85618c78a7cb9488e2d1635db9a250c72761c2d2c3" {
			t.Fatal("unbound capture freeze")
		}
		if f.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || f.AcceptedMerge != "87f0f6283d0ce6bd481c688aaf1b0980a9974156" || len(f.Records) != 48 {
			t.Fatal("unbound corpus")
		}
		seen := map[string]bool{}
		public, private := 0, 0
		for _, r := range f.Records {
			key := r.Scope + "/" + r.Name
			if seen[key] || r.Name != r.Input.Name || fmt.Sprintf("%x", sha256.Sum256([]byte(r.Raw))) != r.SHA256 {
				t.Fatal("record identity/hash", key)
			}
			seen[key] = true
			switch r.Scope {
			case "public":
				public++
			case "private":
				private++
			default:
				t.Fatal("scope", r.Scope)
			}
		}
		if public != 40 || private != 8 {
			t.Fatal("scope inventory")
		}
		return f
	}
	p := load("testdata/math_tilde_mathjax_3_2_2.json")
	a := load("testdata/math_tilde_boundaries.json")
	am := map[string]mathTildeStored{}
	for _, r := range a.Records {
		am[r.Scope+"/"+r.Name] = r
	}
	for _, r := range p.Records {
		a, ok := am[r.Scope+"/"+r.Name]
		if !ok || !reflect.DeepEqual(a.Input, r.Input) {
			t.Fatal("accepted input binding", r.Name)
		}
	}
	return p, am
}
func mathTildeDecode(t *testing.T, s mathTildeStored) mathTildeRecord {
	t.Helper()
	var r mathTildeRecord
	if err := json.Unmarshal([]byte(s.Raw), &r); err != nil {
		t.Fatal(err)
	}
	// Original ordinary captures predate the name/registration envelope; exact TeX/mode remain bound.
	if r.Input.TeX != s.Input.TeX || r.Input.Display != s.Input.Display {
		t.Fatal("raw input binding", s.Name)
	}
	return r
}
func mathTildePrimaryTree(n *mathTildeRawTree) *mathTildeTree {
	if n == nil {
		return nil
	}
	r := &mathTildeTree{Kind: n.Kind, Text: n.Text, Attributes: n.Attributes.Explicit, Properties: n.Properties, Children: []*mathTildeTree{}}
	for _, c := range n.Children {
		r.Children = append(r.Children, mathTildePrimaryTree(c))
	}
	return r
}
func mathTildeNodeTree(n *mml.Node) *mathTildeTree {
	r := &mathTildeTree{Kind: n.Kind, Attributes: []mathTildeField{}, Properties: []mathTildeField{}, Children: []*mathTildeTree{}}
	if n.Kind == "mrow" && n.Flags.Inferred {
		r.Kind = "inferredMrow"
	}
	if n.Kind == "text" {
		s := n.Text
		r.Text = &s
	}
	for _, k := range n.Attributes.ExplicitNames() {
		v, _ := n.Attributes.GetExplicit(k)
		r.Attributes = append(r.Attributes, mathTildeField{k, v})
	}
	for _, k := range n.Properties.Keys() {
		v, _ := n.Property(k)
		r.Properties = append(r.Properties, mathTildeField{k, v})
	}
	for _, c := range n.Children {
		if c.Parent != n {
			panic("invalid parent pointer")
		}
		r.Children = append(r.Children, mathTildeNodeTree(c))
	}
	return r
}
func mathTildeCanonical(n *mathTildeTree) *mathTildeTree {
	// JSON's numeric type is the fixture wire type; this changes no field/value/order.
	b, err := json.Marshal(n)
	if err != nil {
		panic(err)
	}
	var r *mathTildeTree
	if err = json.Unmarshal(b, &r); err != nil {
		panic(err)
	}
	return r
}
func mathTildeNodeAt(n *mathTildeTree, path []int) *mathTildeTree {
	for _, i := range path {
		if n == nil || i < 0 || i >= len(n.Children) {
			return nil
		}
		n = n.Children[i]
	}
	return n
}
func mathTildePrimeName(c mathTildeStored) bool {
	return c.Input.TeX == "x'~_iy" && c.Input.Registration == nil && ((c.Scope == "public" && ((c.Name == "prime-inline" && !c.Input.Display) || (c.Name == "prime-display" && c.Input.Display))) || (c.Scope == "private" && c.Name == "real-prime-recipient" && !c.Input.Display))
}
func mathTildeCompare(c mathTildeStored, primary, actual *mathTildeTree) error {
	if !mathTildePrimeName(c) {
		if !reflect.DeepEqual(primary, actual) {
			return fmt.Errorf("complete primary tree differs")
		}
		return nil
	}
	path := []int{0, 0, 1}
	p, a := mathTildeNodeAt(primary, path), mathTildeNodeAt(actual, path)
	wantP := []mathTildeField{{"variantForm", true}, {"pseudoscript", false}}
	wantA := []mathTildeField{{"variantForm", true}}
	if p == nil || a == nil || p.Kind != "mo" || a.Kind != "mo" || len(p.Children) != 1 || len(a.Children) != 1 || p.Children[0].Text == nil || *p.Children[0].Text != "′" || a.Children[0].Text == nil || *a.Children[0].Text != "′" || !reflect.DeepEqual(p.Properties, wantP) || !reflect.DeepEqual(a.Properties, wantA) {
		return fmt.Errorf("named prime property pair changed")
	}
	visited := 0
	var compare func(*mathTildeTree, *mathTildeTree, []int) bool
	compare = func(p, a *mathTildeTree, at []int) bool {
		if p == nil || a == nil {
			return p == a
		}
		if p.Kind != a.Kind || !reflect.DeepEqual(p.Text, a.Text) || !reflect.DeepEqual(p.Attributes, a.Attributes) || len(p.Children) != len(a.Children) {
			return false
		}
		if reflect.DeepEqual(at, path) {
			visited++
		} else if !reflect.DeepEqual(p.Properties, a.Properties) {
			return false
		}
		for i := range p.Children {
			if !compare(p.Children[i], a.Children[i], append(append([]int{}, at...), i)) {
				return false
			}
		}
		return true
	}
	if !compare(primary, actual, []int{}) || visited != 1 {
		return fmt.Errorf("difference outside the one named prime property array")
	}
	return nil
}
func mathTildeParserState(p *parser) mathTildeState {
	return mathTildeState{Source: p.source, Remaining: p.source[p.pos:], Font: p.activeFont, MultiLetterFont: p.multiLetterFont, VectorFont: p.vectorFont, ByteCursor: p.pos, CursorUTF16: len(utf16.Encode([]rune(p.source[:p.pos]))), MacroCount: p.state.macroCount, VectorFactory: p.vectorFactory, VectorStar: p.vectorStar}
}
func mathTildeCompile(c mathTildeInput) (*mml.Node, *parser, *Error, error) {
	s := newParseState()
	if r := c.Registration; r != nil {
		s.macros[r.Name] = macroDefinition{body: r.Body, arguments: r.Arguments}
	}
	p := &parser{source: c.TeX, state: s, display: c.Display}
	children, stop, err := p.parseRow(0, false)
	if err == nil && stop != "" {
		return nil, p, nil, fmt.Errorf("unexpected stop %q", stop)
	}
	if err == nil {
		children, err = p.amsTagFinalize(children)
	}
	if err != nil {
		var pe *Error
		if !errors.As(err, &pe) {
			return nil, p, nil, err
		}
		return mathError(pe.Message, c.Display), p, pe, nil
	}
	r := node("math", children...)
	if c.Display {
		r.Attributes.Set("display", "block")
	}
	r.Walk(func(n *mml.Node) bool {
		for _, k := range []string{resolvedFontScope, ambientFontSource, vectorFactoryToken, vectorFactoryDone, limitsScriptOrigin} {
			n.RemoveProperty(k)
		}
		return true
	})
	setMathMLInheritance(r, c.Display)
	r = moveMathLimits(r)
	cleanMathMLAttributes(r)
	return r, p, nil, nil
}

func TestMathTildePinnedOutputs(t *testing.T) {
	f, accepted := mathTildeFixtures(t)
	for _, c := range f.Records {
		t.Run(c.Scope+"/"+c.Name, func(t *testing.T) {
			primary := mathTildeDecode(t, c)
			old := mathTildeDecode(t, accepted[c.Scope+"/"+c.Name])
			root, p, pe, err := mathTildeCompile(c.Input)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(mathTildeParserState(p), old.Parser) {
				t.Fatal("accepted source/cursor/counter/font state changed")
			}
			got := mathTildeCanonical(mathTildeNodeTree(root))
			want := mathTildePrimaryTree(primary.Tree)
			isAccent := c.Scope == "public" && (c.Name == "accent-inline" || c.Name == "accent-display")
			isEscaped := c.Scope == "public" && (c.Name == "escaped-inline" || c.Name == "escaped-display")
			if isAccent || isEscaped {
				tex := `\tilde{x}`
				if isEscaped {
					tex = `\~`
				}
				if c.Input.TeX != tex || c.Input.Display != strings.HasSuffix(c.Name, "-display") || c.Input.Registration != nil {
					t.Fatal("boundary input")
				}
				if !reflect.DeepEqual(got, mathTildePrimaryTree(old.Tree)) || !reflect.DeepEqual(pe, old.Error) {
					t.Fatal("whole accepted control changed")
				}
				if isEscaped {
					if primary.FormattedError == nil || primary.FormattedError.ID != "UndefinedControlSequence" || primary.FormattedError.Message != `Undefined control sequence \~` || old.Error != nil {
						t.Fatal("escaped-command diagnostic authority")
					}
				} else {
					pn, an := mathTildeNodeAt(want, []int{0, 0, 0, 0, 1}), mathTildeNodeAt(got, []int{0, 0, 0, 0, 1})
					if pn == nil || an == nil || !reflect.DeepEqual(pn.Properties, []mathTildeField{{"mathaccent", true}}) || !reflect.DeepEqual(an.Properties, []mathTildeField{{"texClass", float64(0)}, {"mathaccent", true}}) {
						t.Fatal("accent property authority")
					}
				}
			} else {
				if !reflect.DeepEqual(pe, primary.FormattedError) {
					t.Fatal("structured error differs")
				}
				if err = mathTildeCompare(c, want, got); err != nil {
					t.Fatal(err)
				}
			}
			opts := pipeline.DefaultOptions()
			opts.Display = c.Input.Display
			s, err := svg.NewTypesetter().Typeset(root, opts)
			if err != nil {
				t.Fatal(err)
			}
			expected := primary.SVG
			if isEscaped {
				expected = old.SVG
			}
			if s != expected {
				t.Fatalf("whole SVG differs: got %x want %x", sha256.Sum256([]byte(s)), sha256.Sum256([]byte(expected)))
			}
		})
	}
}

func TestMathTildeActualCharacterState(t *testing.T) {
	f, accepted := mathTildeFixtures(t)
	for _, c := range f.Records {
		if c.Scope != "private" {
			continue
		}
		t.Run(c.Name, func(t *testing.T) {
			old := mathTildeDecode(t, accepted[c.Scope+"/"+c.Name])
			if len(old.Events) != 1 {
				t.Fatal("actual event inventory")
			}
			b, a := old.Events[0].Before, old.Events[0].After
			state := newParseState()
			state.macroCount = b.MacroCount
			p := &parser{source: b.Source, pos: b.ByteCursor, state: state, display: c.Input.Display, activeFont: b.Font, multiLetterFont: b.MultiLetterFont, vectorFactory: b.VectorFactory, vectorFont: b.VectorFont, vectorStar: b.VectorStar}
			before := *p
			n := p.parseCharacter()
			if p.state != state || p.source != before.source || p.pos != before.pos+1 || !reflect.DeepEqual(mathTildeParserState(p), a) {
				t.Fatal("character source/cursor/state changed")
			}
			if n.Kind != "mtext" || len(n.Children) != 1 || n.Children[0].Kind != "text" || n.Children[0].Text != "\u00a0" || n.Children[0].Parent != n || len(n.Attributes.ExplicitNames()) != 0 {
				t.Fatal("actual token shape/ownership")
			}
			if !reflect.DeepEqual(n.Properties.Keys(), []string{vectorFactoryToken}) {
				t.Fatal("token factory provenance or unrelated marker")
			}
			if origin, _ := n.Property(vectorFactoryToken); origin != true {
				t.Fatal("token factory bypass")
			}
			text := n.Children[0]
			p.applyVectorFactory(n)
			if n.Children[0] != text || text.Parent != n || len(n.Attributes.ExplicitNames()) != 0 {
				t.Fatal("vector factory changed NBSP/identity")
			}
			if b.VectorFactory {
				if done, _ := n.Property(vectorFactoryDone); done != true {
					t.Fatal("active token factory did not visit NBSP")
				}
			}
		})
	}
}

func TestMathTildeRealRecipients(t *testing.T) {
	f, _ := mathTildeFixtures(t)
	for _, c := range f.Records {
		if c.Scope != "private" && c.Name != "repeated-inline" && c.Name != "repeated-display" {
			continue
		}
		t.Run(c.Scope+"/"+c.Name, func(t *testing.T) {
			s := newParseState()
			p := &parser{source: c.Input.TeX, state: s, display: c.Input.Display}
			nodes, stop, err := p.parseRow(0, false)
			if err != nil || stop != "" {
				t.Fatal(err, stop)
			}
			root := node("math", nodes...)
			var spaces []*mml.Node
			root.Walk(func(n *mml.Node) bool {
				for _, child := range n.Children {
					if child.Parent != n {
						t.Fatal("parent pointer")
					}
				}
				if n.Kind == "mtext" && textContent(n) == "\u00a0" {
					spaces = append(spaces, n)
				}
				return true
			})
			count := 1
			if strings.HasPrefix(c.Name, "repeated-") {
				count = 2
			}
			if len(spaces) != count {
				t.Fatalf("got %d distinct NBSP recipients, want %d", len(spaces), count)
			}
			for i, n := range spaces {
				origin, _ := n.Property(vectorFactoryToken)
				if origin != true || len(n.Children) != 1 || n.Children[0].Parent != n {
					t.Fatal("real recipient factory/text ownership")
				}
				for _, m := range spaces[:i] {
					if n == m || n.Children[0] == m.Children[0] {
						t.Fatal("space tokens/text children were coalesced")
					}
				}
			}
			if c.Name == "real-script-recipient" && (spaces[0].Parent.Kind != "msup" || spaces[0].Parent.Children[1] != spaces[0]) {
				t.Fatal("superscript recipient")
			}
			if c.Name == "real-prime-recipient" && (spaces[0].Parent.Kind != "msub" || spaces[0].Parent.Children[0] != spaces[0]) {
				t.Fatal("prime-following subscript base")
			}
		})
	}
}

func TestMathTildePrimeBoundaryRejectsDrift(t *testing.T) {
	f, _ := mathTildeFixtures(t)
	var c mathTildeStored
	for _, r := range f.Records {
		if r.Name == "prime-inline" {
			c = r
		}
	}
	primary := mathTildePrimaryTree(mathTildeDecode(t, c).Tree)
	actual := mathTildeCanonical(primary)
	prime := mathTildeNodeAt(actual, []int{0, 0, 1})
	prime.Properties = []mathTildeField{{"variantForm", true}}
	if err := mathTildeCompare(c, primary, actual); err != nil {
		t.Fatal("source-bound exact pair", err)
	}
	for _, name := range []string{"wrong-name", "wrong-tex", "wrong-mode", "extra-property", "missing-property", "reordered-primary", "unrelated-node", "wrong-path", "prime-text", "primary-false-changed"} {
		t.Run(name, func(t *testing.T) {
			cc := c
			pp, aa := mathTildeCanonical(primary), mathTildeCanonical(actual)
			p, a := mathTildeNodeAt(pp, []int{0, 0, 1}), mathTildeNodeAt(aa, []int{0, 0, 1})
			switch name {
			case "wrong-name":
				cc.Name = "another-prime-inline"
			case "wrong-tex":
				cc.Input.TeX = "x'~_iz"
			case "wrong-mode":
				cc.Input.Display = true
			case "extra-property":
				a.Properties = append(a.Properties, mathTildeField{"other", false})
			case "missing-property":
				a.Properties = []mathTildeField{}
			case "reordered-primary":
				p.Properties[0], p.Properties[1] = p.Properties[1], p.Properties[0]
			case "unrelated-node":
				aa.Attributes = append(aa.Attributes, mathTildeField{"mathcolor", "red"})
			case "wrong-path":
				aa.Children[0].Children[0].Children[0], aa.Children[0].Children[0].Children[1] = aa.Children[0].Children[0].Children[1], aa.Children[0].Children[0].Children[0]
			case "prime-text":
				v := "″"
				a.Children[0].Text = &v
			case "primary-false-changed":
				p.Properties[1].Value = true
			}
			if mathTildeCompare(cc, pp, aa) == nil {
				t.Fatal("invalid qualification accepted")
			}
		})
	}
}
