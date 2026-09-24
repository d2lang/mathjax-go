// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.
//
// Sources: ts/input/tex/ParseUtil.ts, base/{BaseItems,BaseMethods}.ts,
// and ams/{AmsMethods,AmsItems,AmsMappings}.ts.

package tex

import (
	"strings"

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
	style = amsGenfracTrimStyle(style)
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
	if style != "" {
		styleIndex, ok := amsGenfracStyleIndex(style)
		if !ok {
			return nil, texError("BadMathStyleFor", "Bad math style for \\%s", name)
		}
		attributes := map[string]any{}
		if styleIndex == 0 {
			attributes["displaystyle"] = true
			attributes["scriptlevel"] = 0
		} else {
			attributes["displaystyle"] = false
			attributes["scriptlevel"] = styleIndex - 1
		}
		styled := node("mstyle", content)
		styled.Attributes.Set("displaystyle", attributes["displaystyle"])
		styled.Attributes.Set("scriptlevel", attributes["scriptlevel"])
		content = styled
	}
	return []*mml.Node{content}, nil
}

// Genfrac reads its style with ParseUtil.trimSpaces, including the one-space
// restoration for a terminal backslash followed by an original ASCII space.
func amsGenfracTrimStyle(raw string) string {
	style := strings.TrimFunc(raw, internalTextSpace)
	if strings.HasSuffix(style, "\\") && strings.HasSuffix(raw, " ") {
		style += " "
	}
	return style
}

// Genfrac only observes whether parseInt(style, 10) selects an index 0..3.
// Keep that decision independent of machine integer width: an accumulated
// value above 3 cannot return to range as further decimal digits are read.
func amsGenfracStyleIndex(style string) (int, bool) {
	i := 0
	negative := false
	if len(style) != 0 && (style[0] == '+' || style[0] == '-') {
		negative = style[0] == '-'
		i++
	}
	start := i
	value := 0
	for i < len(style) && style[i] >= '0' && style[i] <= '9' {
		value = value*10 + int(style[i]-'0')
		if value > 3 {
			return 0, false
		}
		i++
	}
	if i == start || (negative && value != 0) {
		return 0, false
	}
	return value, true
}

// amsFixedFencePalette ports ParseUtil.mathPalette.  MathChoice selects bigg
// in display style and big in text, script, and scriptscript styles during
// the inherited-attribute pass.
func amsFixedFencePalette(character string, class mml.TeXClass) *mml.Node {
	return amsFixedFencePaletteWithToken(character, class, token)
}

func amsFixedFencePaletteWithToken(character string, class mml.TeXClass, makeToken func(string, string) *mml.Node) *mml.Node {
	display := amsFixedFenceWithToken(character, class, "2.047em", makeToken)
	text := amsFixedFenceWithToken(character, class, "1.2em", makeToken)
	script := amsFixedFenceWithToken(character, class, "1.2em", makeToken)
	scriptScript := amsFixedFenceWithToken(character, class, "1.2em", makeToken)
	return node("MathChoice", display, text, script, scriptScript)
}

// amsFixedFence ports BaseMethods.MakeBig.  The sizes preserve JavaScript's
// source truncation after multiplying by P_HEIGHT (1.2/.85).
func amsFixedFence(character string, class mml.TeXClass, size string) *mml.Node {
	return amsFixedFenceWithToken(character, class, size, token)
}

