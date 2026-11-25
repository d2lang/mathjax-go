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
// the caller's own and may be modified.
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

// findDelimiter returns the position of codepoint in the generated delimiter
// tables, or -1.
func findDelimiter(codepoint rune) int {
	low, high := 0, len(delimiterCodepoints)
	for low < high {
		middle := int(uint(low+high) >> 1)
		if rune(delimiterCodepoints[middle]) < codepoint {
			low = middle + 1
		} else {
			high = middle
		}
	}
	if low < len(delimiterCodepoints) && rune(delimiterCodepoints[low]) == codepoint {
		return low
	}
	return -1
}

// Flags in the first word of a delimiter record.
const (
	delimiterHorizontal = 1 << iota
	delimiterHasHDW
	delimiterHasMin
	delimiterHasAlias
	delimiterHasFullExt
)

// decodeDelimiter rebuilds a Delimiter from its record in the generated
// table. A record is a sequence of 16-bit words:
//
//	flags (the constants above)
//	five 3-bit list lengths, lowest bits first: sizes, size variants,
//	    size characters, stretch characters, stretch variants
//	the entries of those five lists
//	HDW (three lengths), min (one), alias (one code point) and fullExt (two
//	    lengths), each only if its flag is set
//
// Lengths are int16 values in the unit of the glyph tables. Every call
// allocates new slices, so callers own what they get; an absent list is nil.
func decodeDelimiter(record []uint16) Delimiter {
	flags, counts := record[0], record[1]
	next := 2
	count := func() int {
		n := int(counts & 7)
		counts >>= 3
		return n
	}
	length := func() float64 {
		value := glyphLength(int16(record[next]))
		next++
		return value
	}
	delimiter := Delimiter{Direction: DirectionVertical}
	if flags&delimiterHorizontal != 0 {
		delimiter.Direction = DirectionHorizontal
	}
	if n := count(); n != 0 {
		delimiter.Sizes = make([]float64, n)
		for i := range delimiter.Sizes {
			delimiter.Sizes[i] = length()
		}
	}
	if n := count(); n != 0 {
		delimiter.SizeVariants = make([]int, n)
		for i := range delimiter.SizeVariants {
			delimiter.SizeVariants[i] = int(record[next])
			next++
		}
	}
	if n := count(); n != 0 {
		delimiter.SizeChars = make([]rune, n)
		for i := range delimiter.SizeChars {
			delimiter.SizeChars[i] = rune(record[next])
			next++
		}
	}
	if n := count(); n != 0 {
		delimiter.Stretch = make([]rune, n)
		for i := range delimiter.Stretch {
			delimiter.Stretch[i] = rune(record[next])
			next++
		}
	}
	if n := count(); n != 0 {
		delimiter.StretchVariants = make([]int, n)
		for i := range delimiter.StretchVariants {
			delimiter.StretchVariants[i] = int(record[next])
			next++
		}
	}
	if flags&delimiterHasHDW != 0 {
		delimiter.HDW, delimiter.HasHDW = [3]float64{length(), length(), length()}, true
	}
	if flags&delimiterHasMin != 0 {
		delimiter.Min, delimiter.HasMin = length(), true
	}
	if flags&delimiterHasAlias != 0 {
		delimiter.Alias, delimiter.HasAlias = rune(record[next]), true
		next++
	}
	if flags&delimiterHasFullExt != 0 {
		delimiter.FullExt, delimiter.HasFullExt = [2]float64{length(), length()}, true
	}
	return delimiter
}

// LookupDelimiter returns the exact MathJax delimiter-table entry.
func LookupDelimiter(codepoint rune) (Delimiter, bool) {
	index := findDelimiter(codepoint)
	if index < 0 {
		return Delimiter{}, false
	}
	return decodeDelimiter(delimiterData[delimiterOffsets[index]:delimiterOffsets[index+1]]), true
}

// DelimiterCount returns the number of entries in the upstream table.
func DelimiterCount() int { return len(delimiterCodepoints) }

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
