// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
)

type numberPair struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}
type numberTree struct {
	Kind       string        `json:"kind"`
	Text       *string       `json:"text"`
	Attributes []numberPair  `json:"attributes"`
	Properties []numberPair  `json:"properties"`
	Children   []*numberTree `json:"children"`
}

func numberProjection(n *mml.Node) *numberTree {
	r := &numberTree{Kind: n.Kind, Attributes: []numberPair{}, Properties: []numberPair{}, Children: []*numberTree{}}
	if n.Kind == "mrow" && n.Flags.Inferred {
		r.Kind = "inferredMrow"
	}
	if n.Kind == "text" {
		s := n.Text
		r.Text = &s
	}
	for _, k := range n.Attributes.ExplicitNames() {
		v, _ := n.Attributes.GetExplicit(k)
		r.Attributes = append(r.Attributes, numberPair{k, v})
	}
	n.Properties.Range(func(k string, v any) bool { r.Properties = append(r.Properties, numberPair{k, v}); return true })
	for _, c := range n.Children {
		r.Children = append(r.Children, numberProjection(c))
	}
	return r
}
func numberAt(n *numberTree, path []int) *numberTree {
	for _, i := range path {
		n = n.Children[i]
	}
	return n
}
func numberDecode(t *testing.T, n *mml.Node) *numberTree {
	t.Helper()
	b, e := json.Marshal(numberProjection(n))
	if e != nil {
		t.Fatal(e)
	}
	var r *numberTree
	if e = json.Unmarshal(b, &r); e != nil {
		t.Fatal(e)
	}
	return r
}

func TestOrdinaryNumberScannerReferences(t *testing.T) {
	var f struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVGSHA256 string
			Display              bool
			Tree                 *numberTree
			Thrown               *struct{ Type, Message string }
			Registration         *struct {
				Name, Body string
				Arguments  int
			}
		}
	}
	readArgumentJSON(t, "../../testdata/ordinary_number_mathjax_3_2_2.json", &f)
	var b struct {
		AcceptedBase string
		Cases        map[string]struct {
			Status, BaselineSVG string
			BaselineTree        *numberTree
			PrimaryThrown       *struct{ Type, Message string }
			Fields              []struct {
				Path, BaselinePath        []int
				PrimaryNode, BaselineNode *numberTree
			}
		}
	}
	readArgumentJSON(t, "../../testdata/ordinary_number_boundaries.json", &b)
	if f.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(f.Cases) != 95 || len(b.Cases) != 7 || b.AcceptedBase != "bca7909b2de5fc00d0057b24092c31ebd15ac7b6" {
		t.Fatal("unbound scanner references")
	}
	counts := map[string]int{}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var root *mml.Node
			var err error
			if c.Registration == nil {
				root, err = NewCompiler().Compile(c.TeX, c.Display)
			} else {
				// Same compiler phases, with this one registered macro; no D089 source
				// reset/join behavior is changed or presented as primary-identical.
				if c.Name != "private-registered-number-tail" || c.Registration.Name != "numtail" || c.Registration.Body != "12{,}345.6" || c.Registration.Arguments != 0 {
					t.Fatal("unexpected registration")
				}
				state := newParseState()
				state.macros[c.Registration.Name] = macroDefinition{body: c.Registration.Body, arguments: 0}
				p := &parser{source: c.TeX, state: state, display: c.Display}
				children, stop, e := p.parseRow(0, false)
				if e != nil || stop != "" || state.macroCount != 1 {
					t.Fatal("registered parser changed", e, stop)
				}
				children, e = p.amsTagFinalize(children)
				if e != nil {
					t.Fatal(e)
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
			if err != nil {
				t.Fatal(err)
			}
			root.Walk(func(n *mml.Node) bool {
				for _, ch := range n.Children {
					if ch.Parent != n {
						t.Fatal("lost child ownership")
					}
				}
				return true
			})
			want, hash := c.Tree, c.SVGSHA256
			if q, ok := b.Cases[c.Name]; ok {
				counts[q.Status]++
				switch q.Status {
				case "unsupported-primary-unchanged-Go":
					if c.Name != "public-unicode-segmented-inline" && c.Name != "public-unicode-segmented-display" && c.Name != "private-other-segmented" {
						t.Fatal("unexpected unsupported input")
					}
					if c.Tree != nil || c.SVGSHA256 != "" || c.Thrown == nil || c.Thrown.Type != "TypeError" || c.Thrown.Message != "Cannot read properties of null (reading '4')" || !reflect.DeepEqual(q.PrimaryThrown, c.Thrown) {
						t.Fatal("invented successful primary result")
					}
					want, hash = q.BaselineTree, q.BaselineSVG
				case "exact-inherited-token-properties":
					if len(q.Fields) != 1 || c.Thrown != nil {
						t.Fatal("unbound metadata scope")
					}
					d := q.Fields[0]
					at := numberAt(want, d.Path)
					if !reflect.DeepEqual(at, d.PrimaryNode) || d.PrimaryNode.Kind != d.BaselineNode.Kind || !reflect.DeepEqual(d.PrimaryNode.Text, d.BaselineNode.Text) || !reflect.DeepEqual(d.PrimaryNode.Attributes, d.BaselineNode.Attributes) || !reflect.DeepEqual(d.PrimaryNode.Children, d.BaselineNode.Children) {
						t.Fatal("boundary extends beyond exact properties")
					}
					switch c.Name {
					case "public-physics-vector-inline", "public-physics-vector-display":
						if !reflect.DeepEqual(d.PrimaryNode.Properties, []numberPair{{"mathaccent", true}}) || !reflect.DeepEqual(d.BaselineNode.Properties, []numberPair{{"texClass", float64(0)}, {"mathaccent", true}}) {
							t.Fatal("arrow metadata changed")
						}
					case "public-prime-after-inline", "public-prime-after-display":
						if !reflect.DeepEqual(d.PrimaryNode.Properties, []numberPair{{"variantForm", true}, {"pseudoscript", false}}) || !reflect.DeepEqual(d.BaselineNode.Properties, []numberPair{{"variantForm", true}}) {
							t.Fatal("prime metadata changed")
						}
					default:
						t.Fatal("unexpected metadata input")
					}
					at.Properties = d.BaselineNode.Properties
				default:
					t.Fatal("unknown boundary")
				}
			} else {
				counts["raw-primary"]++
				if c.Thrown != nil {
					t.Fatal("unqualified missing primary output")
				}
			}
			if got := numberDecode(t, root); !reflect.DeepEqual(got, want) {
				t.Fatalf("complete ordered explicit/own tree differs: %#v", got)
			}
			opts := pipeline.DefaultOptions()
			opts.Display = c.Display
			s, e := svg.NewTypesetter().Typeset(root, opts)
			if e != nil {
				t.Fatal(e)
			}
			if actual := fmt.Sprintf("%x", sha256.Sum256([]byte(s))); actual != hash {
				t.Fatalf("whole SVG %s want %s", actual, hash)
			}
		})
	}
	if !reflect.DeepEqual(counts, map[string]int{"raw-primary": 88, "exact-inherited-token-properties": 4, "unsupported-primary-unchanged-Go": 3}) {
		t.Fatal("changed classification", counts)
	}
}

