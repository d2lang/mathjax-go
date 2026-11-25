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

// Positions of the variants in the variants array, which is the order of
// TeXFont.defaultChars and of the generated glyph tables.
const (
	indexNormal = iota
	indexBold
	indexItalic
	indexBoldItalic
	indexDoubleStruck
	indexFraktur
	indexBoldFraktur
	indexScript
	indexBoldScript
	indexSansSerif
	indexBoldSansSerif
	indexSansSerifItalic
	indexSansSerifBoldItalic
	indexMonospace
	indexSmallOp
	indexLargeOp
	indexSize3
	indexSize4
	indexTeXCalligraphic
	indexTeXBoldCalligraphic
	indexTeXMathItalic
	indexTeXOldStyle
	indexTeXBoldOldStyle
	indexTeXVariant
	variantCount
)

var variants = [variantCount]Variant{
	indexNormal:              Normal,
	indexBold:                Bold,
	indexItalic:              Italic,
	indexBoldItalic:          BoldItalic,
	indexDoubleStruck:        DoubleStruck,
	indexFraktur:             Fraktur,
	indexBoldFraktur:         BoldFraktur,
	indexScript:              Script,
	indexBoldScript:          BoldScript,
	indexSansSerif:           SansSerif,
	indexBoldSansSerif:       BoldSansSerif,
	indexSansSerifItalic:     SansSerifItalic,
	indexSansSerifBoldItalic: SansSerifBoldItalic,
	indexMonospace:           Monospace,
	indexSmallOp:             SmallOp,
	indexLargeOp:             LargeOp,
	indexSize3:               Size3,
	indexSize4:               Size4,
	indexTeXCalligraphic:     TeXCalligraphic,
	indexTeXBoldCalligraphic: TeXBoldCalligraphic,
	indexTeXMathItalic:       TeXMathItalic,
	indexTeXOldStyle:         TeXOldStyle,
	indexTeXBoldOldStyle:     TeXBoldOldStyle,
	indexTeXVariant:          TeXVariant,
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

// noVariant marks a variant without a link, or the end of an inheritance
// chain.
const noVariant = -1

type variantRule struct {
	inherit int8
	link    int8
}

// These rules reproduce FontData.defaultVariants followed by the additional
// CommonTeXFontMixin variants. A link contributes only its own character map;
// inheritance follows the complete fallback chain.
var variantRules = [variantCount]variantRule{
	indexNormal:              {inherit: noVariant, link: noVariant},
	indexBold:                {inherit: indexNormal, link: noVariant},
	indexItalic:              {inherit: indexNormal, link: noVariant},
	indexBoldItalic:          {inherit: indexItalic, link: indexBold},
	indexDoubleStruck:        {inherit: indexBold, link: noVariant},
	indexFraktur:             {inherit: indexNormal, link: noVariant},
	indexBoldFraktur:         {inherit: indexBold, link: indexFraktur},
	indexScript:              {inherit: indexItalic, link: noVariant},
	indexBoldScript:          {inherit: indexBoldItalic, link: indexScript},
	indexSansSerif:           {inherit: indexNormal, link: noVariant},
	indexBoldSansSerif:       {inherit: indexBold, link: indexSansSerif},
	indexSansSerifItalic:     {inherit: indexItalic, link: indexSansSerif},
	indexSansSerifBoldItalic: {inherit: indexBoldItalic, link: indexBoldSansSerif},
	indexMonospace:           {inherit: indexNormal, link: noVariant},
	indexSmallOp:             {inherit: indexNormal, link: noVariant},
	indexLargeOp:             {inherit: indexNormal, link: noVariant},
	indexSize3:               {inherit: indexNormal, link: noVariant},
	indexSize4:               {inherit: indexNormal, link: noVariant},
	indexTeXCalligraphic:     {inherit: indexItalic, link: noVariant},
	indexTeXBoldCalligraphic: {inherit: indexBoldItalic, link: noVariant},
	indexTeXMathItalic:       {inherit: indexItalic, link: noVariant},
	indexTeXOldStyle:         {inherit: indexNormal, link: noVariant},
	indexTeXBoldOldStyle:     {inherit: indexBold, link: noVariant},
	indexTeXVariant:          {inherit: indexNormal, link: noVariant},
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

// location is where an effective lookup ended: at an entry of the generated
// tables, or (glyph < 0) at the code point that a variant substitutes from
// Unicode's Mathematical Alphanumeric Symbols block and that still has to be
// looked up itself.
type location struct {
	variant   int
	codepoint rune
	glyph     int
}

// isRemap reports whether l is a substitution for codepoint, not a glyph.
func (l location) isRemap(codepoint rune) bool { return l.glyph < 0 && l.codepoint != codepoint }

func lookupOwn(variant int, codepoint rune) (location, bool) {
	if glyph := findGlyph(variant, codepoint); glyph >= 0 {
		return location{variant: variant, codepoint: codepoint, glyph: glyph}, true
	}
	if smp, ok := smpCodepoint(variant, codepoint); ok {
		return location{variant: variant, codepoint: smp, glyph: -1}, true
	}
	return location{}, false
}

func lookupEffective(variant int, codepoint rune) (location, bool) {
	for {
		if found, ok := lookupOwn(variant, codepoint); ok {
			return found, true
		}
		rule := variantRules[variant]
		if rule.link != noVariant {
			if found, ok := lookupOwn(int(rule.link), codepoint); ok {
				return found, true
			}
		}
		if rule.inherit == noVariant {
			return location{}, false
		}
		variant = int(rule.inherit)
	}
}

// RemapCodepoint applies the effective variant's mathematical-alphabet mapping
// before glyph lookup, as CommonWrapper.unicodeChars does. Keep the mapped
// character even when no bundled outline exists: it is also the text fallback.
func RemapCodepoint(variant Variant, codepoint rune) rune {
	index := variantIndex(variant)
	if index < 0 {
		return codepoint
	}
	if found, ok := lookupEffective(index, codepoint); ok && found.isRemap(codepoint) {
		return found.codepoint
	}
	return codepoint
}

// locate performs MathJax's effective lookup, including variant link and
// inheritance precedence and Basic Latin/Greek remapping into Unicode's
// Mathematical Alphanumeric Symbols block.
func locate(variant Variant, codepoint rune) (location, bool) {
	index := variantIndex(variant)
	if index < 0 {
		return location{}, false
	}
	found, ok := lookupEffective(index, codepoint)
	if !ok {
		return location{}, false
	}
	// Resolve a substituted code point through the same variant, just as
	// Wrapper.unicodeChars() does before SVG lookup.
	if found.isRemap(codepoint) {
		return lookupEffective(index, found.codepoint)
	}
	return found, true
}

// Lookup performs MathJax's effective lookup, including variant link and
// inheritance precedence and Basic Latin/Greek remapping into Unicode's
// Mathematical Alphanumeric Symbols block.
func Lookup(variant Variant, codepoint rune) (Glyph, bool) {
	found, ok := locate(variant, codepoint)
	if !ok {
		return Glyph{}, false
	}
	glyph := Glyph{Codepoint: found.codepoint, Variant: variants[found.variant]}
	if found.glyph >= 0 {
		glyph.Metrics = glyphMetricsAt(found.glyph)
		glyph.Path, glyph.Content = glyphStringsAt(found.glyph)
	}
	return glyph, true
}

// LookupMetrics is Lookup for layout code, which does not need the outline.
func LookupMetrics(variant Variant, codepoint rune) (Metrics, bool) {
	found, ok := locate(variant, codepoint)
	if !ok || found.glyph < 0 {
		return Metrics{}, ok
	}
	return glyphMetricsAt(found.glyph), true
}

// LookupPath is a convenience wrapper for SVG code. The boolean reports glyph
// presence; an empty path is valid for spaces and Content substitutions.
func LookupPath(variant Variant, codepoint rune) (path, content string, ok bool) {
	found, ok := locate(variant, codepoint)
	if !ok || found.glyph < 0 {
		return "", "", ok
	}
	path, content = glyphStringsAt(found.glyph)
	return path, content, true
}

// ExactGlyphCount returns the number of entries in the upstream source table.
func ExactGlyphCount(variant Variant) int {
	index := variantIndex(variant)
	if index < 0 {
		return 0
	}
	return int(glyphVariantStart[index+1]) - int(glyphVariantStart[index])
}
