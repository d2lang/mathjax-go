// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"encoding/json"
	"errors"
	"github.com/d2lang/mathjax-go/internal/mml"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFramedTextRegisteredHandlers(t *testing.T) {
	var fixture struct {
		Rows []struct {
			Input struct {
				Name, Command, Source, Font string
				OuterMacroCount             int
				Registration                *struct {
					Name, Body string
					Arguments  int
				}
			}
			Error                                                                                                                *struct{ ID, Message string }
			Nodes                                                                                                                []*binomialTree
			Cursor, Counter, PushCount, OuterCreationCount                                                                       int
			Remaining, LateColor                                                                                                 string
			EnvBefore, EnvAfter                                                                                                  map[string]string
			SourceUnchanged, EnvIdentity, ConfigIdentity, ColorModelIdentity, AllPushedCreated, RootParentsNull, Unique, Parents bool
		}
	}
	b, err := os.ReadFile("testdata/framed_text_method_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Rows) != 12 {
		t.Fatal("missing actual primary handlers")
	}
	for _, c := range fixture.Rows {
		t.Run(c.Input.Name, func(t *testing.T) {
			state := newParseState()
			state.macroCount = c.Input.OuterMacroCount
			if r := c.Input.Registration; r != nil {
				state.macros[r.Name] = macroDefinition{body: r.Body, arguments: r.Arguments}
			}
			model := state.colorModel
			p := &parser{source: c.Input.Source, state: state, activeFont: c.Input.Font}
			nodes, err := p.command(c.Input.Command)
			late, e := model.GetColor("named", "late")
			if e != nil {
				t.Fatal(e)
			}
			// Optional evidence output preserves the complete actual own maps before
			// the three existing compiler-private scope markers are projected away below.
			if out := os.Getenv("D082_METHOD_OUT"); out != "" {
				if e := os.MkdirAll(out, 0755); e != nil {
					t.Fatal(e)
				}
				raw := []*binomialTree{}
				for _, n := range nodes {
					raw = append(raw, binomialProjection(n))
				}
				r := map[string]any{"nodes": raw, "cursor": p.pos, "remaining": p.source[p.pos:], "counter": state.macroCount, "lateColor": late}
				if err != nil {
					r["error"] = err.Error()
				}
				b, _ := json.MarshalIndent(r, "", "  ")
				if e := os.WriteFile(filepath.Join(out, c.Input.Name+".json"), append(b, '\n'), 0644); e != nil {
					t.Fatal(e)
				}
			}
			if !c.SourceUnchanged || !c.EnvIdentity || !c.ConfigIdentity || !c.ColorModelIdentity || !c.AllPushedCreated || !c.RootParentsNull || !c.Unique || !c.Parents {
				t.Fatal("invalid primary ownership/state control")
			}
			if p.source != c.Input.Source || p.pos != c.Cursor || p.source[p.pos:] != c.Remaining || p.activeFont != c.Input.Font || p.state != state || state.colorModel != model || state.macroCount != c.Counter || c.Counter != c.Input.OuterMacroCount || late != c.LateColor || !reflect.DeepEqual(c.EnvBefore, c.EnvAfter) {
				t.Fatalf("handler state differs: cursor=%d counter=%d late=%q", p.pos, state.macroCount, late)
			}
			if r := c.Input.Registration; r != nil {
				if !reflect.DeepEqual(state.macros[r.Name], macroDefinition{body: r.Body, arguments: r.Arguments}) {
					t.Fatal("mutated shared registration")
				}
			}

			if c.Error != nil {
				var pe *Error
				if !errors.As(err, &pe) || pe.ID != c.Error.ID || pe.Message != c.Error.Message || len(nodes) != 0 || len(c.Nodes) != 0 || c.PushCount != 0 || c.OuterCreationCount != 0 {
					t.Fatalf("error/outer return %v; want %+v", err, c.Error)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(nodes) != 1 || c.PushCount != 1 || c.OuterCreationCount != 1 {
				t.Fatal("wrong outer box count")
			}
			scopeTokens := 0
			seen := map[*mml.Node]bool{}
			got := []*binomialTree{}
			for _, n := range nodes {
				if n.Parent != nil {
					t.Fatal("returned box already owned")
				}
				n.Walk(func(n *mml.Node) bool {
					if seen[n] {
						t.Fatal("duplicate child identity")
					}
					seen[n] = true
					for _, ch := range n.Children {
						if ch.Parent != n {
							t.Fatal("lost child parent")
						}
					}
					// These six scalar inner parses carry exactly three inherited Go
					// implementation markers until Compiler cleanup. Require the
					// complete triple on mi(x), retain all other own properties,
					// and require the named-case token count below.
					if _, ok := n.Property(resolvedFontScope); ok {
						if n.Kind != "mi" || len(n.Children) != 1 || n.Children[0].Kind != "text" || n.Children[0].Text != "x" {
							t.Fatal("unexpected scope token")
						}
						for _, key := range []string{resolvedFontScope, ambientFontSource, vectorFactoryToken} {
							if v, ok := n.Property(key); !ok || v != true {
								t.Fatal("changed internal scope marker", key)
							}
							n.RemoveProperty(key)
						}
						scopeTokens++
					}
					return true
				})
				got = append(got, binomialProjection(n))
			}
			wantScopeTokens := 0
			switch c.Input.Name {
			case "colorbox-mixed", "fcolorbox-mixed", "font-environment", "shared-macro-success", "late-color", "late-frame":
				wantScopeTokens = 1
			}
			if scopeTokens != wantScopeTokens {
				t.Fatal("changed private scope coverage")
			}

			b, err := json.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			if err = json.Unmarshal(b, &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, c.Nodes) {
				t.Errorf("actual handler tree differs: %s", b)
			}
		})
	}
}
