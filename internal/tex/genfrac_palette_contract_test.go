// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

func TestGenfracDelimiterContracts(t *testing.T) {
	for _, c := range genfracReferences(t).Readers {
		t.Run(c.Name, func(t *testing.T) {
			if c.PrimaryCommand != "\\"+c.Command || len(c.Origins) == 0 {
				t.Fatal("unbound primary command")
			}
			state := newParseState()
			state.macroCount = c.Before.MacroCount
			p := &parser{source: c.Before.Source, pos: c.Before.Cursor, state: state}
			key, err := p.amsGenfracDelimiter(c.Command)
			genfracExpectError(t, err, c.Error)
			if !reflect.DeepEqual(key, c.Key) {
				t.Fatalf("raw delimiter key %v; want %v", key, c.Key)
			}
			got := genfracReferenceState{Source: p.source, Cursor: p.pos, MacroCount: state.macroCount}
			if got != c.After || p.state != state {
				t.Fatalf("reader state %+v; want %+v", got, c.After)
			}
		})
	}
}

func genfracCheckOwnFence(t *testing.T, got *mml.Node, open, close *string) {
	t.Helper()
	if got == nil || got.Kind != "mrow" || got.Flags.Inferred || got.TeXClass != mml.TeXClassOrd || !reflect.DeepEqual(got.Properties.Keys(), []string{"open", "close", "texClass"}) {
		t.Fatal("fixedFence explicit ORD row/own-property order")
	}
	for key, pointer := range map[string]*string{"open": open, "close": close} {
		value, present := got.Property(key)
		if !present || (pointer == nil && value != nil) || (pointer != nil && value != *pointer) {
			t.Fatalf("own %s value %v present=%v", key, value, present)
		}
	}
	if value, present := got.Property("texClass"); !present || value != mml.TeXClassOrd {
		t.Fatal("missing own ORD class")
	}
}

func TestGenfracFixedFenceIdentity(t *testing.T) {
	for _, c := range genfracReferences(t).Fences {
		t.Run(c.Name, func(t *testing.T) {
			p := &parser{state: newParseState(), display: true}
			fraction := node("mfrac", token("mi", "a"), token("mi", "b"))
			got, err := p.amsGenfracFixedFence(c.Open, fraction, c.Close)
			if err != nil {
				t.Fatal(err)
			}
			genfracCheckOwnFence(t, got, c.Open, c.Close)
			index, count := 0, 1
			if c.Open != nil {
				index++
				count++
			}
			if c.Close != nil {
				count++
			}
			if len(got.Children) != count || got.Children[index] != fraction || fraction.Parent != got {
				t.Fatal("fixedFence changed fraction identity/parent or absent-vs-dot children")
			}
			// The live caller, not fixedFence, owns the withDelims annotation.
			caller := &parser{source: c.TeX, pos: len("\\genfrac"), state: newParseState(), display: true}
			nodes, err := caller.amsGenfrac("genfrac")
			if err != nil || len(nodes) != 1 {
				t.Fatalf("live Genfrac: nodes=%d error=%v", len(nodes), err)
			}
			genfracCheckOwnFence(t, nodes[0], c.Open, c.Close)
			f := nodes[0].Children[index]
			if f.Kind != "mfrac" || f.Parent != nodes[0] {
				t.Fatal("caller fraction ownership")
			}
			if value, present := f.Property("withDelims"); !present || value != true {
				t.Fatal("caller omitted withDelims")
			}
		})
	}
}
