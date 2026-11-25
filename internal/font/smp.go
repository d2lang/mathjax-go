// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
//
// This file is a Go translation and modification of MathJax 3.2.2.
// Source: ts/output/common/FontData.ts.

package font

type smpBases [5]rune

// The zero value belongs to the variants that MathJax gives no bases.
var variantSMP = [variantCount]smpBases{
	indexBold:                {0x1D400, 0x1D41A, 0x1D6A8, 0x1D6C2, 0x1D7CE},
	indexItalic:              {0x1D434, 0x1D44E, 0x1D6E2, 0x1D6FC, 0},
	indexBoldItalic:          {0x1D468, 0x1D482, 0x1D71C, 0x1D736, 0},
	indexScript:              {0x1D49C, 0x1D4B6, 0, 0, 0},
	indexBoldScript:          {0x1D4D0, 0x1D4EA, 0, 0, 0},
	indexFraktur:             {0x1D504, 0x1D51E, 0, 0, 0},
	indexDoubleStruck:        {0x1D538, 0x1D552, 0, 0, 0x1D7D8},
	indexBoldFraktur:         {0x1D56C, 0x1D586, 0, 0, 0},
	indexSansSerif:           {0x1D5A0, 0x1D5BA, 0, 0, 0x1D7E2},
	indexBoldSansSerif:       {0x1D5D4, 0x1D5EE, 0x1D756, 0x1D770, 0x1D7EC},
	indexSansSerifItalic:     {0x1D608, 0x1D622, 0, 0, 0},
	indexSansSerifBoldItalic: {0x1D63C, 0x1D656, 0x1D790, 0x1D7AA, 0},
	indexMonospace:           {0x1D670, 0x1D68A, 0, 0, 0x1D7F6},
}

// smpRemap returns the letterlike symbol that Unicode assigns in place of a
// reserved Mathematical Alphanumeric Symbols code point, or 0.
func smpRemap(smp rune) rune {
	switch smp {
	case 0x1D455:
		return 0x210E
	case 0x1D49D:
		return 0x212C
	case 0x1D4A0:
		return 0x2130
	case 0x1D4A1:
		return 0x2131
	case 0x1D4A3:
		return 0x210B
	case 0x1D4A4:
		return 0x2110
	case 0x1D4A7:
		return 0x2112
	case 0x1D4A8:
		return 0x2133
	case 0x1D4AD:
		return 0x211B
	case 0x1D4BA:
		return 0x212F
	case 0x1D4BC:
		return 0x210A
	case 0x1D4C4:
		return 0x2134
	case 0x1D506:
		return 0x212D
	case 0x1D50B:
		return 0x210C
	case 0x1D50C:
		return 0x2111
	case 0x1D515:
		return 0x211C
	case 0x1D51D:
		return 0x2128
	case 0x1D53A:
		return 0x2102
	case 0x1D53F:
		return 0x210D
	case 0x1D545:
		return 0x2115
	case 0x1D547:
		return 0x2119
	case 0x1D548:
		return 0x211A
	case 0x1D549:
		return 0x211D
	case 0x1D551:
		return 0x2124
	}
	return 0
}

// smpGreekUpper and smpGreekLower return the offset of a Greek symbol that
// lies outside the contiguous alphabet, or 0.
func smpGreekUpper(codepoint rune) rune {
	switch codepoint {
	case 0x2207:
		return 0x19
	case 0x03F4:
		return 0x11
	}
	return 0
}

func smpGreekLower(codepoint rune) rune {
	switch codepoint {
	case 0x03D1:
		return 0x1B
	case 0x03D5:
		return 0x1D
	case 0x03D6:
		return 0x1F
	case 0x03F0:
		return 0x1C
	case 0x03F1:
		return 0x1E
	case 0x03F5:
		return 0x1A
	case 0x2202:
		return 0x19
	}
	return 0
}

func smpCodepoint(variant int, codepoint rune) (rune, bool) {
	base := &variantSMP[variant]
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
	case base[2] != 0 && smpGreekUpper(codepoint) != 0:
		smp = base[2] + smpGreekUpper(codepoint)
	case base[3] != 0 && smpGreekLower(codepoint) != 0:
		smp = base[3] + smpGreekLower(codepoint)
	default:
		if variant == indexBold && (codepoint == 0x3DC || codepoint == 0x3DD) {
			return 0x1D7CA + codepoint - 0x3DC, true
		}
		return 0, false
	}
	if remapped := smpRemap(smp); remapped != 0 {
		smp = remapped
	}
	return smp, true
}
