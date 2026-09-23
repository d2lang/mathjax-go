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
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
)

type tagTextTree struct {
	Kind       string           `json:"kind"`
	Text       *string          `json:"text"`
	Attributes []map[string]any `json:"attributes"`
	Properties []map[string]any `json:"properties"`
	Children   []*tagTextTree   `json:"children"`
}

func tagTextProjection(n *mml.Node) *tagTextTree {
	if n == nil {
		return nil
	}
	kind := n.Kind
	if kind == "mrow" && n.Flags.Inferred {
		kind = "inferredMrow"
	}
	r := &tagTextTree{Kind: kind, Attributes: macroReferencePairs(n.Attributes.Explicit()), Properties: macroReferencePairs(n.Properties), Children: []*tagTextTree{}}
	if n.Kind == "text" {
		x := n.Text
		r.Text = &x
	}
	for _, c := range n.Children {
		if c.Parent != n {
			panic("lost parent")
		}
		r.Children = append(r.Children, tagTextProjection(c))
	}
	return r
}
func tagTextDecoded(t *testing.T, n *mml.Node) *tagTextTree {
	t.Helper()
	b, e := json.Marshal(tagTextProjection(n))
	if e != nil {
		t.Fatal(e)
	}
	var r *tagTextTree
	if e = json.Unmarshal(b, &r); e != nil {
		t.Fatal(e)
	}
	return r
}
func tagTextRender(t *testing.T, n *mml.Node, display bool, want string) {
	t.Helper()
	o := pipeline.DefaultOptions()
	o.Display = display
	s, e := svg.NewTypesetter().Typeset(n, o)
	if e != nil {
		t.Fatal(e)
	}
	if fmt.Sprintf("%x", sha256.Sum256([]byte(s))) != want {
		t.Fatal("complete primary SVG differs")
	}
}
func tagTextFinalize(children []*mml.Node, display bool) *mml.Node {
	n := node("math", children...)
	if display {
		n.Attributes.Set("display", "block")
	}
	n.Walk(func(x *mml.Node) bool {
		for _, k := range []string{resolvedFontScope, ambientFontSource, vectorFactoryToken, vectorFactoryDone, limitsScriptOrigin} {
			x.RemoveProperty(k)
		}
		return true
	})
	setMathMLInheritance(n, display)
	n = moveMathLimits(n)
	cleanMathMLAttributes(n)
	return n
}
func TestTagTextCompletePrimary(t *testing.T) {
	var f struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVGSHA256 string
			Display              bool
			Tree                 *tagTextTree
			Error                *Error
			Registration         *struct {
				Name, Body string
				Arguments  int
			}
		}
	}
	b, e := os.ReadFile("testdata/ams_tag_text_mathjax_3_2_2.json")
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(b, &f); e != nil {
		t.Fatal(e)
	}
	if f.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(f.Cases) != 62 {
		t.Fatal("unbound complete primary records")
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			state := newParseState()
			if m := c.Registration; m != nil {
				state.macros[m.Name] = macroDefinition{body: m.Body, arguments: m.Arguments}
			}
			p := &parser{source: c.TeX, state: state, display: c.Display}
			children, stop, err := p.parseRow(0, false)
			if stop != "" {
				t.Fatal("unexpected row stop")
			}
			position := p.pos
			if err == nil {
				children, err = p.amsTagFinalize(children)
			}
			macroReferenceError(t, err, c.Error)
			if p.source != c.TeX || p.pos != position || p.state != state {
				t.Fatal("tag finalization changed outer source/cursor/configuration")
			}
			var root *mml.Node
			if err != nil {
				var pe *Error
				if !errors.As(err, &pe) {
					t.Fatal(err)
				}
				root = mathError(pe.Message, c.Display)
			} else {
				root = tagTextFinalize(children, c.Display)
			}
			if !reflect.DeepEqual(tagTextDecoded(t, root), c.Tree) {
				t.Fatal("complete ordered explicit/own tree differs")
			}
			tagTextRender(t, root, c.Display, c.SVGSHA256)
		})
	}
}

// These two bridges allow the accepted-production negative overlay to use its
// original no-parser signatures. Only call arity changes in that test overlay;
// every assertion and both exact accepted production files remain authoritative.
func tagTextMake(p *parser) (*mml.Node, error) { return p.amsTags().makeTag(p) }
func tagTextGet(p *parser) (*mml.Node, error)  { return p.amsTags().getTag(p) }

