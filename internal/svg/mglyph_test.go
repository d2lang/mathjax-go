// Copyright 2018-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package svg

import (
	"math"
	"strings"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

func mglyphNode(attributes map[string]any) *mml.Node {
	node := mml.NewNode("mglyph", nil, nil)
	node.Flags.Token = true
	node.TeXClass = mml.TeXClassOrd
	for name, value := range attributes {
		node.Attributes.Set(name, value)
	}
	return node
}

func TestGlyphImageFrozenSourceShapedSVG(t *testing.T) {
	node := mglyphNode(map[string]any{
		"src": "glyph.png", "alt": "g",
		"width": "2em", "height": "1.5em", "valign": "-.25em",
	})
	got := typesetTrancheWrapper(t, wrapperTrancheRoot(node))
	const want = "b0b5d5076ac028f2d55589b27415ee826dafa186fcc9d06a8d452a278ac7a4bb"
	requireFrozenWrapperHash(t, got, want)
}

func TestGlyphDeprecatedCharacterFrozenSourceShapedSVG(t *testing.T) {
	node := mglyphNode(map[string]any{"fontfamily": "serif", "index": "65"})
	got := typesetTrancheWrapper(t, wrapperTrancheRoot(node))
	const want = "0b2edc662e02ca193ac44209b1b5641d7ce204b8ac50b952c681ec9c01dffadf"
	requireFrozenWrapperHash(t, got, want)
}

func TestGlyphAutoDimensionsAndValign(t *testing.T) {
	node := mglyphNode(map[string]any{
		"src": "g.svg", "alt": "glyph", "width": "auto", "height": "auto", "valign": ".2em",
	})
	w := wrapperTrancheWrapper(node)
	w.computeGlyphBBox(w.bbox)
	w.bboxComputed = true
	for name, pair := range map[string][2]float64{
		"width": {w.bbox.W, 1}, "height": {w.bbox.H, 1.2}, "depth": {w.bbox.D, -.2},
	} {
		if math.Abs(pair[0]-pair[1]) > 1e-12 {
			t.Errorf("%s = %g, want %g", name, pair[0], pair[1])
		}
	}
	root := NewElement("root")
	w.glyphToSVG(root)
	output := root.String()
	for _, fragment := range []string{
		`width="1000" height="1000"`,
		`transform="translate(0 1200) matrix(1 0 0 -1 0 0)"`,
		`preserveAspectRatio="none" aria-label="glyph" href="g.svg"`,
	} {
		if !strings.Contains(output, fragment) {
			t.Errorf("image is missing %q:\n%s", fragment, output)
		}
	}
}
