// Copyright (c) 2020-2022 MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.
//
// Source: ts/input/tex/mathtools/MathtoolsMethods.ts and MathtoolsUtil.ts.

package tex

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/d2lang/mathjax-go/internal/mml"
)

const (
	mathtoolsPairSeparator = "\x00mathtools-pair\x00"
	mathtoolsOptionPrefix  = "\x00mathtools-option:"
	mathtoolsTagPrefix     = "\x00mathtools-tag:"
	mathtoolsCurrentTag    = "\x00mathtools-current-tag"
	mathtoolsTagSeparator  = "\x00mathtools-tag-field\x00"
)

var mathtoolsNumber = regexp.MustCompile(`^[-+]?(?:\d+(?:\.\d*)?|\.\d+)$`)

var mathtoolsDefaults = map[string]string{
	"multlinegap":           "1em",
	"multlined-pos":         "c",
	"firstline-afterskip":   "",
	"lastline-preskip":      "",
	"smallmatrix-align":     "c",
	"shortvdotsadjustabove": ".2em",
	"shortvdotsadjustbelow": ".2em",
	"centercolon":           "false",
	"centercolon-offset":    ".04em",
	"thincolon-dx":          "-.04em",
	"thincolon-dw":          "-.08em",
	"use-unicode":           "false",
	"prescript-sub-format":  "",
	"prescript-sup-format":  "",
	"prescript-arg-format":  "",
	"allow-mathtoolsset":    "true",
}

func (p *parser) mathtoolsOption(name string) string {
	if definition, ok := p.state.macros[mathtoolsOptionPrefix+name]; ok {
		return definition.body
	}
	return mathtoolsDefaults[name]
}

func (p *parser) mathtoolsOptionBool(name string) bool {
	return p.mathtoolsOption(name) == "true"
}

func (p *parser) mathtoolsUnderOverBracket(name string) ([]*mml.Node, error) {
	defaultThickness := ".1em"
	thickness, _, err := p.readBrackets(&defaultThickness)
	if err != nil {
		return nil, err
	}
	defaultHeight := ".2em"
	height, _, err := p.readBrackets(&defaultHeight)
	if err != nil {
		return nil, err
	}
	raw, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	base, err := p.parseString(raw)
	if err != nil {
		return nil, err
	}
	copy, err := p.parseString(raw)
	if err != nil {
		return nil, err
	}
	t := mathtoolsEm(thickness, .1)
	script := node("mpadded", node("mphantom", copy))
	border := "bottom"
	kind, accent := "mover", "accent"
	if name == "underbracket" {
		border = "top"
		kind, accent = "munder", "accentunder"
	}
	script.Attributes.Set("style", "border: "+t+" solid; border-"+border+": none")
	script.Attributes.Set("height", height)
	script.Attributes.Set("depth", 0)
	result := node(kind, base, script)
	result.Attributes.Set(accent, true)
	// ParseUtil.underOver(..., stack=true) wraps the stack as an OP atom and
	// installs both script-stacking properties on that outer TeXAtom.
	stack := texAtom(result, mml.TeXClassOp)
	stack.SetProperty("movesupsub", true)
	stack.SetProperty("subsupOK", true)
	return []*mml.Node{stack}, nil
}

func mathtoolsEm(value string, fallback float64) string {
	value = strings.TrimSpace(value)
	if strings.HasSuffix(value, "em") {
		if number, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(value, "em")), 64); err == nil {
			return emLength(number)
		}
	}
	if number, err := strconv.ParseFloat(value, 64); err == nil {
		return emLength(number)
	}
	return emLength(fallback)
}

