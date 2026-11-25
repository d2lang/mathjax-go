// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
//
// This file is a Go translation and modification of MathJax 3.2.2.
// Sources: ts/output/{common,svg}/FontData.ts,
// ts/output/{common,svg}/fonts/tex.ts, and the per-variant TeX font data in
// ts/output/{common,svg}/fonts/tex/*.ts.

// Package font provides the MathJax 3.2.2 TeX font metrics, SVG outlines,
// variant fallback rules, and stretchy-delimiter data.
package font

//go:generate python3 generate.py

// Variant is a MathJax TeX font variant name.  The string values intentionally
// match TeXFont.defaultChars in MathJax 3.2.2.
type Variant string

const (
	Normal              Variant = "normal"
	Bold                Variant = "bold"
	Italic              Variant = "italic"
	BoldItalic          Variant = "bold-italic"
	DoubleStruck        Variant = "double-struck"
	Fraktur             Variant = "fraktur"
	BoldFraktur         Variant = "bold-fraktur"
	Script              Variant = "script"
	BoldScript          Variant = "bold-script"
	SansSerif           Variant = "sans-serif"
	BoldSansSerif       Variant = "bold-sans-serif"
	SansSerifItalic     Variant = "sans-serif-italic"
	SansSerifBoldItalic Variant = "sans-serif-bold-italic"
	Monospace           Variant = "monospace"
	SmallOp             Variant = "-smallop"
	LargeOp             Variant = "-largeop"
	Size3               Variant = "-size3"
	Size4               Variant = "-size4"
	TeXCalligraphic     Variant = "-tex-calligraphic"
	TeXBoldCalligraphic Variant = "-tex-bold-calligraphic"
	TeXMathItalic       Variant = "-tex-mathit"
	TeXOldStyle         Variant = "-tex-oldstyle"
	TeXBoldOldStyle     Variant = "-tex-bold-oldstyle"
	TeXVariant          Variant = "-tex-variant"
)

var variants = [...]Variant{
	Normal, Bold, Italic, BoldItalic, DoubleStruck, Fraktur, BoldFraktur,
	Script, BoldScript, SansSerif, BoldSansSerif, SansSerifItalic,
	SansSerifBoldItalic, Monospace, SmallOp, LargeOp, Size3, Size4,
	TeXCalligraphic, TeXBoldCalligraphic, TeXMathItalic, TeXOldStyle,
	TeXBoldOldStyle, TeXVariant,
}

// Variants returns the canonical variants in MathJax source order.
func Variants() []Variant { return append([]Variant(nil), variants[:]...) }

// variantIndex returns the position of variant in variants, which is also its
// position in the generated glyph tables, or -1 for an unknown variant.
func variantIndex(variant Variant) int {
	for i := range variants {
		if variants[i] == variant {
			return i
		}
	}
	return -1
}

// ValidVariant reports whether variant is a MathJax 3.2.2 TeX variant.
func ValidVariant(variant Variant) bool { return variantIndex(variant) >= 0 }

// Metrics is MathJax CharData: height, depth, width, and optional character
// options.  Presence flags preserve the distinction between an omitted option
// and an explicitly supplied zero.
type Metrics struct {
	Height float64
	Depth  float64
	Width  float64

	ItalicCorrection    float64
	Skew                float64
	HasItalicCorrection bool
	HasSkew             bool
}

// Glyph is a resolved TeX glyph. Codepoint and Variant identify the exact
// generated table supplying Metrics and Path. Content is MathJax's optional
// SVG character substitution (the `c` SVGCharOption); when non-empty, callers
// render its runes instead of Path.
type Glyph struct {
	Codepoint rune
	Variant   Variant
	Metrics   Metrics
	Path      string
	Content   string
}

type variantRule struct {
	inherit Variant
	link    Variant
}

// These rules reproduce FontData.defaultVariants followed by the additional
// CommonTeXFontMixin variants. A link contributes only its own character map;
// inheritance follows the complete fallback chain.
var variantRules = map[Variant]variantRule{
	Normal:              {},
	Bold:                {inherit: Normal},
	Italic:              {inherit: Normal},
	BoldItalic:          {inherit: Italic, link: Bold},
	DoubleStruck:        {inherit: Bold},
	Fraktur:             {inherit: Normal},
	BoldFraktur:         {inherit: Bold, link: Fraktur},
	Script:              {inherit: Italic},
	BoldScript:          {inherit: BoldItalic, link: Script},
	SansSerif:           {inherit: Normal},
	BoldSansSerif:       {inherit: Bold, link: SansSerif},
	SansSerifItalic:     {inherit: Italic, link: SansSerif},
	SansSerifBoldItalic: {inherit: BoldItalic, link: BoldSansSerif},
	Monospace:           {inherit: Normal},
	SmallOp:             {inherit: Normal},
	LargeOp:             {inherit: Normal},
	Size3:               {inherit: Normal},
	Size4:               {inherit: Normal},
	TeXCalligraphic:     {inherit: Italic},
	TeXBoldCalligraphic: {inherit: BoldItalic},
	TeXMathItalic:       {inherit: Italic},
	TeXOldStyle:         {inherit: Normal},
	TeXBoldOldStyle:     {inherit: Bold},
	TeXVariant:          {inherit: Normal},
}

