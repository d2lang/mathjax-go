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

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
)

type autoOpenField struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}
type autoOpenTree struct {
	Kind       string          `json:"kind"`
	Text       *string         `json:"text"`
	Attributes []autoOpenField `json:"attributes"`
	Properties []autoOpenField `json:"properties"`
	Children   []*autoOpenTree `json:"children"`
}

func autoOpenProjection(n *mml.Node) *autoOpenTree {
	r := &autoOpenTree{Kind: n.Kind, Attributes: []autoOpenField{}, Properties: []autoOpenField{}, Children: []*autoOpenTree{}}
	if n.Kind == "mrow" && n.Flags.Inferred {
		r.Kind = "inferredMrow"
	}
	if n.Kind == "text" {
		v := n.Text
		r.Text = &v
	}
	for _, k := range n.Attributes.ExplicitNames() {
		v, _ := n.Attributes.GetExplicit(k)
		r.Attributes = append(r.Attributes, autoOpenField{k, v})
	}
	for _, k := range n.Properties.Keys() {
		v, _ := n.Property(k)
		r.Properties = append(r.Properties, autoOpenField{k, v})
	}
	for _, c := range n.Children {
		r.Children = append(r.Children, autoOpenProjection(c))
	}
	return r
}

type autoOpenCase struct {
	Name, TeX, Classification, PrimarySVG, AcceptedSVG string
	Display                                            bool
	Registrations                                      []struct {
		Name, Body string
		Arguments  int
	}
	OuterMacroCount           int
	Env                       map[string]any
	PrimaryTree, AcceptedTree *autoOpenTree
	PrimaryError              *struct{ ID, Message string }
	AcceptedError             *Error
}

