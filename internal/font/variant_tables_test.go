// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package font

import "testing"

// The maps below are the tables as they were written before they became
// arrays and switches indexed by variant position. They are kept as the
// reference: every key must still give the same answer.

type namedVariantRule struct {
	inherit Variant
	link    Variant
}

var referenceVariantRules = map[Variant]namedVariantRule{
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

var referenceVariantSMP = map[Variant]smpBases{
	Bold:                {0x1D400, 0x1D41A, 0x1D6A8, 0x1D6C2, 0x1D7CE},
	Italic:              {0x1D434, 0x1D44E, 0x1D6E2, 0x1D6FC, 0},
	BoldItalic:          {0x1D468, 0x1D482, 0x1D71C, 0x1D736, 0},
	Script:              {0x1D49C, 0x1D4B6, 0, 0, 0},
	BoldScript:          {0x1D4D0, 0x1D4EA, 0, 0, 0},
	Fraktur:             {0x1D504, 0x1D51E, 0, 0, 0},
	DoubleStruck:        {0x1D538, 0x1D552, 0, 0, 0x1D7D8},
	BoldFraktur:         {0x1D56C, 0x1D586, 0, 0, 0},
	SansSerif:           {0x1D5A0, 0x1D5BA, 0, 0, 0x1D7E2},
	BoldSansSerif:       {0x1D5D4, 0x1D5EE, 0x1D756, 0x1D770, 0x1D7EC},
	SansSerifItalic:     {0x1D608, 0x1D622, 0, 0, 0},
	SansSerifBoldItalic: {0x1D63C, 0x1D656, 0x1D790, 0x1D7AA, 0},
	Monospace:           {0x1D670, 0x1D68A, 0, 0, 0x1D7F6},
}

var referenceSMPRemap = map[rune]rune{
	0x1D455: 0x210E, 0x1D49D: 0x212C, 0x1D4A0: 0x2130, 0x1D4A1: 0x2131,
	0x1D4A3: 0x210B, 0x1D4A4: 0x2110, 0x1D4A7: 0x2112, 0x1D4A8: 0x2133,
	0x1D4AD: 0x211B, 0x1D4BA: 0x212F, 0x1D4BC: 0x210A, 0x1D4C4: 0x2134,
	0x1D506: 0x212D, 0x1D50B: 0x210C, 0x1D50C: 0x2111, 0x1D515: 0x211C,
	0x1D51D: 0x2128, 0x1D53A: 0x2102, 0x1D53F: 0x210D, 0x1D545: 0x2115,
	0x1D547: 0x2119, 0x1D548: 0x211A, 0x1D549: 0x211D, 0x1D551: 0x2124,
}

var referenceSMPGreekUpper = map[rune]rune{0x2207: 0x19, 0x03F4: 0x11}
var referenceSMPGreekLower = map[rune]rune{
	0x03D1: 0x1B, 0x03D5: 0x1D, 0x03D6: 0x1F, 0x03F0: 0x1C,
	0x03F1: 0x1E, 0x03F5: 0x1A, 0x2202: 0x19,
}

func TestVariantTablesMatchReferenceMaps(t *testing.T) {
	if len(referenceVariantRules) != len(variants) {
		t.Fatalf("%d reference rules for %d variants", len(referenceVariantRules), len(variants))
	}
	name := func(index int8) Variant {
		if index == noVariant {
			return ""
		}
		return variants[index]
	}
	for index, variant := range variants {
		if variantIndex(variant) != index {
			t.Errorf("%s: variantIndex = %d, want %d", variant, variantIndex(variant), index)
		}
		want, ok := referenceVariantRules[variant]
		if !ok {
			t.Errorf("%s: no reference rule", variant)
			continue
		}
		rule := variantRules[index]
		if name(rule.inherit) != want.inherit || name(rule.link) != want.link {
			t.Errorf("%s: inherit %q link %q, want %q %q", variant, name(rule.inherit), name(rule.link), want.inherit, want.link)
		}
		// A missing map entry and the array's zero value both mean no bases.
		if variantSMP[index] != referenceVariantSMP[variant] {
			t.Errorf("%s: alphanumeric bases %X, want %X", variant, variantSMP[index], referenceVariantSMP[variant])
		}
	}
	for variant := range referenceVariantSMP {
		if variantIndex(variant) < 0 {
			t.Errorf("reference bases name unknown variant %q", variant)
		}
	}
}

func TestAlphanumericRemapsMatchReferenceMaps(t *testing.T) {
	for codepoint := rune(-1); codepoint <= 0x110000; codepoint++ {
		if got, want := smpRemap(codepoint), referenceSMPRemap[codepoint]; got != want {
			t.Fatalf("smpRemap(U+%04X) = U+%04X, want U+%04X", codepoint, got, want)
		}
		if got, want := smpGreekUpper(codepoint), referenceSMPGreekUpper[codepoint]; got != want {
			t.Fatalf("smpGreekUpper(U+%04X) = %#x, want %#x", codepoint, got, want)
		}
		if got, want := smpGreekLower(codepoint), referenceSMPGreekLower[codepoint]; got != want {
			t.Fatalf("smpGreekLower(U+%04X) = %#x, want %#x", codepoint, got, want)
		}
	}
}
