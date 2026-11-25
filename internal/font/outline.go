// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package font

import _ "embed"

// glyphOutlineData holds every distinct SVG outline of the MathJax fonts, one
// record after the other. As text the outlines take 846 kB; these records
// take a third of that and decode to the same text. generate.py writes the
// file and describes the format.
//
//go:embed data_outlines_gen.bin
var glyphOutlineData string

// outlineCommands lists the path commands in the order of their 4-bit codes.
const outlineCommands = "MLHVQTZ"

const (
	outlineH = 2
	outlineV = 3
	outlineQ = 4
	outlineT = 5
	outlineZ = 6
)

// decodeOutline returns the SVG path that record stands for, exactly as the
// MathJax source spells it: a move without its M, then commands whose integer
// arguments are separated by single spaces.
//
// A record is a varint count of the commands after the first move, one 4-bit
// code per command (two to a byte, low half first), and then the numbers of
// the first move and of every command. Each number is stored as its
// difference from the same coordinate of the previous point of the control
// polygon, so the end of a curve is relative to its control point, whether
// that point is written (Q) or implied (T). A difference is one signed byte,
// or 0x80 followed by sixteen bits, low byte first.
func decodeOutline(record string) string {
	if record == "" {
		return ""
	}
	next, count := 0, 0
	for shift := uint(0); ; shift += 7 {
		b := record[next]
		next++
		count |= int(b&0x7F) << shift
		if b < 0x80 {
			break
		}
	}
	codes := record[next : next+(count+1)/2]
	numbers := outlineNumbers{data: record[next+len(codes):]}

	// Most differences take one byte and become three or four of text.
	text := make([]byte, 0, 4*len(record))
	x := numbers.next()
	y := numbers.next()
	controlX, controlY := x, y
	text = appendOutlinePair(text, x, y)
	for i := 0; i < count; i++ {
		code := codes[i>>1] >> (uint(i&1) * 4) & 0xF
		text = append(text, outlineCommands[code])
		switch code {
		case outlineH:
			x += numbers.next()
			controlX, controlY = x, y
			text = appendOutlineNumber(text, x)
		case outlineV:
			y += numbers.next()
			controlX, controlY = x, y
			text = appendOutlineNumber(text, y)
		case outlineZ:
			controlX, controlY = x, y
		case outlineQ:
			controlX = x + numbers.next()
			controlY = y + numbers.next()
			x = controlX + numbers.next()
			y = controlY + numbers.next()
			text = appendOutlinePair(text, controlX, controlY)
			text = append(text, ' ')
			text = appendOutlinePair(text, x, y)
		case outlineT:
			controlX, controlY = 2*x-controlX, 2*y-controlY
			x = controlX + numbers.next()
			y = controlY + numbers.next()
			text = appendOutlinePair(text, x, y)
		default:
			x += numbers.next()
			y += numbers.next()
			controlX, controlY = x, y
			text = appendOutlinePair(text, x, y)
		}
	}
	return string(text)
}

type outlineNumbers struct {
	data string
	at   int
}

func (n *outlineNumbers) next() int {
	value := int(int8(n.data[n.at]))
	n.at++
	if value == -0x80 {
		value = int(int16(uint16(n.data[n.at]) | uint16(n.data[n.at+1])<<8))
		n.at += 2
	}
	return value
}

func appendOutlinePair(text []byte, x, y int) []byte {
	text = appendOutlineNumber(text, x)
	text = append(text, ' ')
	return appendOutlineNumber(text, y)
}

func appendOutlineNumber(text []byte, value int) []byte {
	if value < 0 {
		text = append(text, '-')
		value = -value
	}
	var digits [8]byte
	at := len(digits)
	for {
		at--
		digits[at] = byte('0' + value%10)
		value /= 10
		if value == 0 {
			break
		}
	}
	return append(text, digits[at:]...)
}
