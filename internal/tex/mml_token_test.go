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

	"github.com/d2lang/mathjax-go/internal/mml"
)

type mmlTokenTree struct {
	Kind       string          `json:"kind"`
	Text       *string         `json:"text"`
	Attributes map[string]any  `json:"attributes"`
	Children   []*mmlTokenTree `json:"children"`
}

func mmlTokenProject(n *mml.Node) *mmlTokenTree {
	r := &mmlTokenTree{Kind: n.Kind, Attributes: map[string]any{}, Children: []*mmlTokenTree{}}
	if n.Kind == "mrow" && n.Flags.Inferred {
		r.Kind = "inferredMrow"
	}
	if n.Kind == "text" {
		s := n.Text
		r.Text = &s
	}
	for _, name := range n.Attributes.ExplicitNames() {
		r.Attributes[name], _ = n.Attributes.GetExplicit(name)
	}
	for _, child := range n.Children {
		r.Children = append(r.Children, mmlTokenProject(child))
	}
	return r
}

func TestMmlTokenFrozenParser(t *testing.T) {
	data, err := os.ReadFile("testdata/mml_token_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		AssetsSHA256     map[string]string
		Cases            []struct {
			Name, TeX string
			Tree      *mmlTokenTree
			KeptNodes []struct {
				Kind          string
				Attributes    map[string]any
				Movablelimits any
			}
			Error *struct{ ID, Message string }
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 62 {
		t.Fatal("incomplete pinned token matrix")
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || !reflect.DeepEqual(fixture.AssetsSHA256, map[string]string{
		"polyfills.js": "7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01",
		"mathjax.js":   "cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869",
		"setup.js":     "a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881",
	}) {
		t.Fatal("unpinned token oracle assets")
	}
	for _, test := range fixture.Cases {
		t.Run(test.Name, func(t *testing.T) {
			root, err := NewCompiler().Compile(test.TeX, true)
			if err != nil {
				t.Fatal(err)
			}
			got := mmlTokenProject(root)
			// JSON makes the numeric attribute representation identical across
			// the Go and JavaScript decoders, without discarding any tree fields.
			encoded, _ := json.Marshal(got)
			if err = json.Unmarshal(encoded, &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, test.Tree) {
				want, _ := json.Marshal(test.Tree)
				t.Errorf("compiled token tree:\n got %s\nwant %s", encoded, want)
			}
			if test.Error != nil {
				p := &parser{source: strings.TrimPrefix(test.TeX, `\mmlToken`), state: newParseState(), display: true}
				_, err = p.command("mmlToken")
				var failure *Error
				if !errors.As(err, &failure) || failure.ID != test.Error.ID || failure.Message != test.Error.Message {
					t.Errorf("raw parser error = %#v, want %s: %s", err, test.Error.ID, test.Error.Message)
				}
			}
			if strings.HasPrefix(test.TeX, `\mmlToken`) && len(test.KeptNodes) == 1 {
				p := &parser{source: strings.TrimPrefix(test.TeX, `\mmlToken`), state: newParseState(), display: true}
				nodes, err := p.command("mmlToken")
				if err != nil || len(nodes) != 1 {
					t.Fatalf("raw command: %v", err)
				}
				want := test.KeptNodes[0]
				if nodes[0].Kind != want.Kind || !reflect.DeepEqual(mmlTokenProject(nodes[0]).Attributes, want.Attributes) {
					t.Fatalf("pre-cleanup kept attributes = %#v, want %#v", mmlTokenProject(nodes[0]).Attributes, want.Attributes)
				}
				got, _ := nodes[0].Property("movablelimits")
				if got != want.Movablelimits {
					t.Errorf("movablelimits property = %#v, want %#v", got, want.Movablelimits)
				}
			}
		})
	}
}

func TestMmlTokenKeptVariantBoundary(t *testing.T) {
	// Only a variant explicitly retained by MmlToken bypasses the existing
	// font-environment assignment. Empty/unrelated keep lists do not.
	for _, keep := range []string{"", "id", "notmathvariant", "mathvariantx"} {
		n := token("mi", "x")
		n.Attributes.Set("mathvariant", "normal")
		if keep != "" {
			n.Attributes.Set("mjx-keep-attrs", keep)
		}
		applyMathVariant(n, "bold")
		if got, _ := n.Attributes.GetExplicit("mathvariant"); got != "bold" {
			t.Fatalf("ordinary token changed for keep %q: %v", keep, got)
		}
	}
}