func (p *parser) mathtoolsDeclarePairedDelimiter(name string) error {
	cs, err := p.readCSNameArgument(name)
	if err != nil {
		return err
	}
	arguments := 1
	pre, body, post := "", "#1", ""
	isX := strings.Contains(name, "X")
	isXPP := strings.Contains(name, "XPP")
	if isX {
		count, _, err := p.readBrackets(nil)
		if err != nil {
			return err
		}
		if count != "" {
			arguments, err = parseInteger(count, "IllegalMacroParam")
			if err != nil {
				return err
			}
		}
	}
	if isXPP {
		pre, _, err = p.readArgument(name, false)
		if err != nil {
			return err
		}
	}
	open, _, err := p.readArgument(name, false)
	if err != nil {
		return err
	}
	close, _, err := p.readArgument(name, false)
	if err != nil {
		return err
	}
	if isXPP {
		post, _, err = p.readArgument(name, false)
		if err != nil {
			return err
		}
	}
	if isX {
		body, _, err = p.readArgument(name, false)
		if err != nil {
			return err
		}
	}
	p.state.pairedDelimiters[cs] = pairedDelimiter{
		open: open, close: close, arguments: arguments,
		body: pre + mathtoolsPairSeparator + body + mathtoolsPairSeparator + post,
	}
	return nil
}

func (p *parser) mathtoolsPairedDelimiter(name string, definition pairedDelimiter) ([]*mml.Node, error) {
	parts := strings.Split(definition.body, mathtoolsPairSeparator)
	pre, body, post := "", definition.body, ""
	if len(parts) == 3 {
		pre, body, post = parts[0], parts[1], parts[2]
	}
	star := p.readStar()
	size := ""
	if !star {
		var err error
		size, _, err = p.readBrackets(nil)
		if err != nil {
			return nil, err
		}
	}
	left, right, middle := "", "", ""
	if star {
		left, right, middle = "\\left", "\\right", "\\middle"
	} else if size != "" {
		left, right, middle = size+"l", size+"r", size
	}
	if definition.arguments > 0 {
		args := make([]string, definition.arguments)
		for i := range args {
			var err error
			args[i], _, err = p.readArgument(name, false)
			if err != nil {
				return nil, err
			}
		}
		var err error
		if pre, err = substituteMacroArguments(pre, args); err != nil {
			return nil, err
		}
		if body, err = substituteMacroArguments(body, args); err != nil {
			return nil, err
		}
		if post, err = substituteMacroArguments(post, args); err != nil {
			return nil, err
		}
	}
	body = strings.ReplaceAll(body, "\\delimsize", middle)
	expansion := ""
	for _, part := range []string{pre, left, definition.open, body, right, definition.close, post, p.source[p.pos:]} {
		var err error
		expansion, err = macroAddArgs(expansion, part, maxMacroBuffer)
		if err != nil {
			return nil, err
		}
	}
	// PairedDelimiters rewrites this caller before charging the expansion.
	// Its pending row/script items consume the resulting source directly.
	p.source, p.pos = expansion, 0
	p.state.macroCount++
	if p.state.macroCount > maxMacros {
		return nil, texError("MaxMacroSub1", "MathJax maximum macro substitution count exceeded; is here a recursive macro call?")
	}
	return nil, nil
}

func (p *parser) mathtoolsCenterColon(center, force, thin bool) *mml.Node {
	colon := p.token("mo", ":")
	if !center || (!force && !p.mathtoolsOptionBool("centercolon")) {
		return colon
	}
	dy := p.mathtoolsOption("centercolon-offset")
	padded := node("mpadded", colon)
	padded.Attributes.Set("voffset", dy)
	padded.Attributes.Set("height", "+"+dy)
	padded.Attributes.Set("depth", "-"+dy)
	if thin {
		padded.Attributes.Set("width", p.mathtoolsOption("thincolon-dw"))
		padded.Attributes.Set("lspace", p.mathtoolsOption("thincolon-dx"))
	}
	return padded
}

func (p *parser) mathtoolsRelation(name string) ([]*mml.Node, error) {
	table := map[string][2]string{
		"coloneqq": {":=", "≔"}, "Coloneqq": {"::=", "⩴"},
		"coloneq": {":-", ""}, "Coloneq": {"::-", ""},
		"eqqcolon": {"=:", "≕"}, "Eqqcolon": {"=::", ""},
		"eqcolon": {"-:", "∹"}, "Eqcolon": {"-::", ""},
		"colonapprox": {":\\approx", ""}, "Colonapprox": {"::\\approx", ""},
		"colonsim": {":\\sim", ""}, "Colonsim": {"::\\sim", ""},
		"dblcolon": {"::", "∷"},
	}
	relation := table[name]
	if p.mathtoolsOptionBool("use-unicode") && relation[1] != "" {
		return []*mml.Node{p.operator(relation[1], mml.TeXClassRel, nil)}, nil
	}
	expansion := strings.ReplaceAll(relation[0], ":", "\\MTThinColon")
	expansion = strings.ReplaceAll(expansion, "-", "\\mathrel{-}")
	return p.parseContinuationExpansion("\\mathrel{" + expansion + "}")
}

