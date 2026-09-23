// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Source: TexParser.GetDimen/GetArgument and ParseUtil.matchDimen (MathJax 3.2.2).
package tex

import (
	"regexp"
	"strings"
)

const dimensionNumber = `([-+]?(?:[.,][0-9]+|[0-9]+(?:[.,][0-9]*)?))`
const dimensionUnit = `(pt|em|ex|mu|px|mm|cm|in|pc)`

var (
	dimensionSpace  = regexp.MustCompile(`^` + muDimensionSpace + `*`)
	dimensionFull   = regexp.MustCompile(`^` + muDimensionSpace + `*` + dimensionNumber + muDimensionSpace + `*` + dimensionUnit + muDimensionSpace + `*$`)
	dimensionPrefix = regexp.MustCompile(`^` + muDimensionSpace + `*` + dimensionNumber + muDimensionSpace + `*` + dimensionUnit + ` ?`)
)

// readDimension owns only the source GetDimen argument. Its local whitespace,
// balanced-brace scan and failure cursor do not change generic argument parsing.
// name is the invoking control sequence used by primary parser.currentCS.
func (p *parser) readDimension(name string) (string, error) {
	if name == "vspace" || name == "raisebox" {
		return p.readLegacyDimension(name)
	}
	missing := func() (string, error) {
		return "", texError("MissingDimOrUnits", "Missing dimension or its units for \\%s", name)
	}
	if p.pos >= len(p.source) {
		return missing()
	}
	p.pos += len(dimensionSpace.FindString(p.source[p.pos:]))
	if p.pos >= len(p.source) {
		return missing()
	}
	if p.source[p.pos] == '{' {
		start := p.pos + 1
		p.pos++
		depth := 1
		for p.pos < len(p.source) {
			switch p.consumeRune() {
			case '\\':
				if p.pos < len(p.source) {
					p.consumeRune()
				} else {
					// GetArgument advances once even after a final escape. This
					// error-only cursor is deliberately one past physical EOF.
					p.pos++
				}
			case '{':
				depth++
			case '}':
				depth--
				if depth == 0 {
					match := dimensionFull.FindStringSubmatch(p.source[start : p.pos-1])
					if match == nil {
						return missing()
					}
					return dimensionValue(match), nil
				}
			}
		}
		return "", texError("MissingCloseBrace", "Missing close brace")
	}
	match := dimensionPrefix.FindStringSubmatch(p.source[p.pos:])
	if match == nil {
		return missing()
	}
	p.pos += len(match[0])
	return dimensionValue(match), nil
}

func dimensionValue(match []string) string {
	return normalizeTeXMu(strings.Replace(match[1], ",", ".", 1) + match[2])
}