func TestTagTextActualMethods(t *testing.T) {
	type label struct{ Tag, ID string }
	type input struct {
		Name, Operation, Tag, Format, Font, Label, Environment, TeX string
		Sources                                                     []string
		MacroCount                                                  int
		PriorLabel, Taggable, DefaultTags, NoTag, Display           bool
		Registration                                                *struct {
			Name, Body string
			Arguments  int
		}
	}
	var f struct {
		MathjaxGitCommit string
		Cases            []struct {
			Input                  input
			FullTreePrimary        bool
			Tree                   *tagTextTree
			Error                  *Error
			ReturnIsInput, Parents bool
			ContentReused          *bool
			Outer                  struct {
				Source, Remaining, LateColor                                                                                           string
				Cursor, MacroCount                                                                                                     int
				ConfigurationIdentity, CurrentIdentity, LabelsIdentity, StackIdentity, HistoryIdentity, FactoryIdentity, ColorIdentity bool
			}
			After struct {
				Current struct {
					Env, TagId, TagFormat, LabelId string
					Tag                            *string
					Taggable, DefaultTags, NoTag   bool
				}
				Labels map[string]label
			}
			Provenance []struct {
				Path                                  []int
				Kind                                  string
				Text                                  *string
				Attributes, Primary, SourceProvenance []map[string]any
			}
			Stages []struct {
				TeX, SVGSHA256     string
				Tree               *tagTextTree
				Error              *Error
				CompleteSVGPrimary bool
			}
		}
	}
	b, e := os.ReadFile("testdata/ams_tag_method_mathjax_3_2_2.json")
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(b, &f); e != nil {
		t.Fatal(e)
	}
	if f.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(f.Cases) != 18 {
		t.Fatal("unbound actual-method records")
	}
	paths := map[string][][]int{"make-tag-tilde": {{0, 0}}, "formatted-empty": {{0, 0}}, "outer-bold-empty-env": {{0, 0, 0}, {0, 0, 1, 0, 0}, {0, 0, 2}}, "inner-macro-success": {{0, 0, 0}, {0, 0, 1, 0, 0}, {0, 0, 2}}, "tag-state-identity": {{0, 0}}, "finalize-inline": {{}}, "finalize-display": {{0, 0, 0, 0}, {0, 1, 0, 0}}}
	paths["text-override-one-argument"] = [][]int{{0, 0}, {0, 1}, {0, 2}, {0, 3}, {0, 4}, {0, 5}, {0, 6}}
	paths["text-override-zero-argument"] = [][]int{{0, 0}, {0, 1, 0, 0}, {0, 1, 0, 1}, {0, 1, 0, 2}, {0, 1, 0, 3}, {0, 1, 0, 4}}
	markerNodes := 0
	for _, c := range f.Cases {
		t.Run(c.Input.Name, func(t *testing.T) {
			in := c.Input
			if in.Operation == "compile" || in.Operation == "sequence" {
				compiler := NewCompiler()
				if len(c.Stages) != len(in.Sources) && in.Operation == "sequence" {
					t.Fatal("sequence length")
				}
				for _, s := range c.Stages {
					state := newParseState()
					p := &parser{source: s.TeX, state: state, display: in.Display}
					children, stop, err := p.parseRow(0, false)
					if stop != "" {
						t.Fatal("unexpected row stop")
					}
					if err == nil {
						_, err = p.amsTagFinalize(children)
					}
					macroReferenceError(t, err, s.Error)
					if in.Name == "equation-error-teardown" {
						tags := p.amsTags()
						if tags.current.environment != "" || len(tags.stack) != 0 || len(tags.history) != 1 || tags.labels["L"].id != "mjx-eqn:L" {
							t.Fatal("existing Go deferred end/early label write changed")
						}
					}
					n, err := compiler.Compile(s.TeX, in.Display)
					if err != nil {
						t.Fatal(err)
					}
					fresh, err := NewCompiler().Compile(s.TeX, in.Display)
					if err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(tagTextDecoded(t, n), s.Tree) || !reflect.DeepEqual(tagTextDecoded(t, n), tagTextDecoded(t, fresh)) {
						t.Fatal("complete primary tree or compiler lifetime differs")
					}
					if s.CompleteSVGPrimary {
						tagTextRender(t, n, in.Display, s.SVGSHA256)
					} else if in.Name != "alignment-row-order" {
						t.Fatal("undeclared complete-SVG exclusion")
					}
				}
				return
			}
			state := newParseState()
			state.macroCount = in.MacroCount
			if m := in.Registration; m != nil {
				state.macros[m.Name] = macroDefinition{body: m.Body, arguments: m.Arguments}
			}
			p := &parser{source: "outer remaining", pos: 2, state: state, activeFont: in.Font, display: in.Display}
			tags := p.amsTags()
			tags.current = &amsTagInfo{environment: in.Environment, taggable: in.Taggable, defaultTags: in.DefaultTags, tag: stringPointer(in.Tag), tagFormat: in.Format, noTag: in.NoTag, labelID: in.Label}
			if in.PriorLabel {
				tags.labels["previous"] = amsLabel{tag: "P", id: "prior"}
			}
			current, model, labelMap := tags.current, state.colorModel, reflect.ValueOf(tags.labels).Pointer()
			var value, content *mml.Node
			var err error
			switch in.Operation {
			case "makeTag":
				value, err = tagTextMake(p)
			case "getTag":
				value, err = tagTextGet(p)
			case "finalize":
				content = token("mi", "x")
				var nodes []*mml.Node
				nodes, err = p.amsTagFinalize([]*mml.Node{content})
				if len(nodes) == 1 {
					value = nodes[0]
				} else if err == nil {
					t.Fatal("finalize cardinality")
				}
			default:
				t.Fatal("unknown actual method")
			}
			macroReferenceError(t, err, c.Error)
			if p.source != c.Outer.Source || p.pos != c.Outer.Cursor || p.source[p.pos:] != c.Outer.Remaining || p.activeFont != in.Font || state.macroCount != c.Outer.MacroCount || state.macroCount != in.MacroCount || p.state != state || state.amsTags != tags || tags.current != current || state.colorModel != model || reflect.ValueOf(tags.labels).Pointer() != labelMap || len(tags.stack) != 0 || len(tags.history) != 0 {
				t.Fatal("outer lexical/counter/shared identity changed")
			}
			if !c.Outer.ConfigurationIdentity || !c.Outer.CurrentIdentity || !c.Outer.LabelsIdentity || !c.Outer.StackIdentity || !c.Outer.HistoryIdentity || !c.Outer.FactoryIdentity || !c.Outer.ColorIdentity {
				t.Fatal("primary identity contract changed")
			}
			late, err := model.GetColor("named", "late")
			if err != nil || late != c.Outer.LateColor {
				t.Fatal("shared color side effect differs")
			}
			q := c.After.Current
			if current.environment != q.Env || current.tagID != q.TagId || current.tagFormat != q.TagFormat || current.labelID != q.LabelId || !reflect.DeepEqual(current.tag, q.Tag) || current.taggable != q.Taggable || current.defaultTags != q.DefaultTags || current.noTag != q.NoTag {
				t.Fatal("live current tag differs")
			}
			labels := map[string]label{}
			for k, v := range tags.labels {
				labels[k] = label{v.tag, v.id}
			}
			if !reflect.DeepEqual(labels, c.After.Labels) {
				t.Fatal("ordered early ID/label writes differ")
			}
			if m := in.Registration; m != nil {
				v := state.macros[m.Name]
				if v.body != m.Body || v.arguments != m.Arguments {
					t.Fatal("shared registration mutated")
				}
			}
			if c.Error != nil && value != nil {
				t.Fatal("failed content returned outer cell")
			}
			seen := map[*mml.Node]bool{}
			if value != nil {
				value.Walk(func(n *mml.Node) bool {
					if seen[n] {
						t.Fatal("duplicate node identity")
					}
					seen[n] = true
					for _, ch := range n.Children {
						if ch.Parent != n {
							t.Fatal("child ownership lost")
						}
					}
					return true
				})
			}
			if !c.Parents || (value == content) != c.ReturnIsInput {
				t.Fatal("returned identity differs")
			}
			if c.ContentReused != nil && seen[content] != *c.ContentReused {
				t.Fatal("original content identity lost")
			}
			if !c.FullTreePrimary {
				t.Fatal("unexpected private full-tree exclusion")
			}
			got := tagTextDecoded(t, value)
			want := c.Tree
			wantPaths := paths[in.Name]
			if len(c.Provenance) != len(wantPaths) {
				t.Fatal("finite provenance inventory changed")
			}
			for i, mark := range c.Provenance {
				if !reflect.DeepEqual(mark.Path, wantPaths[i]) {
					t.Fatal("provenance path changed")
				}
				g, w := got, want
				for _, j := range mark.Path {
					if j >= len(g.Children) || j >= len(w.Children) {
						t.Fatal("missing provenance node")
					}
					g, w = g.Children[j], w.Children[j]
				}
				if g.Kind != mark.Kind || w.Kind != mark.Kind || !reflect.DeepEqual(g.Text, mark.Text) || !reflect.DeepEqual(w.Text, mark.Text) || !reflect.DeepEqual(g.Attributes, mark.Attributes) || !reflect.DeepEqual(w.Attributes, mark.Attributes) || !reflect.DeepEqual(w.Properties, mark.Primary) || !reflect.DeepEqual(g.Properties, mark.SourceProvenance) {
					t.Fatal("exact source provenance pair differs")
				}
				keep := []map[string]any{}
				for _, v := range mark.SourceProvenance {
					switch v["name"] {
					case resolvedFontScope, ambientFontSource, vectorFactoryToken:
						if v["value"] != true {
							t.Fatal("invalid provenance value")
						}
					default:
						keep = append(keep, v)
					}
				}
				if !reflect.DeepEqual(keep, mark.Primary) {
					t.Fatal("provenance changes primary own fields")
				}
				w.Properties = mark.SourceProvenance
				markerNodes++
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatal("complete ordered method tree differs outside exact provenance paths")
			}
		})
	}
	if markerNodes != 25 {
		t.Fatal("finite provenance node coverage changed", markerNodes)
	}
}
