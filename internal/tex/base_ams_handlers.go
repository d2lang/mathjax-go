// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.
//
// Sources: ts/input/tex/ParseUtil.ts, base/{BaseItems,BaseMethods}.ts,
// and ams/{AmsMethods,AmsItems,AmsMappings}.ts.

package tex

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// baseAMSCommand is the narrow hook for the remaining source methods whose
// stack behavior is not represented by the legacy direct handlers.
func (p *parser) baseAMSCommand(name string) (nodes []*mml.Node, handled bool, err error) {
	switch name {
	case "genfrac":
		nodes, err = p.amsGenfrac(name)
	case "DeclareMathOperator":
		err = p.amsDeclareMathOperator(name)
	case "operatorname":
		nodes, err = p.amsOperatorName(name)
	case "boxed":
		nodes, err = p.amsBoxed(name)
	case "idotsint":
		nodes, err = p.amsMultiIntegral(name)
	case "sideset":
		nodes, err = p.amsSideSet(name)
	case "xrightarrow", "xleftarrow", "xleftrightarrow", "xLeftarrow", "xRightarrow", "xLeftrightarrow",
		"xhookleftarrow", "xhookrightarrow", "xmapsto", "xrightharpoondown", "xleftharpoondown",
		"xrightleftharpoons", "xrightharpoonup", "xleftharpoonup", "xleftrightharpoons":
		nodes, err = p.amsXArrow(name)
	default:
		return nil, false, nil
	}
	return nodes, true, err
}

// baseAMSEnvironment owns flalign/flalign*, whose row padding is performed by
// FlalignItem after ordinary alignment cells are complete.
func (p *parser) baseAMSEnvironment(name string) (nodes []*mml.Node, handled bool, err error) {
	if name != "flalign" && name != "flalign*" {
		return nil, false, nil
	}
	body, err := p.captureEnvironment(name)
	if err != nil {
		return nil, true, err
	}
	rows := splitTable(body)
	mrows := make([]*mml.Node, 0, len(rows))
	for _, cells := range rows {
		parsed := make([]*mml.Node, 0, len(cells)+1)
		for _, raw := range cells {
			content, err := p.parseString(strings.TrimSpace(raw))
			if err != nil {
				return nil, true, err
			}
			parsed = append(parsed, node("mtd", content))
		}
		padded := make([]*mml.Node, 0, len(parsed)+len(parsed)/2)
		for len(parsed) != 0 {
			padded = append(padded, parsed[0])
			parsed = parsed[1:]
			if len(parsed) != 0 {
				padded = append(padded, parsed[0])
				parsed = parsed[1:]
			}
			if len(parsed) != 0 {
				padded = append(padded, node("mtd"))
			}
		}
		mrows = append(mrows, node("mtr", padded...))
	}
	table := node("mtable", mrows...)
	prefixRelationColumns(table)
	resetTableAttributes(table,
		"width", "100%",
		"displaystyle", true,
		"columnalign", "right left center right left",
		"columnspacing", "0em",
		"columnwidth", "auto auto fit auto auto",
		"rowspacing", "3pt",
		"side", "right",
		"minlabelspacing", "0.8em",
		"data-width-includes-label", true,
	)
	return []*mml.Node{table}, true, nil
}

func (p *parser) amsGenfrac(name string) ([]*mml.Node, error) {
	left, err := p.amsGenfracDelimiter(name)
	if err != nil {
		return nil, err
	}
	right, err := p.amsGenfracDelimiter(name)
	if err != nil {
		return nil, err
	}
	thickness, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	style, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	numerator, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	denominator, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	fraction := node("mfrac", numerator, denominator)
	if thickness != "" {
		fraction.Attributes.Set("linethickness", thickness)
	}
	var content *mml.Node = fraction
	if left != nil || right != nil {
		fraction.SetProperty("withDelims", true)
		content, err = p.amsGenfracFixedFence(left, fraction, right)
		if err != nil {
			return nil, err
		}
	}
	style = strings.TrimSpace(style)
	if style != "" {
		attributes := map[string]any{}
		switch style {
		case "0":
			attributes["displaystyle"] = true
			attributes["scriptlevel"] = 0
		case "1", "2", "3":
			attributes["displaystyle"] = false
			attributes["scriptlevel"] = int(style[0] - '1')
		default:
			return nil, texError("BadMathStyleFor", "Bad math style for \\%s", name)
		}
		styled := node("mstyle", content)
		styled.Attributes.Set("displaystyle", attributes["displaystyle"])
		styled.Attributes.Set("scriptlevel", attributes["scriptlevel"])
		content = styled
	}
	return []*mml.Node{content}, nil
}

