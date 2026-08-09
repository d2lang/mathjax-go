// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
//
// This file is a Go translation and modification of MathJax 3.2.2.
// Sources: ts/output/common/FontData.ts and
// ts/output/common/fonts/tex.ts and fonts/tex/delimiters.ts.

package font

// Direction is MathJax FontData.DIRECTION.
type Direction uint8

const (
	DirectionNone Direction = iota
	DirectionVertical
	DirectionHorizontal
)

// Delimiter is a source-shaped copy of MathJax DelimiterData. Optional-field
// flags distinguish absent data from zero values. Returned slice fields are
// cloned and may be modified by callers.
type Delimiter struct {
	Direction Direction

	Sizes           []float64
	SizeVariants    []int
	SizeChars       []rune
	Stretch         []rune
	StretchVariants []int

	HDW    [3]float64
	HasHDW bool

	Min    float64
	HasMin bool

	Alias    rune
	HasAlias bool

	FullExt    [2]float64
	HasFullExt bool
}

func cloneDelimiter(delimiter Delimiter) Delimiter {
	delimiter.Sizes = append([]float64(nil), delimiter.Sizes...)
	delimiter.SizeVariants = append([]int(nil), delimiter.SizeVariants...)
	delimiter.SizeChars = append([]rune(nil), delimiter.SizeChars...)
	delimiter.Stretch = append([]rune(nil), delimiter.Stretch...)
	delimiter.StretchVariants = append([]int(nil), delimiter.StretchVariants...)
	return delimiter
}

// LookupDelimiter returns the exact MathJax delimiter-table entry.
func LookupDelimiter(codepoint rune) (Delimiter, bool) {
	delimiter, ok := delimiterTable[codepoint]
	if !ok {
		return Delimiter{}, false
	}
	return cloneDelimiter(delimiter), true
}

// DelimiterCount returns the number of entries in the upstream table.
func DelimiterCount() int { return len(delimiterTable) }

var sizeVariants = [...]Variant{Normal, SmallOp, LargeOp, Size3, Size4, TeXVariant}
var stretchVariants = [...]Variant{Size4}

// SizeVariant returns the font variant for the delimiter's i-th fixed size.
func (delimiter Delimiter) SizeVariant(i int) (Variant, bool) {
	if i < 0 {
		return "", false
	}
	index := i
	if len(delimiter.SizeVariants) != 0 {
		if i >= len(delimiter.SizeVariants) {
			return "", false
		}
		index = delimiter.SizeVariants[i]
	}
	if index < 0 || index >= len(sizeVariants) {
		return "", false
	}
	return sizeVariants[index], true
}

// StretchVariant returns the font variant for the delimiter's i-th assembly
// part.
func (delimiter Delimiter) StretchVariant(i int) (Variant, bool) {
	if i < 0 {
		return "", false
	}
	index := 0
	if len(delimiter.StretchVariants) != 0 {
		if i >= len(delimiter.StretchVariants) {
			return "", false
		}
		index = delimiter.StretchVariants[i]
	}
	if index < 0 || index >= len(stretchVariants) {
		return "", false
	}
	return stretchVariants[index], true
}
