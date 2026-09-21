// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"github.com/d2lang/mathjax-go/internal/mml"
	"testing"
)

func TestVectorFactorySelectionAndOrigin(t *testing.T) {
	for _, c := range []struct {
		text         string
		star, accent bool
		want         string
	}{
		{"", false, true, ""}, {"x", false, false, "bold"}, {"Γ", false, false, "bold"}, {"1", false, false, "bold"},
		{"α", false, false, ""}, {"α", true, false, "bold-italic"}, {"ϕ", true, false, ""},
		{"12", false, false, ""}, {"abc", false, false, ""}, {"𝑥", true, false, ""},
		{"+", false, false, ""}, {"^", false, true, "bold"},
	} {
		t.Run(c.text, func(t *testing.T) {
			p := &parser{vectorFactory: true, vectorFont: "bold", vectorStar: c.star}
			if c.star {
				p.vectorFont = "bold-italic"
			}
			n := token("mo", c.text)
			if c.accent {
				n.Attributes.Set("accent", true)
			}
			p.applyVectorFactory(n)
			got, _ := n.Attributes.GetExplicit("mathvariant")
			if c.want == "" {
				if got != nil {
					t.Fatalf("unexpected vector variant %v", got)
				}
			} else if got != c.want {
				t.Fatalf("variant %v, want %s", got, c.want)
			}
			// A later parent factory must not replace the creating environment.
			p.vectorFont = "monospace"
			p.applyVectorFactory(n)
			again, _ := n.Attributes.GetExplicit("mathvariant")
			if again != got {
				t.Fatal("factory reapplied to processed token")
			}
		})
	}
	for _, kind := range []string{"mi", "mn", "mo", "mtext", "ms"} {
		p := &parser{vectorFactory: true, vectorFont: "bold"}
		authored := node(kind, mml.NewText("x"))
		authored.Attributes.Set("mathvariant", "normal")
		p.applyVectorFactory(authored)
		if v, _ := authored.Attributes.GetExplicit("mathvariant"); v != "normal" {
			t.Fatal("node factory was treated as token factory")
		}
		ordinary := token(kind, "x")
		p.activeFont = "normal"
		p.applyVectorFactory(ordinary)
		if _, ok := ordinary.Attributes.GetExplicit("mathvariant"); ok {
			t.Fatal("active inner font did not suppress vector factory")
		}
	}
}

func TestVectorCompileCleanupAndIsolation(t *testing.T) {
	compiler := NewCompiler()
	for _, source := range []string{`\vb{x}`, `\vb{x\vb*{y}z}`, `\vb{\mathrm{x}}`, `\vb{`, `x+\alpha+1`} {
		root, err := compiler.Compile(source, true)
		if err != nil {
			t.Fatal(err)
		}
		root.Walk(func(n *mml.Node) bool {
			for _, key := range []string{vectorFactoryToken, vectorFactoryDone, resolvedFontScope, ambientFontSource} {
				if _, ok := n.Property(key); ok {
					t.Errorf("%q leaked %s", source, key)
				}
			}
			return true
		})
	}
}
