// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package mml

import (
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

// MactionSelectionNumber applies the numeric coercions used by the source
// maction selected getter to supported scalar attribute values.
func MactionSelectionNumber(value any, exists bool) float64 {
	if !exists {
		return math.NaN()
	}
	switch v := value.(type) {
	case nil:
		return 0
	case bool:
		if v {
			return 1
		}
		return 0
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case float64:
		return v
	case string:
		text := strings.TrimFunc(v, mactionNumberSpace)
		if text == "" {
			return 0
		}
		if len(text) > 2 && text[0] == '0' {
			base := 0
			switch text[1] {
			case 'x', 'X':
				base = 16
			case 'b', 'B':
				base = 2
			case 'o', 'O':
				base = 8
			}
			if base != 0 {
				for _, digit := range text[2:] {
					value := int(digit - '0')
					if digit >= 'a' && digit <= 'f' {
						value = int(digit-'a') + 10
					} else if digit >= 'A' && digit <= 'F' {
						value = int(digit-'A') + 10
					}
					if value < 0 || value >= base {
						return math.NaN()
					}
				}
				integer, ok := new(big.Int).SetString(text[2:], base)
				if !ok {
					return math.NaN()
				}
				number, _ := integer.Float64()
				return number
			}
		}
		if mactionDecimalNumber.MatchString(text) {
			// A valid overflowing decimal converts to Infinity. ParseFloat's
			// range error must not turn it into NaN.
			number, _ := strconv.ParseFloat(text, 64)
			return number
		}
	}
	return math.NaN()
}

// StringNumericLiteral accepts decimal/Infinity and unsigned radix literals;
// Go's parser additionally accepts Inf, hex floats and underscores, so validate
// the decimal grammar before conversion. Whitespace is the ECMAScript set,
// including BOM and excluding Unicode NEL (U+0085).
var mactionDecimalNumber = regexp.MustCompile(`^[+-]?(?:Infinity|(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)(?:[eE][+-]?[0-9]+)?)$`)

func mactionNumberSpace(r rune) bool {
	return r == '\t' || r == '\n' || r == '\v' || r == '\f' || r == '\r' || r == ' ' ||
		r == '\u00a0' || r == '\u1680' || (r >= '\u2000' && r <= '\u200a') ||
		r == '\u2028' || r == '\u2029' || r == '\u202f' || r == '\u205f' || r == '\u3000' || r == '\ufeff'
}