// LookupExact returns only a source-table entry, without variant inheritance
// or mathematical-alphanumeric remapping.
func LookupExact(variant Variant, codepoint rune) (Glyph, bool) {
	index := variantIndex(variant)
	if index < 0 {
		return Glyph{}, false
	}
	glyph := findGlyph(index, codepoint)
	if glyph < 0 {
		return Glyph{}, false
	}
	path, content := glyphStringsAt(glyph)
	return Glyph{Codepoint: codepoint, Variant: variant, Metrics: glyphMetricsAt(glyph), Path: path, Content: content}, true
}

// The generated tables (data_glyphs_gen.go) hold no pointers: the code points
// of each variant are sorted for binary search, lengths are integers, and all
// outlines share one string. See generate.py for the reasons.

// findGlyph returns the table index of codepoint in the variant with the
// given index, or -1 if that variant's source table has no such entry.
func findGlyph(variant int, codepoint rune) int {
	low, end := int(glyphVariantStart[variant]), int(glyphVariantStart[variant+1])
	high := end
	for low < high {
		middle := int(uint(low+high) >> 1)
		if rune(glyphCodepoints[middle]) < codepoint {
			low = middle + 1
		} else {
			high = middle
		}
	}
	if low < end && rune(glyphCodepoints[low]) == codepoint {
		return low
	}
	return -1
}

// glyphLengthUnit is the number of table units in one em. Dividing by it
// gives the same float64 as the decimal literal in the MathJax source, which
// has at most four decimals: both are the correctly rounded quotient.
const glyphLengthUnit = 10000

func glyphLength(units int16) float64 { return float64(units) / glyphLengthUnit }

func glyphMetricsAt(glyph int) Metrics {
	units := glyphMetrics[5*glyph : 5*glyph+5]
	metrics := Metrics{
		Height: glyphLength(units[0]),
		Depth:  glyphLength(units[1]),
		Width:  glyphLength(units[2]),
	}
	if units[3] != na {
		metrics.ItalicCorrection, metrics.HasItalicCorrection = glyphLength(units[3]), true
	}
	if units[4] != na {
		metrics.Skew, metrics.HasSkew = glyphLength(units[4]), true
	}
	return metrics
}

func glyphStringsAt(glyph int) (path, content string) {
	index := int(glyphStrings[glyph])
	text := glyphStringData[glyphStringOffsets[index]:glyphStringOffsets[index+1]]
	if index >= glyphFirstContent {
		return "", text
	}
	return text, ""
}

func lookupOwn(variant Variant, codepoint rune) (Glyph, bool) {
	if glyph, ok := LookupExact(variant, codepoint); ok {
		return glyph, true
	}
	if smp, ok := smpCodepoint(variant, codepoint); ok {
		return Glyph{Codepoint: smp, Variant: variant}, true
	}
	return Glyph{}, false
}

func lookupEffective(variant Variant, codepoint rune) (Glyph, bool) {
	if glyph, ok := lookupOwn(variant, codepoint); ok {
		return glyph, true
	}
	rule, ok := variantRules[variant]
	if !ok {
		return Glyph{}, false
	}
	if rule.link != "" {
		if glyph, ok := lookupOwn(rule.link, codepoint); ok {
			return glyph, true
		}
	}
	if rule.inherit != "" {
		return lookupEffective(rule.inherit, codepoint)
	}
	return Glyph{}, false
}

// RemapCodepoint applies the effective variant's mathematical-alphabet mapping
// before glyph lookup, as CommonWrapper.unicodeChars does. Keep the mapped
// character even when no bundled outline exists: it is also the text fallback.
func RemapCodepoint(variant Variant, codepoint rune) rune {
	glyph, ok := lookupEffective(variant, codepoint)
	if ok && glyph.Codepoint != codepoint && glyph.Path == "" && glyph.Content == "" && glyph.Metrics == (Metrics{}) {
		return glyph.Codepoint
	}
	return codepoint
}

// Lookup performs MathJax's effective lookup, including variant link and
// inheritance precedence and Basic Latin/Greek remapping into Unicode's
// Mathematical Alphanumeric Symbols block.
func Lookup(variant Variant, codepoint rune) (Glyph, bool) {
	if !ValidVariant(variant) {
		return Glyph{}, false
	}
	glyph, ok := lookupEffective(variant, codepoint)
	if !ok {
		return Glyph{}, false
	}
	// A glyph whose metrics are all zero and whose path/content are empty is a
	// synthetic SMP remap marker. Resolve that codepoint through the same
	// variant, just as Wrapper.unicodeChars() does before SVG lookup.
	if glyph.Codepoint != codepoint && glyph.Path == "" && glyph.Content == "" && glyph.Metrics == (Metrics{}) {
		return lookupEffective(variant, glyph.Codepoint)
	}
	return glyph, true
}

// LookupMetrics is a convenience wrapper for layout code.
func LookupMetrics(variant Variant, codepoint rune) (Metrics, bool) {
	glyph, ok := Lookup(variant, codepoint)
	return glyph.Metrics, ok
}

// LookupPath is a convenience wrapper for SVG code. The boolean reports glyph
// presence; an empty path is valid for spaces and Content substitutions.
func LookupPath(variant Variant, codepoint rune) (path, content string, ok bool) {
	glyph, ok := Lookup(variant, codepoint)
	if !ok {
		return "", "", false
	}
	return glyph.Path, glyph.Content, true
}

// ExactGlyphCount returns the number of entries in the upstream source table.
func ExactGlyphCount(variant Variant) int {
	index := variantIndex(variant)
	if index < 0 {
		return 0
	}
	return int(glyphVariantStart[index+1]) - int(glyphVariantStart[index])
}