func mathtoolsNArrow(name string) *mml.Node {
	character, dy := "↑", ".06em"
	if name == "ndownarrow" {
		character, dy = "↓", ".25em"
	}
	space := node("mspace")
	space.Attributes.Set("height", ".2em")
	space.Attributes.Set("depth", 0)
	space.Attributes.Set("width", ".4em")
	enclose := node("menclose", space)
	enclose.Attributes.Set("notation", "updiagonalstrike")
	enclose.Attributes.Set("data-thickness", ".05em")
	enclose.Attributes.Set("data-padding", 0)
	inner := node("mpadded", enclose)
	inner.Attributes.Set("width", 0)
	inner.Attributes.Set("lspace", "-.5width")
	inner.Attributes.Set("voffset", dy)
	outer := node("mpadded", inner, node("mphantom", token("mtext", character)))
	outer.Attributes.Set("width", 0)
	outer.Attributes.Set("lspace", "-.5width")
	result := node("TeXAtom", token("mtext", character), outer)
	result.TeXClass = mml.TeXClassRel
	result.SetProperty("texClass", mml.TeXClassRel)
	return result
}

func (p *parser) mathtoolsSplitFrac(name string, display bool) ([]*mml.Node, error) {
	numerator, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	denominator, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	numChildren := append(unwrapInferred(numerator), token("mi", ""), mathtoolsSpace("1em"))
	numStyle := node("mstyle", numChildren...)
	numStyle.Attributes.Set("scriptlevel", 0)
	denChildren := append([]*mml.Node{mathtoolsSpace("1em"), token("mi", "")}, unwrapInferred(denominator)...)
	denStyle := node("mstyle", denChildren...)
	denStyle.Attributes.Set("scriptlevel", 0)
	fraction := node("mfrac", numStyle, denStyle)
	fraction.Attributes.Set("linethickness", 0)
	fraction.Attributes.Set("numalign", "left")
	fraction.Attributes.Set("denomalign", "right")
	style := node("mstyle", fraction)
	style.Attributes.Set("displaystyle", display)
	style.Attributes.Set("scriptlevel", 0)
	return []*mml.Node{style}, nil
}

func mathtoolsSpace(width string) *mml.Node {
	space := node("mspace")
	space.Attributes.Set("width", width)
	return space
}

func (p *parser) mathtoolsXMathStrut(name string) ([]*mml.Node, error) {
	depth, _, err := p.readBrackets(nil)
	if err != nil {
		return nil, err
	}
	height, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	height, err = mathtoolsPlusOrMinus(name, height)
	if err != nil {
		return nil, err
	}
	if depth == "" {
		depth = height
	}
	depth, err = mathtoolsPlusOrMinus(name, depth)
	if err != nil {
		return nil, err
	}
	paren := p.token("mo", "(")
	paren.Attributes.Set("stretchy", false)
	padded := node("mpadded", node("mphantom", paren))
	padded.Attributes.Set("width", 0)
	padded.Attributes.Set("height", height+"height")
	padded.Attributes.Set("depth", depth+"depth")
	return []*mml.Node{texAtom(padded, mml.TeXClassOrd)}, nil
}

func mathtoolsPlusOrMinus(name, value string) (string, error) {
	value = strings.TrimSpace(value)
	if !mathtoolsNumber.MatchString(value) {
		return "", texError("NotANumber", "Argument to \\%s is not a number", name)
	}
	if value[0] != '+' && value[0] != '-' {
		value = "+" + value
	}
	return value, nil
}