// amsFixedFencePalette ports ParseUtil.mathPalette.  MathChoice selects bigg
// in display style and big in text, script, and scriptscript styles during
// the inherited-attribute pass.
func amsFixedFencePalette(character string, class mml.TeXClass) *mml.Node {
	display := amsFixedFence(character, class, "2.047em")
	text := amsFixedFence(character, class, "1.2em")
	script := amsFixedFence(character, class, "1.2em")
	scriptScript := amsFixedFence(character, class, "1.2em")
	return node("MathChoice", display, text, script, scriptScript)
}

// amsFixedFence ports BaseMethods.MakeBig.  The sizes preserve JavaScript's
// source truncation after multiplying by P_HEIGHT (1.2/.85).
func amsFixedFence(character string, class mml.TeXClass, size string) *mml.Node {
	mo := token("mo", character)
	mo.Attributes.Set("minsize", size)
	mo.Attributes.Set("maxsize", size)
	mo.Attributes.Set("fence", true)
	mo.Attributes.Set("stretchy", true)
	mo.Attributes.Set("symmetric", true)
	return texAtom(mo, class)
}

func (p *parser) amsDeclareMathOperator(name string) error {
	star := p.readStar()
	cs, err := p.readCSNameArgument(name)
	if err != nil {
		return err
	}
	operator, _, err := p.readArgument(name, false)
	if err != nil {
		return err
	}
	body := "\\operatorname"
	if star {
		body += "*"
	}
	p.state.macros[cs] = macroDefinition{body: body + "{" + operator + "}"}
	return nil
}

func (p *parser) amsOperatorName(name string) ([]*mml.Node, error) {
	star := p.readStar()
	raw, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	raw = strings.TrimSpace(raw)
	children := make([]*mml.Node, 0, 3)
	for position := 0; position < len(raw); {
		r, size := utf8.DecodeRuneInString(raw[position:])
		if unicode.IsLetter(r) || r == '-' || r == '*' {
			start := position
			position += size
			for position < len(raw) {
				next, nextSize := utf8.DecodeRuneInString(raw[position:])
				if !unicode.IsLetter(next) && next != '-' && next != '*' {
					break
				}
				position += nextSize
			}
			identifier := token("mi", raw[start:position])
			identifier.Attributes.Set("mathvariant", "normal")
			children = append(children, identifier)
			continue
		}
		if unicode.IsSpace(r) {
			position += size
			continue
		}
		sub := &parser{source: raw, pos: position, state: p.state, display: p.display}
		result, err := sub.parseOneTokenEvent()
		if err != nil {
			return nil, err
		}
		children = append(children, result.nodes...)
		tail, err := result.afterNode.complete(sub)
		if err != nil {
			return nil, err
		}
		children = append(children, tail...)
		raw, position = sub.source, sub.pos
	}
	var result *mml.Node
	if len(children) == 1 && children[0].Kind == "mi" {
		result = children[0]
	} else {
		result = node("TeXAtom", children...)
	}
	// HandleOperatorName reparses in an explicit normal-font environment.
	// That environment suppresses Physics' vector token factory, including
	// for fixed-symbol tokens whose own variant is not replaced by normal.
	result.Walk(func(n *mml.Node) bool {
		if origin, _ := n.Property(vectorFactoryToken); origin == true {
			n.SetProperty(vectorFactoryDone, true)
		}
		return true
	})
	result.TeXClass = mml.TeXClassOp
	// The primary TeXAtom constructor owns texClass before the final writes.
	// A singular mi receives those properties in the handler's source order.
	if result.Kind != "mi" {
		result.SetProperty("texClass", mml.TeXClassOp)
	}
	result.SetProperty("movesupsub", star)
	result.SetProperty("movablelimits", true)
	result.SetProperty("texClass", mml.TeXClassOp)
	if !star {
		start := p.pos
		p.skipSpaces()
		if p.pos < len(p.source) && p.source[p.pos] == '\\' {
			p.pos++
			if p.readControlSequence() != "limits" {
				p.pos = start
			}
		} else {
			p.pos = start
		}
	}
	return []*mml.Node{result}, nil
}

