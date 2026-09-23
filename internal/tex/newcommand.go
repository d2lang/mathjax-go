// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
//
// This file is a Go translation and modification of MathJax 3.2.2.
// Sources: ts/input/tex/ParseUtil.ts, base/BaseMethods.ts, and
// newcommand/{NewcommandMethods,NewcommandUtil}.ts.

package tex

import (
	"strings"
	"unicode/utf8"

	"github.com/d2lang/mathjax-go/internal/mml"
)

func (p *parser) invokeMacro(name string, definition macroDefinition) ([]*mml.Node, error) {
	expansion := definition.body
	if definition.arguments != 0 {
		if definition.prefix != "" {
			p.skipSpaces()
			if !strings.HasPrefix(p.source[p.pos:], definition.prefix) {
				return nil, texError("MismatchUseDef", "Use of \\%s doesn't match its definition", name)
			}
			p.pos += len(definition.prefix)
		}
		args := make([]string, 0, definition.arguments)
		if definition.optionalDefault != nil {
			arg, present, err := p.readBrackets(definition.optionalDefault)
			if err != nil {
				return nil, err
			}
			if !present {
				arg = *definition.optionalDefault
			}
			args = append(args, arg)
		}
		for len(args) < definition.arguments {
			var arg string
			var err error
			index := len(args)
			if index < len(definition.delimiters) && definition.delimiters[index] != "" {
				arg, err = p.readDelimitedParameter(name, definition.delimiters[index])
			} else {
				arg, _, err = p.readArgument(name, false)
			}
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
		var err error
		expansion, err = substituteMacroArguments(expansion, args)
		if err != nil {
			return nil, err
		}
	}
	joined, err := macroAddArgs(expansion, p.source[p.pos:], maxMacroBuffer)
	if err != nil {
		return nil, err
	}
	// Macro and MacroWithTemplate install only the unparsed program before
	// charging the counter. Existing row/prime/script state remains in place.
	p.source, p.pos = joined, 0
	p.state.macroCount++
	if p.state.macroCount > maxMacros {
		return nil, texError("MaxMacroSub1", "MathJax maximum macro substitution count exceeded; is here a recursive macro call?")
	}
	return nil, nil
}

func substituteArguments(body string, args []string) (string, error) {
	var result strings.Builder
	for i := 0; i < len(body); i++ {
		if body[i] != '#' {
			result.WriteByte(body[i])
			continue
		}
		if i+1 >= len(body) {
			return "", texError("IllegalMacroParam", "Illegal macro parameter reference")
		}
		i++
		if body[i] == '#' {
			result.WriteByte('#')
			continue
		}
		if body[i] < '1' || body[i] > '9' {
			return "", texError("IllegalMacroParam", "Illegal macro parameter reference")
		}
		index := int(body[i] - '1')
		if index >= len(args) {
			return "", texError("IllegalMacroParam", "Illegal macro parameter reference")
		}
		result.WriteString(args[index])
	}
	return result.String(), nil
}

func (p *parser) defineCommand(name string) error {
	cs, err := p.readCSNameArgument(name)
	if err != nil {
		return err
	}
	countRaw, present, err := p.readBrackets(nil)
	if err != nil {
		return err
	}
	count := 0
	if present {
		count, err = parseInteger(countRaw, "IllegalMacroParam")
		if err != nil || count < 0 || count > 9 {
			return texError("IllegalMacroParam", "Illegal number of parameters specified in \\%s", name)
		}
	}
	var optional *string
	if count > 0 {
		if value, ok, err := p.readBrackets(nil); err != nil {
			return err
		} else if ok {
			optional = &value
		}
	}
	body, _, err := p.readArgument(name, false)
	if err != nil {
		return err
	}
	if max := macroArguments(body); max > count {
		return texError("IllegalMacroParam", "Illegal macro parameter reference")
	}
	p.state.macros[cs] = macroDefinition{body: body, arguments: count, optionalDefault: optional}
	return nil
}

func (p *parser) defineEnvironment(name string) error {
	environment, _, err := p.readArgument(name, false)
	if err != nil {
		return err
	}
	environment = strings.TrimSpace(environment)
	countRaw, present, err := p.readBrackets(nil)
	if err != nil {
		return err
	}
	count := 0
	if present {
		count, err = parseInteger(countRaw, "IllegalMacroParam")
		if err != nil || count < 0 || count > 9 {
			return texError("IllegalMacroParam", "Illegal number of parameters specified in \\%s", name)
		}
	}
	var optional *string
	if count > 0 {
		if value, ok, err := p.readBrackets(nil); err != nil {
			return err
		} else if ok {
			optional = &value
		}
	}
	begin, _, err := p.readArgument(name, false)
	if err != nil {
		return err
	}
	end, _, err := p.readArgument(name, false)
	if err != nil {
		return err
	}
	p.state.environments[environment] = environmentDefinition{
		begin: begin, end: end, arguments: count, optionalDefault: optional,
	}
	return nil
}

func (p *parser) macroDef(name string) error {
	p.skipSpaces()
	if p.pos >= len(p.source) || p.source[p.pos] != '\\' {
		return texError("MissingCS", "%s must be followed by a control sequence", "\\"+name)
	}
	p.pos++
	cs := p.readControlSequence()
	start := p.pos
	for p.pos < len(p.source) && p.source[p.pos] != '{' {
		if p.source[p.pos] == '\\' {
			p.pos++
			p.readControlSequence()
			continue
		}
		p.consumeRune()
	}
	template := p.source[start:p.pos]
	body, _, err := p.readArgument(name, false)
	if err != nil {
		return err
	}
	definition := macroDefinition{body: body}
	prefix, delimiters, count, err := parseParameterTemplate(template)
	if err != nil {
		return err
	}
	definition.prefix = prefix
	definition.delimiters = delimiters
	definition.arguments = count
	p.state.macros[cs] = definition
	return nil
}

func parseParameterTemplate(template string) (string, []string, int, error) {
	positions := make([]int, 0, 9)
	for i := 0; i < len(template); i++ {
		if template[i] != '#' {
			continue
		}
		if i+1 >= len(template) || template[i+1] < '1' || template[i+1] > '9' {
			return "", nil, 0, texError("IllegalMacroParam", "Illegal macro parameter reference")
		}
		want := len(positions) + 1
		if int(template[i+1]-'0') != want {
			return "", nil, 0, texError("IllegalMacroParam", "Parameters must be numbered consecutively")
		}
		positions = append(positions, i)
		i++
	}
	if len(positions) == 0 {
		return template, nil, 0, nil
	}
	prefix := template[:positions[0]]
	delimiters := make([]string, len(positions))
	for i, position := range positions {
		start := position + 2
		end := len(template)
		if i+1 < len(positions) {
			end = positions[i+1]
		}
		delimiters[i] = template[start:end]
	}
	return prefix, delimiters, len(positions), nil
}

func (p *parser) letCommand(name string) error {
	cs, err := p.readCSName(name)
	if err != nil {
		return err
	}
	p.skipSpaces()
	if p.pos < len(p.source) && p.source[p.pos] == '=' {
		p.pos++
		p.skipSpaces()
	}
	if p.pos >= len(p.source) {
		return nil
	}
	if p.source[p.pos] == '\\' {
		p.pos++
		source := p.readControlSequence()
		if definition, ok := p.state.macros[source]; ok {
			p.state.macros[cs] = definition
			return nil
		}
		if _, ok := identifierSymbols[source]; ok {
			p.state.macros[cs] = macroDefinition{body: "\\" + source}
			return nil
		}
		if _, ok := operatorSymbols[source]; ok {
			p.state.macros[cs] = macroDefinition{body: "\\" + source}
			return nil
		}
		// MathJax's \let of an undefined CS is a no-op.
		return nil
	}
	start := p.pos
	_, size := utf8.DecodeRuneInString(p.source[p.pos:])
	p.pos += size
	p.state.macros[cs] = macroDefinition{body: p.source[start:p.pos]}
	return nil
}

func (p *parser) readCSNameArgument(name string) (string, error) {
	raw, _, err := p.readArgument(name, false)
	if err != nil {
		return "", err
	}
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "\\") {
		return "", texError("MissingCS", "%s must be followed by a control sequence", "\\"+name)
	}
	sub := &parser{source: raw[1:]}
	cs := sub.readControlSequence()
	if sub.pos != len(sub.source) {
		return "", texError("MissingCS", "%s must be followed by a control sequence", "\\"+name)
	}
	return cs, nil
}

func (p *parser) readCSName(name string) (string, error) {
	p.skipSpaces()
	if p.pos >= len(p.source) || p.source[p.pos] != '\\' {
		return "", texError("MissingCS", "%s must be followed by a control sequence", "\\"+name)
	}
	p.pos++
	return p.readControlSequence(), nil
}

func (p *parser) readDelimitedParameter(name, delimiter string) (string, error) {
	p.skipSpaces()
	start := p.pos
	depth := 0
	for p.pos < len(p.source) {
		if depth == 0 && strings.HasPrefix(p.source[p.pos:], delimiter) {
			value := p.source[start:p.pos]
			p.pos += len(delimiter)
			return value, nil
		}
		switch p.source[p.pos] {
		case '\\':
			p.pos++
			p.readControlSequence()
		case '{':
			depth++
			p.pos++
		case '}':
			if depth > 0 {
				depth--
			}
			p.pos++
		default:
			p.consumeRune()
		}
	}
	return "", texError("RunawayArgument", "Runaway argument for \\%s", name)
}

func macroNode(nodes []*mml.Node) *mml.Node { return row(nodes, true) }
