// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package svg

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/d2lang/mathjax-go/internal/layout"
	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/tex"
)

// Overlay compiled MathML rather than use the separately unsupported mmlToken
// TeX command. The hashes come from the unmodified pinned MathJax 3.2.2 output
// with the identical overlays applied before wrapper construction.
func TestExplicitMathMLSpacingFrozenSVG(t *testing.T) {
	data, err := os.ReadFile("testdata/explicit_spacing_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name, TeX, SHA256 string
			Display           bool
			PrependEmpty      bool
			Operators         []struct {
				Text           string
				Lspace, Rspace float64
			}
			Overlays []struct {
				Index               int
				Explicit, Inherited map[string]any
				RemoveClass         bool
			}
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 46 {
		t.Fatal("incomplete frozen spacing matrix")
	}
	for _, test := range fixture.Cases {
		t.Run(test.Name, func(t *testing.T) {
			root, err := tex.NewCompiler().Compile(test.TeX, test.Display)
			if err != nil {
				t.Fatal(err)
			}
			if test.PrependEmpty {
				row := root.Children[0]
				if row.Kind != "mrow" || !row.Flags.Inferred {
					t.Fatal("expected inferred root row")
				}
				empty := mml.NewNode("mo", nil, nil)
				empty.Flags.Token, empty.Flags.Embellished = true, true
				empty.TeXClass = mml.TeXClassOrd
				empty.Attributes.Set("rspace", "0")
				empty.Attributes.SetInherited("scriptlevel", 0)
				// checkOperatorTable's empty-mo dictionary entry, independently
				// verified by TestOperatorMathMLSpacingDefaults in internal/tex.
				empty.OperatorLspace, empty.OperatorRspace = 1.0/18, 1.0/18
				row.SetChildren(append([]*mml.Node{empty}, row.Children...))
				row.Flags.Embellished, root.Flags.Embellished = false, false
			}
			nodes := root.Find("mo")
			if len(nodes) != len(test.Operators) {
				t.Fatal("operator count differs from pinned MathML")
			}
			for i, n := range nodes {
				want := test.Operators[i]
				if n.OperatorLspace != want.Lspace || n.OperatorRspace != want.Rspace {
					t.Fatalf("%s dictionary spaces = %g/%g, want %g/%g", want.Text, n.OperatorLspace, n.OperatorRspace, want.Lspace, want.Rspace)
				}
			}
			for _, o := range test.Overlays {
				if o.Index < 0 || o.Index >= len(nodes) {
					t.Fatal("missing overlay operator")
				}
				n := nodes[o.Index]
				for k, v := range o.Explicit {
					n.Attributes.Set(k, v)
				}
				for k, v := range o.Inherited {
					n.Attributes.SetInherited(k, v)
				}
				if o.RemoveClass {
					n.RemoveProperty("texClass")
				}
			}
			options := pipeline.DefaultOptions()
			options.Display = test.Display
			got, err := NewTypesetter().Typeset(root, options)
			if err != nil {
				t.Fatal(err)
			}
			hash := fmt.Sprintf("%x", sha256.Sum256([]byte(got)))
			if hash != test.SHA256 {
				t.Fatalf("complete SVG = %s, want frozen MathJax 3.2.2 %s", hash, test.SHA256)
			}
			clone, err := NewTypesetter().Typeset(root.Clone(), options)
			if err != nil || clone != got {
				t.Fatalf("cloned operator state changed SVG: %v", err)
			}
		})
	}
}

func TestMathMLSpacingActualParentAndTopWrapper(t *testing.T) {
	mo := func(left, right string) *mml.Node {
		n := mml.NewNode("mo", nil, nil, mml.NewText("+"))
		n.Flags.Token, n.Flags.Embellished = true, true
		if left != "" {
			n.Attributes.Set("lspace", left)
		}
		if right != "" {
			n.Attributes.Set("rspace", right)
		}
		return n
	}
	mi := func() *mml.Node { n := mml.NewNode("mi", nil, nil, mml.NewText("x")); n.Flags.Token = true; return n }
	for _, name := range []string{"inferred multi-row", "one-child row", "fixed-arity parent", "embellished script", "adjacent embellished", "intervening space"} {
		t.Run(name, func(t *testing.T) {
			core := mo("1em", ".5em")
			var root, owner *mml.Node
			wantL, wantR := 1.0, .5
			switch name {
			case "inferred multi-row":
				root = mml.NewNode("mrow", nil, nil, core, mi())
				root.Flags.Inferred = true
				owner = core
			case "one-child row":
				root = mml.NewNode("mrow", nil, nil, core)
				owner = core
				wantL, wantR = 0, 0
			case "fixed-arity parent":
				root = mml.NewNode("mfrac", nil, nil, core, mi())
				owner = core
				wantL, wantR = 0, 0
			case "embellished script":
				owner = mml.NewNode("msub", nil, nil, core, mi())
				owner.Flags.Embellished = true
				root = mml.NewNode("mrow", nil, nil, owner, mi())
			case "adjacent embellished", "intervening space":
				previous := mo("0", ".75em")
				outer := mml.NewNode("msub", nil, nil, previous, mi())
				outer.Flags.Embellished = true
				children := []*mml.Node{outer, core}
				if name == "intervening space" {
					space := mml.NewNode("mspace", nil, nil)
					space.Flags.Spacelike = true
					children = []*mml.Node{outer, space, core}
				} else {
					wantL = .25
				}
				root = mml.NewNode("mrow", nil, nil, children...)
				owner = core
			}
			r := &renderer{params: layout.TeXParameters, pxPerEm: 8 / layout.TeXParameters.XHeight}
			wrapped := r.wrap(root, nil, 0, true)
			var find func(*wrapper, *mml.Node) *wrapper
			find = func(w *wrapper, n *mml.Node) *wrapper {
				if w.node == n {
					return w
				}
				for _, c := range w.children {
					if found := find(c, n); found != nil {
						return found
					}
				}
				return nil
			}
			got := find(wrapped, owner).bbox
			if math.Abs(got.L-wantL) > 1e-15 || math.Abs(got.R-wantR) > 1e-15 {
				t.Fatalf("outer spacing %g/%g, want %g/%g", got.L, got.R, wantL, wantR)
			}
			if owner != core {
				inner := find(wrapped, core).bbox
				if inner.L != 0 || inner.R != 0 {
					t.Fatalf("inner core duplicated spacing %g/%g", inner.L, inner.R)
				}
			}
			if value, _ := core.Attributes.GetExplicit("lspace"); value != "1em" {
				t.Fatal("layout rewrote authored spacing")
			}
		})
	}
}
