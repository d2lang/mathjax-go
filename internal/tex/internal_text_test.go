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
)

func TestInternalTextRegisteredMethod(t *testing.T) {
	var fixture struct {
		Rows []struct {
			Input struct {
				Name, Text, Font string
				Level            *int
				Env              map[string]string
				OuterMacroCount  int
				Registration     *struct {
					Name, Body string
					Arguments  int
				}
			}
			Error                       *struct{ ID, Message string }
			Nodes                       []*binomialTree
			EnvBefore, EnvAfter         map[string]string
			CounterBefore, CounterAfter int
			ReturnedNodesCreated        *bool
		}
	}
	b, err := os.ReadFile("testdata/internal_text_method_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Rows) != 28 {
		t.Fatal("missing actual internalMath records")
	}
	for _, c := range fixture.Rows {
		t.Run(c.Input.Name, func(t *testing.T) {
			state := newParseState()
			state.macroCount = c.Input.OuterMacroCount
			if m := c.Input.Registration; m != nil {
				state.macros[m.Name] = macroDefinition{body: m.Body, arguments: m.Arguments}
			}
			p := &parser{source: "untouched", pos: 2, state: state, activeFont: c.Input.Env["font"], display: true}
			before := *p
			if c.Input.Level != nil && *c.Input.Level != 0 {
				t.Fatal("unexpected source level")
			}
			nodes, err := p.internalMath(c.Input.Text, c.Input.Font, c.Input.Level != nil)
			if !reflect.DeepEqual(*p, before) || p.state != state || state.macroCount != c.CounterBefore || state.macroCount != c.CounterAfter || !reflect.DeepEqual(c.EnvBefore, c.EnvAfter) {
				t.Fatal("inner parser mutated outer cursor, environment or counter")
			}
			if c.Error != nil {
				var e *Error
				if !errors.As(err, &e) || e.ID != c.Error.ID || e.Message != c.Error.Message || nodes != nil || c.Nodes != nil {
					t.Fatalf("error %v; want %+v", err, c.Error)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if c.ReturnedNodesCreated == nil || !*c.ReturnedNodesCreated {
				t.Fatal("primary returned a foreign node")
			}
			got := []*binomialTree{}
			seen := map[*mml.Node]bool{}
			for _, n := range nodes {
				if n.Parent != nil {
					t.Fatal("returned root already has a parent")
				}
				n.Walk(func(n *mml.Node) bool {
					if seen[n] {
						t.Fatal("node has multiple owners")
					}
					seen[n] = true
					for _, child := range n.Children {
						if child.Parent != n {
							t.Fatal("lost child identity")
						}
					}
					// These existing private markers lower lexical scopes. Assert
					// their type before removing them from the primary comparison;
					// the public compiler test separately requires no leaked marker.
					for _, key := range []string{resolvedFontScope, ambientFontSource, vectorFactoryToken, vectorFactoryDone, limitsScriptOrigin} {
						if value, ok := n.Property(key); ok {
							if value != true {
								t.Fatalf("invalid %s marker", key)
							}
							n.RemoveProperty(key)
						}
					}
					return true
				})
				got = append(got, binomialProjection(n))
			}
			b, err := json.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			if err = json.Unmarshal(b, &got); err != nil {
				t.Fatal(err)
			}
			// At this private pre-postfilter boundary the primary retains a
			// two-child msubsup. The inherited Go script constructor already
			// emits msub. Bind only these two exact paths; all public trees
			// are compared without this representation qualification.
			paths := map[string][]int{"mixed": {1, 0, 0}, "nested-braces": {1, 0, 0, 0, 0}}
			if path, ok := paths[c.Input.Name]; ok {
				n := c.Nodes[0]
				for _, i := range path {
					if i >= len(n.Children) {
						t.Fatal("missing inherited script path")
					}
					n = n.Children[i]
				}
				if n.Kind != "msubsup" || len(n.Children) != 2 || len(n.Attributes) != 0 || len(n.Properties) != 0 {
					t.Fatal("changed primary private script boundary")
				}
				n.Kind = "msub"
			}
			if !reflect.DeepEqual(got, c.Nodes) {
				t.Errorf("actual internalMath tree differs: got %s", b)
			}
		})
	}
}
