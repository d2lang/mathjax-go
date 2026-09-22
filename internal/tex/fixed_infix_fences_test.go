// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"fmt"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

func TestFixedInfixFenceOwnership(t *testing.T) {
	for _, c := range []struct{ name, open, close string }{{"choose", "(", ")"}, {"brace", "{", "}"}, {"brack", "[", "]"}} {
		for _, display := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/%t", c.name, display), func(t *testing.T) {
				numerator := token("mi", "x")
				numerator.Attributes.Set("mathvariant", "bold")
				numerator.SetProperty("ownershipWitness", "retained")
				p := &parser{source: "y", state: newParseState(), display: display}
				root, stop, err := p.infixFraction(c.name, []*mml.Node{numerator}, 0, false)
				if err != nil || stop != "" {
					t.Fatal(err, stop)
				}
				if root.Kind != "mrow" || root.Flags.Inferred || root.TeXClass != mml.TeXClassOrd || len(root.Children) != 3 {
					t.Fatal("fixed fence must own an explicit ORD row")
				}
				for key, expected := range map[string]any{"open": c.open, "close": c.close, "texClass": mml.TeXClassOrd} {
					if got, ok := root.Property(key); !ok || got != expected {
						t.Fatal(key, got, expected)
					}
				}
				frac := root.Children[1]
				if frac.Kind != "mfrac" || frac.Parent != root || len(frac.Children) != 2 || frac.Children[0] != numerator || numerator.Parent != frac {
					t.Fatal("original numerator identity or ownership changed")
				}
				if v, ok := frac.Property("withDelims"); !ok || v != true {
					t.Fatal("fraction lost fixed-delimiter padding policy")
				}
				if v, ok := frac.Attributes.GetExplicit("linethickness"); !ok || v != 0 {
					t.Fatal("OverItem thickness must remain numeric zero", v)
				}
				if v, _ := numerator.Property("ownershipWitness"); v != "retained" {
					t.Fatal("numerator properties lost")
				}
				if v, _ := numerator.Attributes.GetExplicit("mathvariant"); v != "bold" {
					t.Fatal("numerator attributes lost")
				}
				for i, character := range map[int]string{0: c.open, 2: c.close} {
					palette := root.Children[i]
					if palette.Kind != "MathChoice" || palette.Parent != root || len(palette.Children) != 4 {
						t.Fatal("fence selected before inherited style")
					}
					for style, atom := range palette.Children {
						class := mml.TeXClassOpen
						if i == 2 {
							class = mml.TeXClassClose
						}
						if atom.Kind != "TeXAtom" || atom.TeXClass != class || atom.Parent != palette {
							t.Fatal("fixed-fence TeXAtom owner changed")
						}
						operators := atom.Find("mo")
						if len(operators) != 1 || textContent(operators[0]) != character {
							t.Fatal("fence content changed")
						}
						size := "1.2em"
						if style == 0 {
							size = "2.047em"
						}
						for _, key := range []string{"minsize", "maxsize"} {
							if v, _ := operators[0].Attributes.GetExplicit(key); v != size {
								t.Fatal("palette size changed", style, key, v)
							}
						}
					}
				}
			})
		}
	}
}

func TestFixedInfixFenceInheritedStyle(t *testing.T) {
	for _, c := range []struct {
		tex, size string
		display   bool
	}{
		{`x\choose y`, "2.047em", true},
		{`x\brace y`, "1.2em", false},
		{`x\brack y`, "2.047em", true},
		{`\displaystyle{x\choose y}`, "2.047em", false},
		{`\textstyle{x\choose y}`, "1.2em", true},
		{`\scriptstyle{x\choose y}`, "1.2em", true},
		{`z_{w_{x\choose y}}`, "1.2em", true},
	} {
		t.Run(c.tex, func(t *testing.T) {
			root, err := NewCompiler().Compile(c.tex, c.display)
			if err != nil {
				t.Fatal(err)
			}
			if len(root.Find("MathChoice")) != 0 {
				t.Fatal("unresolved fixed-fence palette")
			}
			fences := 0
			for _, n := range root.Find("mo") {
				if min, ok := n.Attributes.GetExplicit("minsize"); ok {
					max, _ := n.Attributes.GetExplicit("maxsize")
					if min != c.size || max != c.size {
						t.Fatal("inherited fixed-fence size", min, max)
					}
					fences++
				}
			}
			if fences != 2 {
				t.Fatal("fixed-fence count", fences)
			}
		})
	}
}
