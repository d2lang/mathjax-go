// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package font

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestDecodeOutline(t *testing.T) {
	tests := []struct {
		name   string
		record string
		want   string
	}{
		{"no outline", "", ""},
		{"a move alone", "\x00\x0B\x7F", "11 127"},
		{"lines, with steps of more than a byte", "\x05\x32\x23\x06" + "\x0B\x80\xB3\x00" + "\x80\x0A\x01" + "\x49" + "\xB7" + "\x80\xF6\xFE",
			"11 179H277V252V179H11Z"},
		{"a curve and its mirrored continuation", "\x03\x54\x01" + "\x00\x00" + "\x0A\x00\x00\x0A" + "\x00\xF6" + "\x05\x05",
			"0 0Q10 0 10 10T10 10L15 15"},
		{"a new contour does not mirror the old control point", "\x04\x64\x50" + "\x00\x00" + "\x05\x05\x05\xFB" + "\x14\x00" + "\x03\x04",
			"0 0Q5 5 10 0ZM30 0T33 4"},
	}
	for _, test := range tests {
		if got := decodeOutline(test.record); got != test.want {
			t.Errorf("%s: decoded %q, want %q", test.name, got, test.want)
		}
		if got := encodeOutline(t, test.want); got != test.record {
			t.Errorf("%s: the reference encoder gives %q, want %q", test.name, got, test.record)
		}
	}
}

// outlineText is the form of every outline in the MathJax source: a move
// without its letter, then commands with single spaces between integers.
var outlineText = regexp.MustCompile(`^(0|-?[1-9][0-9]*) (0|-?[1-9][0-9]*)([MLHVQTZ]((0|-?[1-9][0-9]*)( (0|-?[1-9][0-9]*))*)?)*$`)

// TestGeneratedTableHash pins the decoded text of every outline. This test
// ties the stored bytes to that text in the other direction as well, so a
// record is the one encoding of its outline.
func TestOutlineRecordsAreCanonical(t *testing.T) {
	distinct := make(map[string]int)
	for index := 0; index < glyphFirstContent; index++ {
		record := glyphOutlineData[glyphOutlineOffsets[index]:glyphOutlineOffsets[index+1]]
		text := decodeOutline(record)
		if index == 0 {
			if text != "" {
				t.Fatalf("outline 0 is %q, want none", text)
			}
			continue
		}
		if !outlineText.MatchString(text) {
			t.Fatalf("outline %d has an unexpected form: %q", index, text)
		}
		if again := encodeOutline(t, text); again != record {
			t.Fatalf("outline %d: record %q is not the encoding %q of %q", index, record, again, text)
		}
		if earlier, ok := distinct[text]; ok {
			t.Fatalf("outlines %d and %d are the same", earlier, index)
		}
		distinct[text] = index
	}
}

// encodeOutline is an independent writer of the record format, for tests.
func encodeOutline(t *testing.T, text string) string {
	t.Helper()
	if text == "" {
		return ""
	}
	type command struct {
		letter  byte
		numbers []int
	}
	var commands []command
	for rest := "M" + text; rest != ""; {
		end := 1
		for end < len(rest) && strings.IndexByte(outlineCommands, rest[end]) < 0 {
			end++
		}
		current := command{letter: rest[0]}
		for _, field := range strings.Fields(rest[1:end]) {
			value, err := strconv.Atoi(field)
			if err != nil {
				t.Fatalf("bad number %q in %q", field, text)
			}
			current.numbers = append(current.numbers, value)
		}
		commands = append(commands, current)
		rest = rest[end:]
	}
	var codes []byte
	var differences []int
	x, y, controlX, controlY := 0, 0, 0, 0
	for i, current := range commands {
		if i > 0 {
			codes = append(codes, byte(strings.IndexByte(outlineCommands, current.letter)))
		}
		n := current.numbers
		switch current.letter {
		case 'H':
			differences = append(differences, n[0]-x)
			x = n[0]
			controlX, controlY = x, y
		case 'V':
			differences = append(differences, n[0]-y)
			y = n[0]
			controlX, controlY = x, y
		case 'Z':
			controlX, controlY = x, y
		case 'Q':
			differences = append(differences, n[0]-x, n[1]-y, n[2]-n[0], n[3]-n[1])
			controlX, controlY, x, y = n[0], n[1], n[2], n[3]
		case 'T':
			controlX, controlY = 2*x-controlX, 2*y-controlY
			differences = append(differences, n[0]-controlX, n[1]-controlY)
			x, y = n[0], n[1]
		default:
			differences = append(differences, n[0]-x, n[1]-y)
			x, y = n[0], n[1]
			controlX, controlY = x, y
		}
	}
	var record []byte
	count := len(codes)
	for ; count >= 0x80; count >>= 7 {
		record = append(record, byte(count)|0x80)
	}
	record = append(record, byte(count))
	for i := 0; i < len(codes); i += 2 {
		packed := codes[i]
		if i+1 < len(codes) {
			packed |= codes[i+1] << 4
		}
		record = append(record, packed)
	}
	for _, difference := range differences {
		if difference >= -127 && difference <= 127 {
			record = append(record, byte(int8(difference)))
		} else {
			record = append(record, 0x80, byte(difference), byte(difference>>8))
		}
	}
	return string(record)
}
