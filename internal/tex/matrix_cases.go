// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Go translation of base/BaseMethods.Entry (MathJax 3.2.2).

package tex

import (
	"strings"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// Entry scans the second cases column as text before ordinary TeX tokenization.
// In particular, percent is literal here, and a row command ends the scan even
// inside braces. close records the first complete group for the \text exception.
func matrixCasesTextEnd(source string, start int) (end, close int, err error) {
	braces, close := 0, -1
	i := start
	for i < len(source) {
		switch source[i] {
		case '{':
			braces++
			i++
		case '}':
			if braces == 0 {
				return i, close, nil
			}
			braces--
			if braces == 0 && close < 0 {
				close = i - start
			}
			i++
		case '&':
			if braces == 0 {
				return i, close, texError("ExtraAlignTab", "Extra alignment tab in \\cases text")
			}
			i++
		case '\\':
			rest := source[i:]
			if strings.HasPrefix(rest, `\\`) || strings.HasPrefix(rest, `\cr`) && len(rest) > 3 && !isASCIILetter(rune(rest[3])) {
				return i, close, nil
			}
			i += 2
		default:
			i++
		}
	}
	return len(source), close, nil
}

func (p *parser) parseMatrixCasesCell(source string, terminator byte) (*mml.Node, error) {
	_, close, err := matrixCasesTextEnd(source, 0)
	if err != nil {
		return nil, err
	}
	leading := strings.TrimLeftFunc(source, internalTextSpace)
	if strings.HasPrefix(leading, `\text`) && len(leading) > 5 && !isASCIILetter(rune(leading[5])) &&
		close == len(strings.TrimRightFunc(source, internalTextSpace))-1 {
		// Preserve the source's backwards compatibility: a complete \text
		// cell is parsed normally, without the automatic level-zero mstyle.
		return p.parseMatrixCell(source, terminator)
	}
	text := strings.TrimFunc(source, internalTextSpace)
	if strings.HasSuffix(text, `\`) && strings.HasSuffix(source, " ") {
		text += " " // ParseUtil.trimSpaces preserves a trailing control-space.
	}
	sub := p.matrixCellParser("")
	children, err := sub.internalMath(text, "", true)
	if err != nil {
		return nil, err
	}
	return matrixCellContent(children), nil
}