func (p *parser) mathtoolsAdjustLimits(name string) ([]*mml.Node, error) {
	first, err := p.readDelimitedParameter(name, "_")
	if err != nil {
		return nil, err
	}
	firstSub, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	second, err := p.readDelimitedParameter(name, "_")
	if err != nil {
		return nil, err
	}
	secondSub, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	expansion := "\\mathop{{" + first + "}\\vphantom{{" + second + "}}}_{{" + firstSub + "}\\vphantom{{" + secondSub + "}}}" +
		"\\mathop{{" + second + "}\\vphantom{{" + first + "}}}_{{" + secondSub + "}\\vphantom{{" + firstSub + "}}}"
	nodes, err := p.parseContinuationExpansion(expansion)
	if err != nil {
		return nil, err
	}
	if len(nodes) == 2 {
		nodes = []*mml.Node{nodes[0], p.operator("\u2061", mml.TeXClassNone, nil), nodes[1]}
	}
	return nodes, nil
}

func (p *parser) mathtoolsSetOptions(name string) error {
	if !p.mathtoolsOptionBool("allow-mathtoolsset") {
		return texError("ForbiddenMathtoolsSet", "\\%s is disabled", name)
	}
	raw, _, err := p.readArgument(name, false)
	if err != nil {
		return err
	}
	for _, entry := range splitTopLevel(raw, ',') {
		parts := strings.SplitN(entry, "=", 2)
		key := strings.TrimSpace(parts[0])
		if _, ok := mathtoolsDefaults[key]; !ok || key == "allow-mathtoolsset" {
			return texError("InvalidOption", "Invalid option: %s", key)
		}
		value := "true"
		if len(parts) == 2 {
			value = strings.TrimSpace(strings.Trim(parts[1], "{}"))
		}
		p.state.macros[mathtoolsOptionPrefix+key] = macroDefinition{body: value}
	}
	// The source command map handles ':' as a special character.  Until the
	// central scanner calls mathtoolsCharacter, preserve the same behavior for
	// the unconsumed source in this parser invocation.
	if p.mathtoolsOptionBool("centercolon") {
		rest := strings.ReplaceAll(p.source[p.pos:], ":", "\\centercolon ")
		p.source = p.source[:p.pos] + rest
	}
	return nil
}

func (p *parser) mathtoolsNewTagForm(name string, renew bool) error {
	id, _, err := p.readArgument(name, false)
	if err != nil {
		return err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return texError("InvalidTagFormID", "Tag form name can't be empty")
	}
	format, _, err := p.readBrackets(nil)
	if err != nil {
		return err
	}
	left, _, err := p.readArgument(name, false)
	if err != nil {
		return err
	}
	right, _, err := p.readArgument(name, false)
	if err != nil {
		return err
	}
	key := mathtoolsTagPrefix + id
	if _, exists := p.state.macros[key]; exists && !renew {
		return texError("DuplicateTagForm", "Duplicate tag form: %s", id)
	}
	p.state.macros[key] = macroDefinition{body: left + mathtoolsTagSeparator + right + mathtoolsTagSeparator + format}
	return nil
}

func (p *parser) mathtoolsUseTagForm(name string) error {
	id, _, err := p.readArgument(name, false)
	if err != nil {
		return err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		delete(p.state.macros, mathtoolsCurrentTag)
		return nil
	}
	definition, exists := p.state.macros[mathtoolsTagPrefix+id]
	if !exists {
		return texError("UndefinedTagForm", "Undefined tag form: %s", id)
	}
	p.state.macros[mathtoolsCurrentTag] = definition
	return nil
}

// mathtoolsFormatTag applies the currently selected Mathtools tag form.  It is
// deliberately a side-effect-free helper so the central AMS tag owner can use
// it when constructing an mlabeledtr without depending on handler internals.
func (p *parser) mathtoolsFormatTag(tag string) string {
	definition, ok := p.state.macros[mathtoolsCurrentTag]
	if !ok {
		return "(" + tag + ")"
	}
	parts := strings.Split(definition.body, mathtoolsTagSeparator)
	if len(parts) != 3 {
		return "(" + tag + ")"
	}
	formatted := tag
	if parts[2] != "" {
		formatted = parts[2] + "{" + tag + "}"
	}
	return parts[0] + formatted + parts[1]
}