func TestOrdinaryNumberRegisteredHandlers(t *testing.T) {
	type state struct {
		Source, Remaining                          string
		CursorUTF16, MacroCount                    int
		Environment                                map[string]any
		EnvironmentIdentity, ConfigurationIdentity int
	}
	var f struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, Character, Handler     string
			Before, After                state
			CreatedToken, PushedToken    *numberTree
			SameCreatedAndPushedIdentity bool
		}
	}
	readArgumentJSON(t, "testdata/ordinary_number_methods_mathjax_3_2_2.json", &f)
	if f.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(f.Cases) != 12 {
		t.Fatal("unbound actual methods")
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			if !c.SameCreatedAndPushedIdentity || c.Before.Source != c.After.Source || c.Before.EnvironmentIdentity != c.After.EnvironmentIdentity || c.Before.ConfigurationIdentity != c.After.ConfigurationIdentity || !reflect.DeepEqual(c.Before.Environment, c.After.Environment) || c.Before.MacroCount != c.After.MacroCount {
				t.Fatal("primary identity/state changed")
			}
			font, _ := c.Before.Environment["font"].(string)
			startUnits := c.Before.CursorUTF16 - len(utf16.Encode([]rune(c.Character)))
			bytePos := 0
			units := 0
			for units < startUnits {
				r, size := utf8.DecodeRuneInString(c.Before.Source[bytePos:])
				bytePos += size
				units += len(utf16.Encode([]rune{r}))
			}
			if units != startUnits {
				t.Fatal("invalid source cursor")
			}
			ps := newParseState()
			ps.macroCount = c.Before.MacroCount
			p := &parser{source: c.Before.Source, pos: bytePos, state: ps, activeFont: font}
			n := p.parseCharacter()
			if n.Parent != nil || p.state != ps || p.activeFont != font || p.source != c.After.Source || p.source[p.pos:] != c.After.Remaining || len(utf16.Encode([]rune(p.source[:p.pos]))) != c.After.CursorUTF16 || ps.macroCount != c.After.MacroCount {
				t.Fatal("token/source/cursor/state changed")
			}
			// Go lowers ambient font in appendNodes. Compare its corresponding Push
			// boundary; node() alone is not primary's complete token factory.
			if font != "" {
				applyScopedMathVariant(n, font)
			}
			p.applyVectorFactory(n)
			got := numberDecode(t, n)
			if len(got.Children) != 1 || n.Children[0].Parent != n {
				t.Fatal("token child identity lost")
			}
			expected := []numberPair{{vectorFactoryToken, true}, {ambientFontSource, true}}
			if c.Name == "other-osmanya-bold" {
				expected = []numberPair{{vectorFactoryToken, true}}
			} else if font != "" {
				expected = append(expected, numberPair{resolvedFontScope, true})
			}
			if !reflect.DeepEqual(got.Properties, expected) {
				t.Fatal("unexpected complete local provenance", got.Properties)
			}
			got.Properties = []numberPair{}
			if !reflect.DeepEqual(got, c.PushedToken) {
				t.Fatal("actual primary Push token differs")
			}
			if c.Name == "other-osmanya-bold" {
				if c.CreatedToken.Kind != "mi" || !reflect.DeepEqual(c.CreatedToken.Attributes, []numberPair{{"mathvariant", "bold"}}) || !reflect.DeepEqual(c.PushedToken.Attributes, []numberPair{{"mathvariant", "normal"}}) {
					t.Fatal("same-node range override ordering lost")
				}
			}
		})
	}
}
