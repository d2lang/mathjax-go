// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package svg

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/ordered"
)

type identifierClassSnapshot struct {
	TeXClass                        mml.TeXClass
	PrevClass                       *mml.TeXClass
	PrevLevel                       *int
	Properties, Explicit, Inherited map[string]any
}

func TestIdentifierOperatorRegisteredMethod(t *testing.T) {
	var fixture struct {
		Rows []struct {
			Input struct {
				Name, Text, Variant  string
				Class                mml.TeXClass
				Properties, Explicit map[string]any
				Previous             *struct {
					Class mml.TeXClass
					Level int
				}
			}
			First, Second                                                          identifierClassSnapshot
			ReturnsSelf, IdentityUnchanged, AttributesUnchanged, PreviousUnchanged bool
		}
	}
	data, err := os.ReadFile("testdata/identifier_operator_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Rows) != 167 {
		t.Fatal("unbound registered method references")
	}
	mapOf := func(m *ordered.Map[any]) map[string]any {
		out := map[string]any{}
		m.Range(func(k string, v any) bool { out[k] = v; return true })
		return out
	}
	for _, c := range fixture.Rows {
		t.Run(c.Input.Name, func(t *testing.T) {
			if !c.ReturnsSelf || !c.IdentityUnchanged || !c.AttributesUnchanged || !c.PreviousUnchanged {
				t.Fatal("changed primary identity contract")
			}
			child := mml.NewText(c.Input.Text)
			n := mml.NewNode("mi", nil, nil, child)
			n.TeXClass = c.Input.Class
			n.Attributes.SetInherited("mathvariant", c.Input.Variant)
			for k, v := range c.Input.Explicit {
				n.Attributes.Set(k, v)
			}
			for k, v := range c.Input.Properties {
				n.SetProperty(k, v)
			}
			var previous *mml.Node
			if c.Input.Previous != nil {
				previous = mml.NewNode("mi", nil, nil, mml.NewText("x"))
				previous.TeXClass = c.Input.Previous.Class
				previous.Attributes.SetInherited("scriptlevel", c.Input.Previous.Level)
			}
			for _, want := range []identifierClassSnapshot{c.First, c.Second} {
				if returned := setTeXClass(n, previous); returned != n || n.Parent != nil || len(n.Children) != 1 || n.Children[0] != child || child.Parent != n {
					t.Fatal("node identity or parent changed")
				}
				// The established Go representation of absent previous state is
				// NONE/0; primary factory nodes retain null/null without a predecessor.
				prevClass, prevLevel := mml.TeXClassNone, 0
				if c.Input.Previous == nil {
					if want.PrevClass != nil || want.PrevLevel != nil {
						t.Fatal("changed null previous-state boundary")
					}
				} else {
					if want.PrevClass == nil || want.PrevLevel == nil {
						t.Fatal("missing primary previous state")
					}
					prevClass, prevLevel = *want.PrevClass, *want.PrevLevel
				}
				if n.TeXClass != want.TeXClass || n.PrevClass != prevClass || n.PrevLevel != prevLevel {
					t.Error("class or previous state differs")
				}
				if !reflect.DeepEqual(mapOf(n.Properties), want.Properties) || !reflect.DeepEqual(mapOf(n.Attributes.Explicit()), want.Explicit) || !reflect.DeepEqual(mapOf(n.Attributes.Inherited()), want.Inherited) {
					t.Error("complete own properties or attribute layers differ")
				}
				if previous != nil {
					level, ok := previous.Attributes.Get("scriptlevel")
					if previous.TeXClass != c.Input.Previous.Class || !ok || level != c.Input.Previous.Level {
						t.Fatal("previous node mutated")
					}
				}
			}
		})
	}
}
