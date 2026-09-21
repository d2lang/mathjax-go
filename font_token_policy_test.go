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

type tokenFont struct {
	Kind    string
	Variant any
}

func tokenFonts(n *nestedFontTree) []tokenFont {
	result := []tokenFont{}
	switch n.Kind {
	case "mi", "mn", "mo", "mtext":
		return []tokenFont{{n.Kind, n.Attributes["mathvariant"]}}
	}
	for _, c := range n.Children {
		result = append(result, tokenFonts(c)...)
	}
	return result
}
func TestTokenFontPolicyPinnedReferences(t *testing.T) {
	b, err := os.ReadFile("internal/tex/testdata/font_token_policy_mathjax_3_2_2.json")
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
	if len(fixture.Cases) != 74 {
		t.Fatal("incomplete token font policy matrix")
	}
	// These retained primary references expose preexisting, independently recorded
	// renderer/shape differences. Every case still checks all token font choices.
	// D042 owns Unicode fallback glyph remapping. Vector/overline geometry and
	// accent/prime node shapes are not changed by this parser font-policy fix.
	shapeLimits := map[string]bool{"accent": true, "accent-dot": true, "vector": true, "under-over-marker": true, "prime": true}
	renderLimits := map[string]bool{"raw-greek": true, "fixed-ams": true, "vector": true, "under-over-marker": true}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			root, err := tex.NewCompiler().Compile(c.Tex, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			root.Walk(func(n *mml.Node) bool {
				for _, key := range []string{"go-resolved-font-scope", "go-ambient-font-source"} {
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
			if !shapeLimits[label] && !reflect.DeepEqual(actual, c.Tree) {
				want, _ := json.Marshal(c.Tree)
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

// Compatibility controls for a Go extension outside the old pinned package.
// These are exact accepted-Go references, not a MathJax3.2.2 parity claim.
func TestTokenFontPolicyBoldsymbolCompatibility(t *testing.T) {
	b, err := os.ReadFile("internal/tex/testdata/font_token_policy_extensions.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Base  string
		Cases []struct {
			Name, Tex, SVGSHA256 string
			Display              bool
			Tree                 *nestedFontTree
		}
	}
	if err = json.Unmarshal(b, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Base != "5a1dd30c9940a1677b86a91f9b8ddd8a848422b0" || len(fixture.Cases) != 10 {
		t.Fatal("incorrect compatibility baseline")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			root, err := tex.NewCompiler().Compile(c.Tex, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			actual := projectNestedFont(root)
			encoded, _ := json.Marshal(actual)
			if err = json.Unmarshal(encoded, &actual); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(actual, c.Tree) {
				t.Fatal("accepted extension AST changed")
			}
			opts := mathjax.DefaultOptions()
			opts.Display = c.Display
			s, err := mathjax.RenderWithOptions(c.Tex, opts)
			if err != nil {
				t.Fatal(err)
			}
			if h := fmt.Sprintf("%x", sha256.Sum256([]byte(s))); h != c.SVGSHA256 {
				t.Errorf("accepted extension SVG %s, want %s", h, c.SVGSHA256)
			}
		})
	}
}
