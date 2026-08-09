// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package font

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math"
	"slices"
	"sort"
	"testing"
)

func TestExactTableCounts(t *testing.T) {
	expected := map[Variant]int{
		Normal: 1298, Bold: 255, Italic: 60, BoldItalic: 6,
		DoubleStruck: 0, Fraktur: 40, BoldFraktur: 39, Script: 0,
		BoldScript: 0, SansSerif: 72, BoldSansSerif: 48,
		SansSerifItalic: 82, SansSerifBoldItalic: 2, Monospace: 78,
		SmallOp: 50, LargeOp: 50, Size3: 24, Size4: 53,
		TeXCalligraphic: 26, TeXBoldCalligraphic: 28, TeXMathItalic: 52,
		TeXOldStyle: 36, TeXBoldOldStyle: 36, TeXVariant: 38,
	}
	if len(expected) != len(Variants()) {
		t.Fatalf("expected %d variants, got %d", len(expected), len(Variants()))
	}
	total := 0
	for _, variant := range Variants() {
		got := ExactGlyphCount(variant)
		if got != expected[variant] {
			t.Errorf("%s: got %d exact glyphs, want %d", variant, got, expected[variant])
		}
		total += got
	}
	if total != 2373 {
		t.Errorf("got %d exact glyphs, want 2373", total)
	}
	if got := DelimiterCount(); got != 119 {
		t.Errorf("got %d delimiters, want 119", got)
	}
}

func TestUpstreamSourceHashes(t *testing.T) {
	if CommonSourceSHA256 != "6270255b8b5b5c21e09a62261be09d677860135be814d0d031f09b3b4c6e8c28" {
		t.Errorf("unexpected common source hash %s", CommonSourceSHA256)
	}
	if SVGSourceSHA256 != "bfc1e17685cb4d3a30eb3ee674f15c6b95519a745030fc3b4d3126cda98ba888" {
		t.Errorf("unexpected SVG source hash %s", SVGSourceSHA256)
	}
	if DelimiterSourceSHA256 != "440c69fd81371a8ae766accdd2f0ffe560520d0b1c757d8ed4874a454301e174" {
		t.Errorf("unexpected delimiter source hash %s", DelimiterSourceSHA256)
	}
}

func TestSourceShapedGlyphLookups(t *testing.T) {
	minus, ok := LookupExact(Normal, 0x2D)
	if !ok {
		t.Fatal("normal U+002D is missing")
	}
	if minus.Metrics != (Metrics{Height: 0.252, Depth: -0.179, Width: 0.333}) {
		t.Errorf("normal U+002D metrics = %#v", minus.Metrics)
	}
	if minus.Path != "11 179V252H277V179H11" || minus.Content != "" {
		t.Errorf("normal U+002D SVG data = path %q, content %q", minus.Path, minus.Content)
	}

	prime, ok := LookupExact(Normal, 0x2033)
	if !ok || prime.Path != "" || prime.Content != "\u2032\u2032" {
		t.Errorf("normal U+2033 SVG substitution = %#v, %v", prime, ok)
	}

	scriptH, ok := LookupExact(Normal, 0x210B)
	if !ok {
		t.Fatal("normal U+210B is missing")
	}
	want := Metrics{
		Height: 0.717, Depth: 0.036, Width: 0.969,
		ItalicCorrection: 0.272, HasItalicCorrection: true,
		Skew: 0.333, HasSkew: true,
	}
	if scriptH.Metrics != want {
		t.Errorf("normal U+210B metrics = %#v, want %#v", scriptH.Metrics, want)
	}
}

func TestResolvedVariantLookups(t *testing.T) {
	tests := []struct {
		variant   Variant
		input     rune
		codepoint rune
		source    Variant
	}{
		{BoldItalic, 'A', 0x1D468, Normal},
		{DoubleStruck, 'C', 0x2102, Normal},
		{Script, 'B', 0x212C, Normal},
		{SansSerifItalic, '1', '1', SansSerifItalic}, // Exact entries take precedence over SMP remapping.
		{SansSerifBoldItalic, '1', 0x1D7ED, Normal},  // SMP inherited through the linked bold-sans-serif map.
		{Bold, '!', '!', Bold},
		{BoldItalic, '!', '!', Bold}, // linked bold map precedes inherited italic.
		{SmallOp, 'A', 'A', Normal},
	}
	for _, test := range tests {
		glyph, ok := Lookup(test.variant, test.input)
		if !ok {
			t.Errorf("Lookup(%q, U+%04X) failed", test.variant, test.input)
			continue
		}
		if glyph.Codepoint != test.codepoint || glyph.Variant != test.source {
			t.Errorf("Lookup(%q, U+%04X) source = %q U+%04X, want %q U+%04X", test.variant, test.input, glyph.Variant, glyph.Codepoint, test.source, test.codepoint)
		}
		if exact, ok := LookupExact(glyph.Variant, glyph.Codepoint); !ok || exact.Metrics != glyph.Metrics || exact.Path != glyph.Path || exact.Content != glyph.Content {
			t.Errorf("Lookup(%q, U+%04X) did not resolve to an exact table entry", test.variant, test.input)
		}
	}
	if _, ok := Lookup("not-a-variant", 'x'); ok {
		t.Error("unknown variant unexpectedly resolved")
	}
}