func autoOpenCompile(c autoOpenCase) (*mml.Node, *parser, *Error, error) {
	state := newParseState()
	state.macroCount = c.OuterMacroCount
	for _, m := range c.Registrations {
		state.macros[m.Name] = macroDefinition{body: m.Body, arguments: m.Arguments}
	}
	font, _ := c.Env["font"].(string)
	p := &parser{source: c.TeX, state: state, display: c.Display, activeFont: font}
	children, _, err := p.parseRow(0, false)
	if err != nil {
		var pe *Error
		if !errors.As(err, &pe) {
			return nil, p, nil, err
		}
		return mathError(pe.Message, c.Display), p, pe, nil
	}
	children, err = p.amsTagFinalize(children)
	if err != nil {
		return nil, p, nil, err
	}
	root := node("math", children...)
	if c.Display {
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
	setMathMLInheritance(root, c.Display)
	root = moveMathLimits(root)
	cleanMathMLAttributes(root)
	return root, p, nil, nil
}
func TestDerivativeAutoOpenPinnedOutputs(t *testing.T) {
	var f struct {
		MathjaxGitCommit, PrimaryCaptureFreezeSHA256, AcceptedControlBase string
		Cases                                                             []autoOpenCase
	}
	b, e := os.ReadFile("../../testdata/derivative_autoopen_mathjax_3_2_2.json")
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(b, &f); e != nil {
		t.Fatal(e)
	}
	if len(f.Cases) != 106 || f.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || f.PrimaryCaptureFreezeSHA256 != "7eb21a59631885838bcf9ef98d7ad4d717c861cc541d619af1e0b5565078f424" || f.AcceptedControlBase != "d40e65bbed60c39f1ced29698e6a1a975b90a459" {
		t.Fatal("unbound corpus")
	}
	expectedScopes := map[string]string{
		"public-optional-dv-inline":             "whole-primary-target-or-control",
		"public-optional-dv-display":            "whole-primary-target-or-control",
		"public-optional-pdv-inline":            "whole-primary-target-or-control",
		"public-optional-pdv-display":           "whole-primary-target-or-control",
		"public-optional-pdv-one-inline":        "whole-primary-target-or-control",
		"public-optional-pdv-one-display":       "whole-primary-target-or-control",
		"public-plain-pdv-inline":               "whole-primary-target-or-control",
		"public-plain-pdv-display":              "whole-primary-target-or-control",
		"public-mixed-pdv-inline":               "whole-primary-target-or-control",
		"public-mixed-pdv-display":              "whole-primary-target-or-control",
		"public-partial-alias-inline":           "whole-primary-target-or-control",
		"public-partial-alias-display":          "whole-primary-target-or-control",
		"public-short-partial-alias-inline":     "whole-primary-target-or-control",
		"public-short-partial-alias-display":    "whole-primary-target-or-control",
		"public-functional-tail-inline":         "whole-primary-target-or-control",
		"public-functional-tail-display":        "whole-primary-target-or-control",
		"public-functional-alias-inline":        "whole-primary-target-or-control",
		"public-functional-alias-display":       "whole-primary-target-or-control",
		"public-short-functional-alias-inline":  "whole-primary-target-or-control",
		"public-short-functional-alias-display": "whole-primary-target-or-control",
		"public-prime-tail-inline":              "inherited-prime-own-property-diagnostic",
		"public-prime-tail-display":             "inherited-prime-own-property-diagnostic",
		"public-ignored-side-effect-inline":     "whole-primary-target-or-control",
		"public-ignored-side-effect-display":    "whole-primary-target-or-control",
		"public-star-held-inline":               "whole-primary-target-or-control",
		"public-star-held-display":              "whole-primary-target-or-control",
		"public-qty-held-inline":                "whole-primary-target-or-control",
		"public-qty-held-display":               "whole-primary-target-or-control",
		"public-differential-held-inline":       "unchanged-excluded-caller",
		"public-differential-held-display":      "unchanged-excluded-caller",
		"private-ignore-false-one-0":            "whole-primary-target-or-control",
		"private-ignore-true-two-0":             "whole-primary-target-or-control",
		"private-ignore-true-mixed-0":           "whole-primary-target-or-control",
		"private-ignore-true-empty-0":           "whole-primary-target-or-control",
		"private-ignored-sideeffect-0":          "whole-primary-target-or-control",
		"private-ignored-local-font-0":          "whole-primary-target-or-control",
		"private-pending-prime-0":               "inherited-prime-own-property-diagnostic",
		"private-alias-functional-0":            "whole-primary-target-or-control",
	}
	for _, c := range f.Cases {
		want := expectedScopes[c.Name]
		if want == "" {
			want = "whole-primary-target-or-control"
		}
		if c.Classification != want {
			t.Fatal("unapproved scope classification", c.Name, c.Classification)
		}
	}
	counts := map[string]int{}
	for _, c := range f.Cases {
		counts[c.Classification]++
		if c.Classification == "unresolved-complete-D095-composite" || c.Classification == "inherited-prime-own-property-diagnostic" {
			continue
		}
		t.Run(c.Name, func(t *testing.T) {
			wantSVG, wantTree := c.PrimarySVG, c.PrimaryTree
			var wantErr *Error
			if c.PrimaryError != nil {
				wantErr = &Error{ID: c.PrimaryError.ID, Message: c.PrimaryError.Message}
			}
			if c.Classification == "unchanged-excluded-caller" {
				wantSVG, wantTree, wantErr = c.AcceptedSVG, c.AcceptedTree, c.AcceptedError
			} else if c.Classification != "whole-primary-target-or-control" {
				t.Fatal("unknown scope")
			}
			root, _, actualErr, e := autoOpenCompile(c)
			if e != nil {
				t.Fatal(e)
			}
			if !reflect.DeepEqual(actualErr, wantErr) {
				t.Fatalf("error %#v want %#v", actualErr, wantErr)
			}
			b, e := json.Marshal(autoOpenProjection(root))
			if e != nil {
				t.Fatal(e)
			}
			var got *autoOpenTree
			if e = json.Unmarshal(b, &got); e != nil {
				t.Fatal(e)
			}
			if !reflect.DeepEqual(got, wantTree) {
				t.Error("complete ordered explicit/own tree differs")
			}
			if len(c.Registrations) == 0 && c.OuterMacroCount == 0 && len(c.Env) == 0 {
				compiled, e := NewCompiler().Compile(c.TeX, c.Display)
				if e != nil {
					t.Fatal(e)
				}
				v, _ := json.Marshal(autoOpenProjection(compiled))
				if string(v) != string(b) {
					t.Fatal("actual Compiler differs from registered-handler capture path")
				}
			}
			opts := pipeline.DefaultOptions()
			opts.Display = c.Display
			s, e := svg.NewTypesetter().Typeset(root, opts)
			if e != nil {
				t.Fatal(e)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256([]byte(s))); got != wantSVG {
				t.Errorf("whole SVG=%s want=%s", got, wantSVG)
			}
		})
	}
	if counts["whole-primary-target-or-control"] != 101 || counts["unchanged-excluded-caller"] != 2 || counts["unresolved-complete-D095-composite"] != 0 || counts["inherited-prime-own-property-diagnostic"] != 3 {
		t.Fatal("scope inventory changed", counts)
	}
}

func TestDerivativeAutoOpenDelivery(t *testing.T) {
	for _, input := range []string{`\dv{f}{x}(g)+Z`, `\pdv[2]{f}{x}(g)+Z`, `\pdv[2]{x}(g)+Z`, `\pdv[]{f}{x}(g)+Z`, `\pdv{f}{x}{y}(g)+Z`} {
		t.Run(input, func(t *testing.T) {
			p := &parser{source: input, state: newParseState()}
			p.pos++
			result, e := p.commandEvent(p.readControlSequence())
			if e != nil {
				t.Fatal(e)
			}
			if len(result.nodes) != 1 || result.nodes[0].Kind != "mfrac" || result.afterNode == nil || p.source[p.pos:] != "(g)+Z" {
				t.Fatal("fraction must precede untouched tail")
			}
			fraction := result.nodes[0]
			delivered := append([]*mml.Node{}, result.nodes...)
			state := p.state
			tail, e := result.afterNode.complete(p)
			if e != nil {
				t.Fatal(e)
			}
			if delivered[0] != fraction || fraction.Parent != nil || p.state != state || p.source[p.pos:] != "+Z" || result.afterNode.openCount != -1 || !result.afterNode.closed {
				t.Fatal("delivery/identity/cursor transition changed")
			}
			ignored := strings.HasPrefix(input, `\pdv`) && !strings.HasPrefix(input, `\pdv[2]{x}`)
			if result.afterNode.ignore != ignored || (len(tail) == 0) != ignored {
				t.Fatal("ignore cardinality/order-presence policy")
			}
			if !ignored {
				if len(tail) != 1 || tail[0].Kind != "mrow" {
					t.Fatal("missing replacement frame")
				}
				for _, name := range []string{"open", "close", "texClass"} {
					if _, ok := tail[0].Property(name); ok {
						t.Fatal("retained row fence marker", name)
					}
				}
				for _, child := range tail[0].Children {
					if child.Parent != tail[0] {
						t.Fatal("replacement ownership")
					}
				}
			}
		})
	}
	t.Run("pending-script-delivery", func(t *testing.T) {
		p := &parser{source: `\dv{f}{x}(g)+Z`, state: newParseState()}
		base := token("mi", "x")
		got, _, after, e := p.attachScriptWithFont([]*mml.Node{base}, '^', "")
		if e != nil {
			t.Fatal(e)
		}
		if len(got) != 1 || got[0].Kind != "msup" || got[0].Children[0] != base || got[0].Children[1].Kind != "mfrac" || p.source[p.pos:] != "(g)+Z" {
			t.Fatal("fraction not delivered to pending script before tail")
		}
		fraction := got[0].Children[1]
		tail, e := after.complete(p)
		if e != nil || len(tail) != 1 || fraction.Parent != got[0] || got[0].Children[1] != fraction {
			t.Fatal("tail consumed script or changed fraction identity", e)
		}
	})
}

// PrimeItem uses a different attachment branch from the ordinary pending
// script. The fraction must enter that superscript, while the AutoOpen
// replacement remains a following sibling in the containing row.
func TestDerivativeAutoOpenPendingPrimeScript(t *testing.T) {
	p := &parser{source: `x'^\dv{f}{x}(g)+Z`, state: newParseState()}
	nodes, _, err := p.parseRow(0, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 4 || nodes[0].Kind != "msup" || nodes[1].Kind != "mrow" {
		t.Fatal("wrong row recipient")
	}
	fractions := 0
	nodes[0].Walk(func(n *mml.Node) bool {
		if n.Kind == "mfrac" {
			fractions++
		}
		return true
	})
	if fractions != 1 || strings.Contains(textContent(nodes[0]), "g") || textContent(nodes[1]) != "(g)" || textContent(nodes[2]) != "+" || textContent(nodes[3]) != "Z" {
		t.Fatal("AutoOpen tail entered the pending prime script or displaced the fraction")
	}
}

func TestDerivativeAutoOpenSameParserState(t *testing.T) {
	for _, start := range []int{998, 999} {
		t.Run(fmt.Sprint(start), func(t *testing.T) {
			state := newParseState()
			state.macroCount = start
			state.macros["tailmark"] = macroDefinition{body: `\definecolor{tailcolor}{RGB}{255,0,0}`}
			state.macros["tailclose"] = macroDefinition{body: ")"}
			p := &parser{source: `\dv{f}{x}(\tailmark g\tailclose+Z`, state: state}
			_, _, e := p.parseRow(0, false)
			if state.macroCount != start+2 || p.state != state {
				t.Fatal("counter/state reset")
			}
			color, ce := state.colorModel.GetColor("named", "tailcolor")
			if ce != nil || color != "#ff0000" {
				t.Fatal("tail side effect lost", color, ce)
			}
			if start == 998 && e != nil {
				t.Fatal(e)
			}
			if start == 999 {
				var pe *Error
				if !errors.As(e, &pe) || pe.ID != "MaxMacroSub1" {
					t.Fatal("counter error suppressed", e)
				}
			}
		})
	}
	for _, tail := range []string{`\bf g`, `\bf g'`} {
		t.Run(tail, func(t *testing.T) {
			p := &parser{source: `\dv{f}{x}(` + tail + `)Z`, state: newParseState(), activeFont: "italic"}
			_, _, e := p.parseRow(0, false)
			if e != nil || p.activeFont != "italic" {
				t.Fatal("frame font escaped", e, p.activeFont)
			}
		})
	}
}