func (p *parser) amsBoxed(name string) ([]*mml.Node, error) {
	argument, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	inner := texAtom(argument, mml.TeXClassOrd)
	style := node("mstyle", inner)
	style.Attributes.Set("displaystyle", true)
	style.Attributes.Set("scriptlevel", 0)
	boxed := node("menclose", texAtom(style, mml.TeXClassOrd))
	boxed.Attributes.Set("notation", "box")
	return []*mml.Node{boxed}, nil
}

func (p *parser) amsMultiIntegral(_ string) ([]*mml.Node, error) {
	return p.parseExpansion("\\int\\cdots\\int")
}

func (p *parser) amsSideSet(name string) ([]*mml.Node, error) {
	pre, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	post, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	base, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	preSub, preSup := amsSideScripts(pre)
	postSub, postSup := amsSideScripts(post)
	if postSub == nil {
		postSub = node("none")
	}
	if postSup == nil {
		postSup = node("none")
	}
	if preSub == nil {
		preSub = node("none")
	}
	if preSup == nil {
		preSup = node("none")
	}
	multi := node("mmultiscripts", base, postSub, postSup, node("mprescripts"), preSub, preSup)
	multi.SetProperty("scriptalign", "left")
	result := texAtom(multi, mml.TeXClassOp)
	result.SetProperty("movesupsub", true)
	result.SetProperty("movablelimits", true)
	return []*mml.Node{result}, nil
}

func amsSideScripts(script *mml.Node) (*mml.Node, *mml.Node) {
	if script == nil {
		return nil, nil
	}
	if script.Kind == "mrow" && script.Flags.Inferred && len(script.Children) == 1 {
		script = script.Children[0]
	}
	switch script.Kind {
	case "msubsup":
		return script.Children[1], script.Children[2]
	case "msub":
		return script.Children[1], nil
	case "msup":
		return nil, script.Children[1]
	}
	return nil, nil
}

type amsArrowSize struct {
	character string
	left      float64
	right     float64
}

var amsArrowSizes = map[string]amsArrowSize{
	"xrightarrow": {"→", 5, 10}, "xleftarrow": {"←", 10, 5},
	"xleftrightarrow": {"↔", 10, 10}, "xLeftarrow": {"⇐", 12, 7},
	"xRightarrow": {"⇒", 7, 12}, "xLeftrightarrow": {"⇔", 12, 12},
	"xhookleftarrow": {"↩", 10, 5}, "xhookrightarrow": {"↪", 5, 10},
	"xmapsto": {"↦", 10, 10}, "xrightharpoondown": {"⇁", 5, 10},
	"xleftharpoondown": {"↽", 10, 5}, "xrightleftharpoons": {"⇌", 10, 10},
	"xrightharpoonup": {"⇀", 5, 10}, "xleftharpoonup": {"↼", 10, 5},
	"xleftrightharpoons": {"⇋", 10, 10},
}

func (p *parser) amsXArrow(name string) ([]*mml.Node, error) {
	belowRaw, hasBelow, err := p.readBrackets(nil)
	if err != nil {
		return nil, err
	}
	above, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	size := amsArrowSizes[name]
	arrow := operator(size.character, mml.TeXClassRel, nil)
	arrow.Attributes.Set("stretchy", true)
	arrowStyle := node("mstyle", arrow)
	arrowStyle.Attributes.Set("scriptlevel", 0)
	overChildren := append(unwrapInferred(above), amsArrowStrut("depth", ".25em"))
	over := node("mpadded", overChildren...)
	amsSetArrowPadding(over, size)
	over.Attributes.Set("voffset", "-.2em")
	over.Attributes.Set("height", "-.2em")
	if !hasBelow {
		result := node("mover", arrowStyle, over)
		result.SetProperty("subsupOK", true)
		return []*mml.Node{result}, nil
	}
	below, err := p.parseString(belowRaw)
	if err != nil {
		return nil, err
	}
	underChildren := append(unwrapInferred(below), amsArrowStrut("height", ".75em"))
	under := node("mpadded", underChildren...)
	amsSetArrowPadding(under, size)
	under.Attributes.Set("voffset", ".15em")
	under.Attributes.Set("depth", "-.15em")
	result := node("munderover", arrowStyle, under, over)
	result.SetProperty("subsupOK", true)
	return []*mml.Node{result}, nil
}

func amsArrowStrut(attribute, value string) *mml.Node {
	strut := node("mspace")
	strut.Attributes.Set(attribute, value)
	return strut
}

func amsSetArrowPadding(padded *mml.Node, size amsArrowSize) {
	padded.Attributes.Set("width", "+"+emLength((size.left+size.right)/18))
	padded.Attributes.Set("lspace", emLength(size.left/18))
}
