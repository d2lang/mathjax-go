// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
)

func operatorNameFinalize(children []*mml.Node, display bool) *mml.Node {
	root := node("math", children...)
	if display {
		root.Attributes.Set("display", "block")
	}
	root.Walk(func(n *mml.Node) bool {
		for _, k := range []string{resolvedFontScope, ambientFontSource, vectorFactoryToken, vectorFactoryDone, limitsScriptOrigin} {
			n.RemoveProperty(k)
		}
		return true
	})
	setMathMLInheritance(root, display)
	root = moveMathLimits(root)
	cleanMathMLAttributes(root)
	return root
}

func operatorNameReferenceOutput(t *testing.T, root *mml.Node, display bool, expected macroReferenceOutput) {
	t.Helper()
	if expected.Error != nil {
		t.Fatal("successful primary fixture required")
	}
	encoded, err := json.Marshal(macroReferenceTree(root))
	if err != nil {
		t.Fatal(err)
	}
	var got, want any
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(expected.Tree, &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatal("complete ordered primary tree differs")
	}
	opts := pipeline.DefaultOptions()
	opts.Display = display
	result, err := svg.NewTypesetter().Typeset(root, opts)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256([]byte(result))) != expected.SVGSHA256 {
		t.Fatal("complete primary SVG differs")
	}
}

func TestOperatorNameClassReferences(t *testing.T) {
	var f struct {
		MathjaxGitCommit string
		Cases            []macroReferenceCase
	}
	b, err := os.ReadFile("testdata/operatorname_class_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	if f.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(f.Cases) != 26 {
		t.Fatal("unbound whole primary fixtures")
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			root, err := NewCompiler().Compile(c.TeX, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			operatorNameReferenceOutput(t, root, c.Display, c.Primary)
		})
	}
}

// These are the eight actual registered-handler inputs retained in the frozen
// primary observation. Cursor values deliberately retain the Go nonstar-space
// boundary; they do not claim primary whitespace consumption. Runtime texClass
// and own properties are checked separately on the same retained nodes.
func TestOperatorNameActualHandlerClassification(t *testing.T) {
	cases := []struct {
		name, source, kind string
		cursor             int
		star               bool
		texts              []string
	}{
		{"single-letter", "{a} tail", "mi", 3, false, []string{"a"}},
		{"single-multi-letter", "{arg} tail", "mi", 5, false, []string{"arg"}},
		{"split-single", "{a b} tail", "TeXAtom", 5, false, []string{"a", "b"}},
		{"split-multi", "{arg max} tail", "TeXAtom", 9, false, []string{"arg", "max"}},
		{"explicit-space", `{arg\,max} tail`, "TeXAtom", 10, false, []string{"arg", "max"}},
		{"empty", "{} tail", "TeXAtom", 2, false, []string{}},
		{"starred", "*{a b}_i^n", "TeXAtom", 6, true, []string{"a", "b"}},
		{"missing-argument", "", "", 0, false, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			state := newParseState()
			p := &parser{source: c.source, state: state}
			nodes, handled, err := p.baseAMSCommand("operatorname")
			if !handled || p.state != state || p.source != c.source || p.pos != c.cursor || state.macroCount != 0 {
				t.Fatal("handler caller ownership/cursor changed")
			}
			if c.kind == "" {
				macroReferenceError(t, err, &Error{ID: "MissingArgFor", Message: `Missing argument for \operatorname`})
				if nodes != nil {
					t.Fatal("failed handler emitted nodes")
				}
				return
			}
			if err != nil || len(nodes) != 1 {
				t.Fatalf("handler result %v %v", nodes, err)
			}
			owner := nodes[0]
			if owner.Kind != c.kind || owner.TeXClass != mml.TeXClassOp {
				t.Fatal("returned owner kind/class changed")
			}
			identifiers := owner.Find("mi")
			if len(identifiers) != len(c.texts) {
				t.Fatal("identifier inventory changed")
			}
			for i, n := range identifiers {
				variant, _ := n.Attributes.Get("mathvariant")
				if textContent(n) != c.texts[i] || variant != "normal" {
					t.Fatal("identifier text/font changed")
				}
				for _, key := range []string{vectorFactoryToken, vectorFactoryDone} {
					if v, ok := n.Property(key); !ok || v != true {
						t.Fatal("vector provenance changed", key)
					}
				}
				if n != owner {
					if n.TeXClass != mml.TeXClassOrd {
						t.Fatal("child identifier must start ORD")
					}
					if _, ok := n.Property("texClass"); ok {
						t.Fatal("child has forced own texClass")
					}
				}
				if _, ok := n.Property("autoOP"); ok {
					t.Fatal("autoOP must belong to rendering phase")
				}
			}
			// Save real pointers, then use the same inherited compiler finalization and
			// actual SVG renderer. No snapshot clone substitutes for the retained nodes.
			root := operatorNameFinalize(nodes, false)
			props := macroReferencePairs(owner.Properties)
			want := []map[string]any{{"name": "texClass", "value": mml.TeXClassOp}, {"name": "movesupsub", "value": c.star}, {"name": "movablelimits", "value": true}}
			if owner.Kind == "mi" {
				want = []map[string]any{want[1], want[2], want[0]}
			}
			if !reflect.DeepEqual(props, want) {
				t.Fatalf("owner property order %v, want %v", props, want)
			}
			beforeParents := map[*mml.Node]*mml.Node{}
			root.Walk(func(n *mml.Node) bool { beforeParents[n] = n.Parent; return true })
			if _, err := svg.NewTypesetter().Typeset(root, pipeline.DefaultOptions()); err != nil {
				t.Fatal(err)
			}
			after := map[*mml.Node]*mml.Node{}
			root.Walk(func(n *mml.Node) bool { after[n] = n.Parent; return true })
			if !reflect.DeepEqual(beforeParents, after) {
				t.Fatal("render replaced/reparented retained nodes")
			}
			if _, ok := after[owner]; !ok {
				t.Fatal("returned node lost from rendered tree")
			}
			if !reflect.DeepEqual(macroReferencePairs(owner.Properties), want) {
				t.Fatal("render rewrote owner properties")
			}
			for _, n := range identifiers {
				auto, hasAuto := n.Property("autoOP")
				if n == owner {
					if n.TeXClass != mml.TeXClassOp || hasAuto {
						t.Fatal("explicit owner classification changed")
					}
					continue
				}
				if _, ok := n.Property("texClass"); ok {
					t.Fatal("renderer introduced explicit child texClass")
				}
				if len(textContent(n)) > 1 {
					if n.TeXClass != mml.TeXClassOp || !hasAuto || auto != true {
						t.Fatal("multi-letter automatic classification missing")
					}
				} else if n.TeXClass != mml.TeXClassOrd || hasAuto {
					t.Fatal("single-letter automatic classification leaked")
				}
			}
		})
	}
}