func TestSourceShapedDelimiters(t *testing.T) {
	paren, ok := LookupDelimiter('(')
	if !ok {
		t.Fatal("left parenthesis delimiter is missing")
	}
	if paren.Direction != DirectionVertical || !slices.Equal(paren.Sizes, []float64{1, 1.2, 1.8, 2.4, 3}) ||
		!slices.Equal(paren.Stretch, []rune{0x239B, 0x239C, 0x239D}) || !paren.HasHDW || paren.HDW != [3]float64{0.85, 0.349, 0.875} {
		t.Errorf("left parenthesis delimiter = %#v", paren)
	}

	surd, ok := LookupDelimiter(0x221A)
	if !ok || !surd.HasFullExt || surd.FullExt != [2]float64{0.65, 2.3} || surd.HDW != [3]float64{0.85, 0.35, 1.056} {
		t.Errorf("square-root delimiter = %#v, %v", surd, ok)
	}
	overbrace, ok := LookupDelimiter(0x23DE)
	if !ok || overbrace.Direction != DirectionHorizontal || !overbrace.HasMin || overbrace.Min != 1.8 ||
		!overbrace.HasAlias || overbrace.Alias != 0x23DE || !slices.Equal(overbrace.Stretch, []rune{0xE150, 0xE154, 0xE151, 0xE155}) {
		t.Errorf("overbrace delimiter = %#v, %v", overbrace, ok)
	}

	// Lookup returns owned slices rather than exposing the generated table.
	paren.Sizes[0] = 99
	again, _ := LookupDelimiter('(')
	if again.Sizes[0] != 1 {
		t.Fatal("delimiter table was mutated through a returned slice")
	}
}

func writeUint64(hash interface{ Write([]byte) (int, error) }, value uint64) {
	var data [8]byte
	binary.BigEndian.PutUint64(data[:], value)
	_, _ = hash.Write(data[:])
}

func writeBool(hash interface{ Write([]byte) (int, error) }, value bool) {
	if value {
		_, _ = hash.Write([]byte{1})
	} else {
		_, _ = hash.Write([]byte{0})
	}
}

func writeString(hash interface{ Write([]byte) (int, error) }, value string) {
	writeUint64(hash, uint64(len(value)))
	_, _ = hash.Write([]byte(value))
}

func writeFloats(hash interface{ Write([]byte) (int, error) }, values []float64) {
	writeUint64(hash, uint64(len(values)))
	for _, value := range values {
		writeUint64(hash, math.Float64bits(value))
	}
}

func writeInts(hash interface{ Write([]byte) (int, error) }, values []int) {
	writeUint64(hash, uint64(len(values)))
	for _, value := range values {
		writeUint64(hash, uint64(value))
	}
}

func writeRunes(hash interface{ Write([]byte) (int, error) }, values []rune) {
	writeUint64(hash, uint64(len(values)))
	for _, value := range values {
		writeUint64(hash, uint64(value))
	}
}

func generatedTableHash() string {
	hash := sha256.New()
	for _, variant := range Variants() {
		writeString(hash, string(variant))
		table := glyphTables[variant]
		codepoints := make([]rune, 0, len(table))
		for codepoint := range table {
			codepoints = append(codepoints, codepoint)
		}
		sort.Slice(codepoints, func(i, j int) bool { return codepoints[i] < codepoints[j] })
		writeUint64(hash, uint64(len(codepoints)))
		for _, codepoint := range codepoints {
			glyph := table[codepoint]
			writeUint64(hash, uint64(codepoint))
			for _, value := range [...]float64{glyph.Metrics.Height, glyph.Metrics.Depth, glyph.Metrics.Width, glyph.Metrics.ItalicCorrection, glyph.Metrics.Skew} {
				writeUint64(hash, math.Float64bits(value))
			}
			writeBool(hash, glyph.Metrics.HasItalicCorrection)
			writeBool(hash, glyph.Metrics.HasSkew)
			writeString(hash, glyph.Path)
			writeString(hash, glyph.Content)
		}
	}
	delimiters := make([]rune, 0, len(delimiterTable))
	for codepoint := range delimiterTable {
		delimiters = append(delimiters, codepoint)
	}
	sort.Slice(delimiters, func(i, j int) bool { return delimiters[i] < delimiters[j] })
	writeUint64(hash, uint64(len(delimiters)))
	for _, codepoint := range delimiters {
		delimiter := delimiterTable[codepoint]
		writeUint64(hash, uint64(codepoint))
		writeUint64(hash, uint64(delimiter.Direction))
		writeFloats(hash, delimiter.Sizes)
		writeInts(hash, delimiter.SizeVariants)
		writeRunes(hash, delimiter.SizeChars)
		writeRunes(hash, delimiter.Stretch)
		writeInts(hash, delimiter.StretchVariants)
		for _, value := range delimiter.HDW {
			writeUint64(hash, math.Float64bits(value))
		}
		writeBool(hash, delimiter.HasHDW)
		writeUint64(hash, math.Float64bits(delimiter.Min))
		writeBool(hash, delimiter.HasMin)
		writeUint64(hash, uint64(delimiter.Alias))
		writeBool(hash, delimiter.HasAlias)
		for _, value := range delimiter.FullExt {
			writeUint64(hash, math.Float64bits(value))
		}
		writeBool(hash, delimiter.HasFullExt)
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func TestGeneratedTableHash(t *testing.T) {
	const expected = "900a71cf3d7d3c1710b3d331efe04b9cbca20cd6020ef0c26b2eced3333d9a71"
	if got := generatedTableHash(); got != expected {
		t.Fatalf("generated table hash = %s, want %s", got, expected)
	}
}
