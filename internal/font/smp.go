// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
//
// This file is a Go translation and modification of MathJax 3.2.2.
// Source: ts/output/common/FontData.ts.

package font

type smpBases [5]rune

var variantSMP = map[Variant]smpBases{
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

var smpRemap = map[rune]rune{
	0x1D455: 0x210E, 0x1D49D: 0x212C, 0x1D4A0: 0x2130, 0x1D4A1: 0x2131,
	0x1D4A3: 0x210B, 0x1D4A4: 0x2110, 0x1D4A7: 0x2112, 0x1D4A8: 0x2133,
	0x1D4AD: 0x211B, 0x1D4BA: 0x212F, 0x1D4BC: 0x210A, 0x1D4C4: 0x2134,
	0x1D506: 0x212D, 0x1D50B: 0x210C, 0x1D50C: 0x2111, 0x1D515: 0x211C,
	0x1D51D: 0x2128, 0x1D53A: 0x2102, 0x1D53F: 0x210D, 0x1D545: 0x2115,
	0x1D547: 0x2119, 0x1D548: 0x211A, 0x1D549: 0x211D, 0x1D551: 0x2124,
}

var smpGreekUpper = map[rune]rune{0x2207: 0x19, 0x03F4: 0x11}
var smpGreekLower = map[rune]rune{
	0x03D1: 0x1B, 0x03D5: 0x1D, 0x03D6: 0x1F, 0x03F0: 0x1C,
	0x03F1: 0x1E, 0x03F5: 0x1A, 0x2202: 0x19,
}

func smpCodepoint(variant Variant, codepoint rune) (rune, bool) {
	base, ok := variantSMP[variant]
	if !ok {
		if variant == Bold && (codepoint == 0x3DC || codepoint == 0x3DD) {
			return 0x1D7CA + codepoint - 0x3DC, true
		}
		return 0, false
	}
	var smp rune
	switch {
	case codepoint >= 'A' && codepoint <= 'Z' && base[0] != 0:
		smp = base[0] + codepoint - 'A'
	case codepoint >= 'a' && codepoint <= 'z' && base[1] != 0:
		smp = base[1] + codepoint - 'a'
	case codepoint >= 0x391 && codepoint <= 0x3A9 && codepoint != 0x3A2 && base[2] != 0:
		smp = base[2] + codepoint - 0x391
	case codepoint >= 0x3B1 && codepoint <= 0x3C9 && base[3] != 0:
		smp = base[3] + codepoint - 0x3B1
	case codepoint >= '0' && codepoint <= '9' && base[4] != 0:
		smp = base[4] + codepoint - '0'
	case base[2] != 0 && smpGreekUpper[codepoint] != 0:
		smp = base[2] + smpGreekUpper[codepoint]
	case base[3] != 0 && smpGreekLower[codepoint] != 0:
		smp = base[3] + smpGreekLower[codepoint]
	default:
		if variant == Bold && (codepoint == 0x3DC || codepoint == 0x3DD) {
			return 0x1D7CA + codepoint - 0x3DC, true
		}
		return 0, false
	}
	if remapped := smpRemap[smp]; remapped != 0 {
		smp = remapped
	}
	return smp, true
}
