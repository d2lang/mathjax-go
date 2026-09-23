// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package tex

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
)

// idAnchorBuild uses the same private constructors and inheritance path as the
// reviewed observations. The excluded TeXAtom property/field diagnostic is not
// a regression golden and must not be reintroduced through this helper.
func idAnchorBuild(raw json.RawMessage, display bool) *mml.Node {
	var build func(json.RawMessage) *mml.Node
	build = func(b json.RawMessage) *mml.Node {
		var s struct {
			Kind, Text             string
			Attributes, Properties json.RawMessage
			Children               []json.RawMessage
		}
		if err := json.Unmarshal(b, &s); err != nil {
			panic(err)
		}
		if s.Kind == "text" {
			return mml.NewText(s.Text)
		}
		children := []*mml.Node{}
		for _, child := range s.Children {
			children = append(children, build(child))
		}
		var n *mml.Node
		if s.Kind == "inferredMrow" {
			n = forcedRow(children, true)
		} else {
			n = node(s.Kind, children...)
		}
		set := func(raw json.RawMessage, property bool) {
			if len(raw) == 0 {
				return
			}
			decoder := json.NewDecoder(bytes.NewReader(raw))
			if token, err := decoder.Token(); err != nil || token != json.Delim('{') {
				panic("invalid ordered attribute/property object")
			}
			for decoder.More() {
				key, err := decoder.Token()
				if err != nil {
					panic(err)
				}
				var value any
				if err := decoder.Decode(&value); err != nil {
					panic(err)
				}
				if property {
					if key == "texClass" {
						panic("texClass property diagnostic is excluded from ID-anchor render goldens")
					}
					n.SetProperty(key.(string), value)
				} else {
					n.Attributes.Set(key.(string), value)
				}
			}
		}
		set(s.Attributes, false)
		set(s.Properties, true)
		return n
	}
	root := build(raw)
	setMathMLInheritance(root, display)
	return root
}

func TestIDAnchorRenderPrimary(t *testing.T) {
	var fixture struct {
		MathJaxGitCommit string
		Cases            []struct {
			Name, Group, ExpectedSVG, SVGSHA256 string
			Spec                                struct {
				Name, TeX string
				Display   bool
				Tree      json.RawMessage
			}
		}
	}
	data, err := os.ReadFile("testdata/id_anchor_render_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathJaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 40 {
		t.Fatal("unbound primary ID-anchor fixtures")
	}
	public, constructed := 0, 0
	for _, c := range fixture.Cases {
		if c.Group == "public" {
			public++
		} else if c.Group == "mml" {
			constructed++
		} else {
			t.Fatalf("unknown fixture group %q", c.Group)
		}
		if fmt.Sprintf("%x", sha256.Sum256([]byte(c.ExpectedSVG))) != c.SVGSHA256 {
			t.Fatalf("primary SVG binding differs: %s", c.Name)
		}
	}
	if public != 14 || constructed != 26 {
		t.Fatal("ID-anchor public/constructed fixture coverage changed")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var root *mml.Node
			if c.Group == "mml" {
				if len(c.Spec.Tree) == 0 {
					t.Fatal("missing constructed MathML")
				}
				root = idAnchorBuild(c.Spec.Tree, c.Spec.Display)
			} else {
				var err error
				root, err = NewCompiler().Compile(c.Spec.TeX, c.Spec.Display)
				if err != nil {
					t.Fatal(err)
				}
			}
			options := pipeline.DefaultOptions()
			options.Display = c.Spec.Display
			got, err := svg.NewTypesetter().Typeset(root, options)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.ExpectedSVG {
				t.Errorf("id-anchor whole SVG mismatch: got sha256=%x, want sha256=%s\ngot: %s\nwant: %s", sha256.Sum256([]byte(got)), c.SVGSHA256, got, c.ExpectedSVG)
			}
		})
	}
}
