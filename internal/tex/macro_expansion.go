// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Based on ParseUtil.addArgs and ParseUtil.substituteArgs (MathJax 3.2.2).
package tex

import (
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

const maxMacroBuffer = 5120

// macroAddArgs preserves a control word at a source-splicing boundary. The
// source regular expression is ASCII-only; Unicode case folding is not used.
// Its size limit counts UTF-16 code units after the separator is inserted.
func macroAddArgs(left, right string, limit int) (string, error) {
	asciiLetter := func(c byte) bool { return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' }
	if len(right) != 0 && asciiLetter(right[0]) {
		i := len(left)
		for i > 0 && asciiLetter(left[i-1]) {
			i--
		}
		if i < len(left) {
			end := i
			for i > 0 && left[i-1] == '\\' {
				i--
			}
			if (end-i)%2 == 1 {
				left += " "
			}
		}
	}
	if len(utf16.Encode([]rune(left)))+len(utf16.Encode([]rune(right))) > limit {
		return "", texError("MaxBufferSize", "MathJax internal buffer size exceeded; is there a recursive macro call?")
	}
	return left + right, nil
}

// Macros and paired delimiters use ParseUtil.substituteArgs before replacing
// their caller source. Environments retain their separate substitution path.
func substituteMacroArguments(body string, args []string) (string, error) {
	return substituteMacroArgumentsWithLimit(body, args, maxMacroBuffer)
}

func substituteMacroArgumentsWithLimit(body string, args []string, limit int) (string, error) {
	var text strings.Builder
	expansion := ""
	for i := 0; i < len(body); {
		c := body[i]
		i++
		switch c {
		case '\\':
			text.WriteByte(c)
			if i < len(body) {
				_, size := utf8.DecodeRuneInString(body[i:])
				text.WriteString(body[i : i+size])
				i += size
			}
		case '#':
			if i >= len(body) {
				return "", texError("IllegalMacroParam", "Illegal macro parameter reference")
			}
			c = body[i]
			i++
			if c == '#' {
				text.WriteByte(c)
				continue
			}
			if c < '1' || c > '9' || int(c-'1') >= len(args) {
				return "", texError("IllegalMacroParam", "Illegal macro parameter reference")
			}
			var err error
			expansion, err = macroAddArgs(expansion, text.String(), limit)
			if err != nil {
				return "", err
			}
			expansion, err = macroAddArgs(expansion, args[c-'1'], limit)
			if err != nil {
				return "", err
			}
			text.Reset()
		default:
			text.WriteByte(c)
		}
	}
	return macroAddArgs(expansion, text.String(), limit)
}
