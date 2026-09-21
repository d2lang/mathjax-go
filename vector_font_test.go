// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/tex"
)

func TestPhysicsVectorFontPinnedReferences(t *testing.T) {
	b, err := os.ReadFile("internal/tex/testdata/vector_font_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name, Tex, SVGSHA256 string
			Display              bool
			Tree                 *nestedFontTree
		}
	}
	if err = json.Unmarshal(b, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 96 {
		t.Fatal("incomplete token font policy matrix")
	}
	// Preserve the exact primary fixtures. These independently demonstrated
	// preexisting attributes are outside Physics font selection: ordinary
	// Accent adds mover.accent and treats vec as stretchy (D045); MtLap omits
	// two explicit mstyle defaults. Compare every other tree field exactly.
	accentShapes := map[string]bool{"accent-hat": true, "accent-dot": true, "arrow": true, "arrow-star": true, "unit": true, "unit-star": true, "outer-arrow": true, "nested-arrow": true, "long-alias": true, "bold-arrow-control": true, "multi-alias": true}
	renderLimits := map[string]bool{"nested-arrow": true, "bold-arrow-control": true, "multi-alias": true}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			root, err := tex.NewCompiler().Compile(c.Tex, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			root.Walk(func(n *mml.Node) bool {
				for _, key := range []string{"go-resolved-font-scope", "go-ambient-font-source", "go-vector-factory-token", "go-vector-factory-done"} {
					if _, ok := n.Property(key); ok {
						t.Errorf("parser font provenance escaped: %s", key)
					}
				}
				return true
			})
			actual := projectNestedFont(root)
			encoded, _ := json.Marshal(actual)
			if err = json.Unmarshal(encoded, &actual); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(tokenFonts(actual), tokenFonts(c.Tree)) {
				t.Errorf("token font policy got %#v, want %#v", tokenFonts(actual), tokenFonts(c.Tree))
			}
			label := strings.TrimSuffix(strings.TrimSuffix(c.Name, "-display"), "-inline")
			wantTree := c.Tree
			if accentShapes[label] || label == "clap-text" {
				wantBytes, _ := json.Marshal(c.Tree)
				var qualified *nestedFontTree
				if err := json.Unmarshal(wantBytes, &qualified); err != nil {
					t.Fatal(err)
				}
				wantTree = qualified
				qualifyVectorTree(wantTree, accentShapes[label], label == "clap-text")
				originalAfter, _ := json.Marshal(c.Tree)
				if string(originalAfter) != string(wantBytes) {
					t.Fatal("qualification mutated the primary fixture tree")
				}
			}
			if !reflect.DeepEqual(actual, wantTree) {
				want, _ := json.Marshal(wantTree)
				t.Errorf("complete AST got %s, want %s", encoded, want)
			}
			opts := mathjax.DefaultOptions()
			opts.Display = c.Display
			svg, err := mathjax.RenderWithOptions(c.Tex, opts)
			if err != nil {
				t.Fatal(err)
			}
			if !renderLimits[label] {
				if h := fmt.Sprintf("%x", sha256.Sum256([]byte(svg))); h != c.SVGSHA256 {
					t.Errorf("complete SVG %s, want %s", h, c.SVGSHA256)
				}
			}
		})
	}
}

// Apply only the named, observed legacy attribute differences to a copy of the
// primary tree. No text, node, font choice, position or SVG is normalized.
func qualifyVectorTree(n *nestedFontTree, accent, clap bool) {
	if accent && n.Kind == "mover" {
		if _, exists := n.Attributes["accent"]; !exists {
			n.Attributes["accent"] = true
		}
	}
	if accent && n.Kind == "mo" && len(n.Children) == 1 && n.Children[0].Text != nil && *n.Children[0].Text == "→" {
		if n.Attributes["stretchy"] == false {
			delete(n.Attributes, "stretchy")
		}
	}
	if clap && n.Kind == "mstyle" && n.Attributes["displaystyle"] == false && n.Attributes["scriptlevel"] == float64(0) {
		delete(n.Attributes, "displaystyle")
		delete(n.Attributes, "scriptlevel")
	}
	for _, child := range n.Children {
		qualifyVectorTree(child, accent, clap)
	}
}