func amsFixedFenceWithToken(character string, class mml.TeXClass, size string, makeToken func(string, string) *mml.Node) *mml.Node {
	mo := makeToken("mo", character)
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
	// HandleOperatorName creates one child parser with a copied environment.
	// Its regex and font replace the caller's choices; noAutoOP is inherited.
	operatorParser := *p
	operatorParser.activeFont = "normal"
	operatorParser.fontExplicitEmpty = false
	operatorParser.identifierPattern = identifierPatternOperator
	operatorParser.operatorLetters = true
	result, err := operatorParser.parseChild(raw)
	if err != nil {
		return nil, err
	}
	if result.Kind != "mi" {
		result = node("TeXAtom", result)
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
	pre, err := p.parseSideSetArgument(name)
	if err != nil {
		return nil, err
	}
	preScripts, preRest := splitSideSet(pre)
	post, err := p.parseSideSetArgument(name)
	if err != nil {
		return nil, err
	}
	postScripts, postRest := splitSideSet(post)
	base, err := p.parseSideSetArgument(name)
	if err != nil {
		return nil, err
	}
	mmlNode := base
	if preScripts != nil {
		if preRest != nil {
			// Copy only at the original phantom-construction point, after
			// all three arguments have registered their creation events.
			padded := node("mpadded", p.copyNode(base))
			padded.Attributes.Set("width", 0)
			phantom := node("mphantom", padded)
			if err := preScripts.ReplaceChild(phantom, preScripts.Children[0]); err != nil {
				return nil, err
			}
			refreshDynamicFlags(preScripts)
		} else {
			children := []*mml.Node{base}
			if postScripts != nil {
				children = append(children, sideSetScript(postScripts, 1), sideSetScript(postScripts, 2))
			}
			children = append(children, node("mprescripts"), sideSetScript(preScripts, 1), sideSetScript(preScripts, 2))
			mmlNode = node("mmultiscripts", children...)
			mmlNode.SetProperty("scriptalign", "left")
		}
	}
	if postScripts != nil && mmlNode == base {
		if err := postScripts.ReplaceChild(base, postScripts.Children[0]); err != nil {
			return nil, err
		}
		refreshDynamicFlags(postScripts)
		mmlNode = postScripts
	}

	// TeXAtom.appendChild appends into its inferred row and flattens an
	// inferred child there. Preserve explicit groups and reuse actual nodes.
	var children []*mml.Node
	appendPart := func(n *mml.Node) {
		if n == nil {
			return
		}
		if n.Flags.Inferred {
			children = append(children, n.Children...)
		} else {
			children = append(children, n)
		}
	}
	if preRest != nil {
		appendPart(preScripts)
		appendPart(preRest)
	}
	appendPart(mmlNode)
	appendPart(postRest)
	result := texAtom(forcedRow(children, true), mml.TeXClassOp)
	result.SetProperty("movesupsub", true)
	result.SetProperty("movablelimits", true)
	return []*mml.Node{result}, nil
}

func (p *parser) parseSideSetArgument(name string) (*mml.Node, error) {
	raw, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	return p.parseChild(raw)
}

// splitSideSet removes only the source's qualifying leading node. In the
// inferred-row branch the source checks its empty mi base, not its family.
// Keep the same non-nil rest row even if removing that child leaves it empty.
func splitSideSet(n *mml.Node) (scripts, rest *mml.Node) {
	if n == nil || n.Flags.Inferred && len(n.Children) == 0 {
		return nil, nil
	}
	if (n.Kind == "msubsup" || n.Kind == "msub" || n.Kind == "msup" || n.Kind == "mmultiscripts") && sideSetEmptyBase(n) {
		return n, nil
	}
	if !n.Flags.Inferred || len(n.Children) == 0 || !sideSetEmptyBase(n.Children[0]) {
		return nil, n
	}
	scripts = n.Children[0]
	// The source splices the slot without clearing the removed child's
	// parent. Later constructor delivery reparents the actual node.
	n.Children = n.Children[1:]
	refreshDynamicFlags(n)
	return scripts, n
}

func sideSetEmptyBase(n *mml.Node) bool {
	return n != nil && len(n.Children) != 0 && n.Children[0] != nil &&
		n.Children[0].Kind == "mi" && textContent(n.Children[0]) == ""
}

// SideSet reads literal slots 1 and 2. Only the parser's marked eager msup
// represents the source's still-generic [base, nil, sup] at this boundary.
// Genuine PrimeItem msup keeps its literal child1, as in the original method.
func sideSetScript(n *mml.Node, index int) *mml.Node {
	if origin, _ := n.Property(limitsScriptOrigin); origin == true && n.Kind == "msup" {
		if index == 1 {
			return node("none")
		}
		index = 1
	}
	if index < len(n.Children) && n.Children[index] != nil {
		return n.Children[index]
	}
	return node("none")
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
	arrow := p.operator(size.character, mml.TeXClassRel, nil)
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
