// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package svg

import (
	"strings"
	"testing"

	"github.com/d2lang/mathjax-go/internal/font"
	"github.com/d2lang/mathjax-go/internal/mml"
)

func TestBacktickVariantPrecedence(t *testing.T) {
	for _, c := range []struct {
		name                                string
		pseudo, display, large, variantForm bool
		explicit, inherited, family         string
		want                                font.Variant
	}{
		{"script", false, false, false, false, "", "", "", font.TeXVariant},
		{"inherited-font", false, false, false, false, "", "bold", "", font.TeXVariant},
		{"authored-family", false, false, false, false, "", "", "serif", font.TeXVariant},
		{"explicit-bold", false, false, false, false, "bold", "", "", font.Bold},
		{"explicit-normal", false, false, false, false, "normal", "", "", font.Normal},
		{"ordinary-inherited", true, false, false, false, "", "bold", "", font.Bold},
		{"largeop", false, true, true, true, "bold", "", "serif", font.LargeOp},
		{"smallop", false, false, true, true, "bold", "", "serif", font.SmallOp},
	} {
		t.Run(c.name, func(t *testing.T) {
			n := wrapperTrancheToken("mo", "‘", mml.TeXClassOrd)
			n.SetProperty("pseudoscript", c.pseudo)
			n.SetProperty("primes", "‵")
			n.SetProperty("variantForm", c.variantForm)
			n.Attributes.SetInherited("displaystyle", c.display)
			n.Attributes.Set("largeop", c.large)
			if c.explicit != "" {
				n.Attributes.Set("mathvariant", c.explicit)
			}
			if c.inherited != "" {
				n.Attributes.SetInherited("mathvariant", c.inherited)
			}
			if c.family != "" {
				n.Attributes.Set("fontfamily", c.family)
			}
			w := wrapperTrancheWrapper(n)
			if w.variant != c.want || w.explicitFont {
				t.Fatalf("variant=%q explicitFont=%v, want %q", w.variant, w.explicitFont, c.want)
			}
		})
	}
}

func TestBacktickMeasureAndPaintRemap(t *testing.T) {
	for _, explicitFont := range []bool{false, true} {
		n := wrapperTrancheToken("mo", "‘", mml.TeXClassOrd)
		n.SetProperty("pseudoscript", true)
		n.SetProperty("primes", "‵")
		if explicitFont {
			n.Attributes.Set("fontfamily", "serif")
		}
		w := wrapperTrancheWrapper(n)
		container := NewElement("g")
		box := *w.getBBox()
		w.toSVG(container)
		if explicitFont {
			if box.W != .6 || !strings.Contains(container.String(), ">‘</text>") {
				t.Fatal("explicit font must measure and emit raw text")
			}
		} else {
			glyph, ok := font.Lookup(font.Normal, '‵')
			if !ok || box.W != glyph.Metrics.Width || !strings.Contains(container.String(), `data-c="2035"`) || strings.Contains(container.String(), `data-c="2018"`) {
				t.Fatal("measured and emitted prime glyph disagree")
			}
		}
		if n.Children[0].Text != "‘" {
			t.Fatal("rendering changed raw text")
		}
	}
	// A selected stretch character bypasses the primes property, as it does
	// CommonMo.remapChars in CommonTextNode.
	n := wrapperTrancheToken("mo", "‘", mml.TeXClassOrd)
	n.SetProperty("primes", "‵")
	w := wrapperTrancheWrapper(n)
	w.stretchGlyph = '|'
	box := *w.children[0].getBBox()
	container := NewElement("g")
	w.children[0].toSVG(container)
	glyph, ok := font.Lookup(font.Normal, '|')
	if !ok || box.W != glyph.Metrics.Width || !strings.Contains(container.String(), `data-c="7C"`) || strings.Contains(container.String(), `data-c="2035"`) {
		t.Fatal("selected stretch character must determine both measurement and paint")
	}
}
