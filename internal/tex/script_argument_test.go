// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/d2lang/mathjax-go/internal/mml"
)

func TestScriptArgumentPrimaryItems(t *testing.T) {
	var fixture struct {
		MathjaxGitCommit string
		Rows             []struct {
			Type, Kind, Marker, Event, Source, Stage, Remaining string
			Moves                                               *bool
			Allowed                                             bool
			Cursor                                              int
			Prepared                                            *struct {
				Kind                   string
				Position               int
				Reused                 bool
				Children               []*string
				Attributes, Properties map[string]any
			}
			Error struct{ ID, Message string }
		}
	}
	data, err := os.ReadFile("testdata/script_argument_items_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Rows) != 264 {
		t.Fatal("unbound actual handler/item references")
	}
	for _, c := range fixture.Rows {
		name := c.Type + "/" + c.Marker + "/" + c.Event + "/unset"
		if c.Moves != nil {
			name += map[bool]string{true: "/moves", false: "/fixed"}[*c.Moves]
		}
		if c.Allowed {
			name += "/allowed"
		}
		t.Run(name, func(t *testing.T) {
			core, under, over := limitToken("mi", "x"), limitToken("mi", "i"), limitToken("mi", "n")
			children := []*mml.Node{core, under, over}
			switch c.Kind {
			case "msub", "munder":
				children = []*mml.Node{core, under}
			case "msup", "mover":
				children = []*mml.Node{core, over}
			}
			if strings.HasSuffix(c.Type, "no-under") {
				children = []*mml.Node{core, nil, over}
			}
			if strings.HasSuffix(c.Type, "no-over") {
				children = []*mml.Node{core, under}
			}
			base := core
			if c.Kind != "mi" {
				base = texMMLFactory.Create(c.Kind)
				base.SetChildren(children)
				refreshDynamicFlags(base)
				base.Attributes.Set("id", "authored")
			}
			base.SetProperty("kept", 17)
			if c.Moves != nil {
				base.SetProperty("movesupsub", *c.Moves)
			}
			if c.Allowed {
				base.SetProperty("subsupOK", true)
			}
			before := base.Clone()
			originalChildren := append([]*mml.Node(nil), base.Children...)
			attachment, actualErr := prepareScriptAttachment(base, c.Marker[0], c.Moves != nil && *c.Moves)
			p := &parser{source: c.Source, state: newParseState()}
			stage := "prepare"
			if actualErr == nil {
				stage = "argument"
				if c.Prepared == nil {
					t.Fatal("primary preparation unexpectedly failed")
				}
				view := attachment.pendingBase()
				position := 1
				if c.Marker == "^" {
					position = 2
				}
				if attachment.base != base || attachment.reuse != c.Prepared.Reused || view.Kind != c.Prepared.Kind || position != c.Prepared.Position {
					t.Fatal("source prepared family, base identity, reuse or slot differs")
				}
				ids := map[*mml.Node]string{core: "core", under: "under", over: "over"}
				ids[base] = "base"
				got := []*string{}
				for _, n := range view.Children {
					if n == nil {
						got = append(got, nil)
					} else {
						id := ids[n]
						got = append(got, &id)
					}
				}
				if !reflect.DeepEqual(got, c.Prepared.Children) {
					t.Fatalf("prepared child identities: %v", got)
				}
				encoded, _ := json.Marshal(map[string]any{"attributes": explicitMap(view), "properties": ownMap(view)})
				var maps map[string]map[string]any
				if err := json.Unmarshal(encoded, &maps); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(maps["attributes"], c.Prepared.Attributes) || !reflect.DeepEqual(maps["properties"], c.Prepared.Properties) {
					t.Fatalf("prepared explicit/own maps differ: %s", encoded)
				}
				_, _, actualErr = p.parseScriptArgument(attachment, "")
			}
			var typed *Error
			if !errors.As(actualErr, &typed) || typed.ID != c.Error.ID || typed.Message != c.Error.Message || stage != c.Stage {
				t.Fatalf("stage/error: %s/%v; want %s/%+v", stage, actualErr, c.Stage, c.Error)
			}
			if len(utf16.Encode([]rune(p.source[:p.pos]))) != c.Cursor || p.source[p.pos:] != c.Remaining {
				t.Fatalf("cursor/suffix: %d/%q; want UTF16 %d/%q", p.pos, p.source[p.pos:], c.Cursor, c.Remaining)
			}
			if !reflect.DeepEqual(base, before) {
				t.Fatal("failed pending argument mutated original base/attrs/properties")
			}
			if len(base.Children) != len(originalChildren) {
				t.Fatal("base children changed")
			}
			for i, n := range originalChildren {
				if base.Children[i] != n || n != nil && n.Parent != base {
					t.Fatal("base child identity or ownership changed")
				}
			}
		})
	}
}
