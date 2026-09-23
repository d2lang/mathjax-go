// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"encoding/json"
	"errors"
	"github.com/d2lang/mathjax-go/internal/mml"
	"os"
	"reflect"
	"testing"
)

func TestReferenceInternalTextMethod(t *testing.T) {
	type pair struct {
		Name  string
		Value any
	}
	type label struct{ Tag, ID string }
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Input struct {
				Name, Source, Command, Tag, ID, Font, Label string
				Eqref                                       bool
				MacroCount                                  int
				Secondary, Unrelated                        *label
				Registration                                *struct {
					Name, Body string
					Arguments  int
				}
			}
			Nodes   []*binomialTree
			Error   *struct{ ID, Message string }
			Parents bool
			Outer   struct {
				CursorUTF16, MacroCount                                                                                     int
				Source, Remaining, LateColor                                                                                string
				EnvironmentIdentity, ConfigurationIdentity, CurrentIdentity, LabelsIdentity, FactoryIdentity, ColorIdentity bool
			}
		}
	}
	var bounds struct {
		Baseline string
		Markers  []struct {
			Name  string
			Nodes []struct {
				Path                            []int
				Kind                            string
				Text                            *string
				Attributes, Properties, Markers []pair
			}
		}
		UnclosedArgument struct {
			Name                                string
			PrimaryCursor, AcceptedCursor       int
			PrimaryRemaining, AcceptedRemaining string
		}
		ObserverError struct{ Name, Primary, Accepted string }
	}
	for file, target := range map[string]any{"testdata/reference_text_method_mathjax_3_2_2.json": &fixture, "testdata/reference_text_method_boundaries.json": &bounds} {
		b, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(b, target); err != nil {
			t.Fatal(err)
		}
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 17 || bounds.Baseline != "787559bb7c9515893592f04861dcac78ff83e399" || len(bounds.Markers) != 5 {
		t.Fatal("unbound reference method fixtures")
	}
	markerPaths := map[string][]int{"mixed-math": {0, 0, 1, 0, 0}, "math-only": {0, 0, 0, 0}, "caller-font": {0, 0, 1, 0, 0}, "nested-reference": {0, 0, 1, 0, 0, 0}, "shared-macro": {0, 0, 1, 0, 0}}
	markerCount := 0
	for _, c := range fixture.Cases {
		t.Run(c.Input.Name, func(t *testing.T) {
			in := c.Input
			state := newParseState()
			state.macroCount = in.MacroCount
			if m := in.Registration; m != nil {
				state.macros[m.Name] = macroDefinition{body: m.Body, arguments: m.Arguments}
			}
			p := &parser{source: in.Source, pos: len(`\DRefObserve`), state: state, activeFont: in.Font, display: true}
			tags := p.amsTags()
			tags.labels[in.Label] = amsLabel{tag: in.Tag, id: in.ID}
			if x := in.Secondary; x != nil {
				tags.labels["u"] = amsLabel{tag: x.Tag, id: x.ID}
			}
			if x := in.Unrelated; x != nil {
				tags.labels["other"] = amsLabel{tag: x.Tag, id: x.ID}
			}
			current, model, labelMap := tags.current, state.colorModel, reflect.ValueOf(tags.labels).Pointer()
			labels := map[string]amsLabel{}
			for k, v := range tags.labels {
				labels[k] = v
			}
			nodes, err := p.amsHandleReference(in.Command, in.Eqref)
			if !c.Parents || !c.Outer.EnvironmentIdentity || !c.Outer.ConfigurationIdentity || !c.Outer.CurrentIdentity || !c.Outer.LabelsIdentity || !c.Outer.FactoryIdentity || !c.Outer.ColorIdentity {
				t.Fatal("primary identity contract changed")
			}
			if p.state != state || state.amsTags != tags || tags.current != current || state.colorModel != model || reflect.ValueOf(tags.labels).Pointer() != labelMap || !reflect.DeepEqual(tags.labels, labels) || p.activeFont != in.Font || p.source != in.Source || state.macroCount != in.MacroCount || state.macroCount != c.Outer.MacroCount {
				t.Fatal("reference changed outer cursor/configuration/font/counter/labels")
			}
			cursor, remaining := c.Outer.CursorUTF16, c.Outer.Remaining
			if in.Name == "unclosed-argument" {
				// The unchanged argument reader stops at {t; primary consumes to EOF.
				// This one preexisting error cursor is preserved, not made primary-equal.
				q := bounds.UnclosedArgument
				if q.Name != in.Name || q.PrimaryCursor != cursor || q.PrimaryRemaining != remaining || q.AcceptedCursor != 12 || q.AcceptedRemaining != "{t" {
					t.Fatal("unbound argument-reader boundary")
				}
				cursor, remaining = q.AcceptedCursor, q.AcceptedRemaining
			}
			if p.pos != cursor || p.source[p.pos:] != remaining {
				t.Fatal("reference argument cursor differs")
			}
			color, colorErr := model.GetColor("named", "late")
			if colorErr != nil || color != c.Outer.LateColor {
				t.Fatalf("shared color state %q %v; want %q", color, colorErr, c.Outer.LateColor)
			}
			if c.Error != nil {
				want := c.Error.Message
				if in.Name == "missing-argument" {
					// Primary GetArgument reports the registered observer currentCS; this
					// direct Go call reports its canonical command argument. Public ref/eqref
					// missing-argument errors are separately whole-primary assertions.
					q := bounds.ObserverError
					if q.Name != in.Name || q.Primary != want || q.Accepted != "Missing argument for \\ref" {
						t.Fatal("unbound observer-name boundary")
					}
					want = q.Accepted
				}
				var pe *Error
				if !errors.As(err, &pe) || pe.ID != c.Error.ID || pe.Message != want || len(nodes) != 0 || len(c.Nodes) != 0 {
					t.Fatalf("error %v; want %s %s and no outer row", err, c.Error.ID, want)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(nodes) != 1 || nodes[0].Kind != "mrow" || nodes[0].Parent != nil {
				t.Fatal("reference outer row/cardinality changed")
			}
			seen := map[*mml.Node]bool{}
			nodes[0].Walk(func(n *mml.Node) bool {
				if seen[n] {
					t.Fatal("duplicate node ownership")
				}
				seen[n] = true
				for _, ch := range n.Children {
					if ch.Parent != n {
						t.Fatal("child identity/parent lost")
					}
				}
				return true
			})
			got := []*binomialTree{binomialProjection(nodes[0])}
			b, _ := json.Marshal(got)
			if err = json.Unmarshal(b, &got); err != nil {
				t.Fatal(err)
			}
			if path, ok := markerPaths[in.Name]; ok {
				found := false
				for _, bound := range bounds.Markers {
					if bound.Name != in.Name {
						continue
					}
					if found || len(bound.Nodes) != 1 || !reflect.DeepEqual(bound.Nodes[0].Path, path) {
						t.Fatal("changed finite marker path")
					}
					found = true
					q := bound.Nodes[0]
					n := got[path[0]]
					for _, i := range path[1:] {
						if i >= len(n.Children) {
							t.Fatal("missing marker node")
						}
						n = n.Children[i]
					}
					attrs, props := map[string]any{}, map[string]any{}
					for _, x := range q.Attributes {
						attrs[x.Name] = x.Value
					}
					for _, x := range q.Properties {
						props[x.Name] = x.Value
					}
					if n.Kind != q.Kind || !reflect.DeepEqual(n.Text, q.Text) || !reflect.DeepEqual(n.Attributes, attrs) || !reflect.DeepEqual(n.Properties, props) {
						t.Fatal("accepted helper raw marker map changed")
					}
					keys := []string{vectorFactoryToken, ambientFontSource, resolvedFontScope}
					if in.Name == "nested-reference" {
						keys = []string{resolvedFontScope}
					}
					if len(q.Markers) != len(keys) {
						t.Fatal("marker set changed")
					}
					for i, k := range keys {
						if q.Markers[i].Name != k || q.Markers[i].Value != true || n.Properties[k] != true {
							t.Fatal("invalid named helper marker")
						}
						delete(n.Properties, k)
					}
				}
				if !found {
					t.Fatal("missing accepted helper provenance")
				}
				markerCount++
			}
			if !reflect.DeepEqual(got, c.Nodes) {
				t.Errorf("complete primary method tree differs: %s", b)
			}
		})
	}
	if markerCount != 5 {
		t.Fatal("changed finite pre-Compile marker coverage")
	}
}
