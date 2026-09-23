// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package tex

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/ordered"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
)

type multiTextPair struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

type multiTextTree struct {
	Kind       string          `json:"kind"`
	Text       any             `json:"text"`
	Attributes []multiTextPair `json:"attributes"`
	Properties []multiTextPair `json:"properties"`
	Children   []multiTextTree `json:"children"`
}

func multiTextOwnTree(n *mml.Node) multiTextTree {
	pairs := func(m *ordered.Map[any]) []multiTextPair {
		out := []multiTextPair{}
		m.Range(func(name string, value any) bool {
			out = append(out, multiTextPair{name, value})
			return true
		})
		return out
	}
	kind := n.Kind
	if n.Kind == "mrow" && n.Flags.Inferred {
		kind = "inferredMrow"
	}
	var text any
	if n.Kind == "text" {
		text = n.Text
	}
	children := []multiTextTree{}
	for _, child := range n.Children {
		children = append(children, multiTextOwnTree(child))
	}
	return multiTextTree{kind, text, pairs(n.Attributes.Explicit()), pairs(n.Properties), children}
}

// These are constructed renderer inputs, not claimed authored-TeX examples.
// Text-node identity must survive: concatenating the children would disguise
// the missing SVG placement groups that the source output jax supplies.
func TestMultiTextRendererPrimary(t *testing.T) {
	var fixture struct {
		Cases []struct {
			Name, Kind, SVG, SVGSHA256 string
			Parts                      []string
			Attrs                      map[string]any
			Display                    bool
			Before, After              multiTextTree
		}
	}
	data, err := os.ReadFile("../../testdata/multi_text_renderer_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 32 {
		t.Fatalf("expected 32 actual primary cases, got %d", len(fixture.Cases))
	}
	for _, test := range fixture.Cases {
		t.Run(test.Name, func(t *testing.T) {
			texts := make([]*mml.Node, len(test.Parts))
			for i, text := range test.Parts {
				texts[i] = mml.NewText(text)
			}
			token := node(test.Kind, texts...)
			for name, value := range test.Attrs {
				token.Attributes.Set(name, value)
			}
			root := node("math", token)
			if test.Display {
				root.Attributes.Set("display", "block")
			}
			setMathMLInheritance(root, test.Display)
			checkTree := func(want multiTextTree) {
				t.Helper()
				// JSON represents all primary Number values uniformly; no fields
				// or own-property entries are removed from either tree.
				encoded, err := json.Marshal(multiTextOwnTree(root))
				if err != nil {
					t.Fatal(err)
				}
				var got multiTextTree
				if err := json.Unmarshal(encoded, &got); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("complete explicit/own-property tree differs from primary: %s", encoded)
				}
				if len(token.Children) != len(texts) {
					t.Fatal("token text children were replaced")
				}
				for i, original := range texts {
					if token.Children[i] != original || original.Parent != token || original.Text != test.Parts[i] {
						t.Fatalf("text child %d lost its original identity, parent, order or content", i)
					}
				}
			}
			checkTree(test.Before)
			options := pipeline.DefaultOptions()
			options.Display = test.Display
			got, err := svg.NewTypesetter().Typeset(root, options)
			if err != nil {
				t.Fatal(err)
			}
			checkTree(test.After)
			repeated, err := svg.NewTypesetter().Typeset(root, options)
			if err != nil || repeated != got {
				t.Fatalf("repeated rendering differs: %v", err)
			}
			checkTree(test.After)
			hash := sha256.Sum256([]byte(test.SVG))
			if hex.EncodeToString(hash[:]) != test.SVGSHA256 {
				t.Fatal("primary fixture SVG hash mismatch")
			}
			if got != test.SVG {
				t.Fatal("complete SVG differs from unmodified pinned primary output")
			}
		})
	}
}
