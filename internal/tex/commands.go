// Copyright (c) 2009-2022 The MathJax Consortium
// Copyright (c) 2018-2022 Omar Al-Ithawi and The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
//
// This file is a Go translation and modification of MathJax 3.2.2.
// Sources: ts/input/tex/ParseUtil.ts,
// base/{BaseItems,BaseMappings,BaseMethods}.ts,
// ams/{AmsMappings,AmsMethods}.ts, braket/{BraketItems,BraketMappings,
// BraketMethods}.ts, cancel/CancelConfiguration.ts,
// color/{ColorConfiguration,ColorMethods}.ts,
// enclose/EncloseConfiguration.ts, gensymb/GensymbConfiguration.ts,
// mathtools/{MathtoolsMappings,MathtoolsMethods,MathtoolsUtil}.ts,
// mhchem/MhchemConfiguration.ts,
// newcommand/{NewcommandMappings,NewcommandMethods}.ts, and
// physics/{PhysicsItems,PhysicsMappings,PhysicsMethods}.ts.

package tex

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/d2lang/mathjax-go/internal/mhchem"
	"github.com/d2lang/mathjax-go/internal/mml"
	texcancel "github.com/d2lang/mathjax-go/internal/tex/extensions/cancel"
	texcolor "github.com/d2lang/mathjax-go/internal/tex/extensions/color"
	"github.com/d2lang/mathjax-go/internal/tex/extensions/enclose"
	"github.com/d2lang/mathjax-go/internal/tex/extensions/gensymb"
)

func (p *parser) commandNodes(name string, after **derivativeAutoOpen) ([]*mml.Node, error) {
	if definition, ok := p.state.macros[name]; ok {
		if definition.builtinNot {
			p.commandNot = true
			return nil, nil
		}
		return p.invokeMacro(name, definition)
	}
	if nodes, handled, err := p.amsTagCommand(name); handled {
		return nodes, err
	}
	if nodes, handled, err := p.mathtoolsCommand(name); handled {
		return nodes, err
	}
	if p.state.augmentedPackages {
		if nodes, handled, err := p.empheqCommand(name); handled {
			return nodes, err
		}
	}
	if nodes, handled, err := p.physicsCommand(name); handled {
		return nodes, err
	}
	if nodes, handled, err := p.amscdCommand(name); handled {
		return nodes, err
	}
	if nodes, handled, err := p.baseAMSCommand(name); handled {
		return nodes, err
	}
	// Base's CrLaTeX handler has command-map precedence over the generated
	// source-symbol fallback.  In particular, \\ is a row break, not the
	// delimiter-map backslash glyph.
	if name == "\n" || name == "\\" {
		return []*mml.Node{setAttributes(node("mspace"), map[string]any{"linebreak": "newline"})}, nil
	}
	if p.state.augmentedPackages {
		switch name {
		case "newcommand", "renewcommand", "providecommand":
			return nil, p.defineCommand(name)
		case "newenvironment", "renewenvironment":
			return nil, p.defineEnvironment(name)
		case "def", "gdef", "edef", "xdef":
			return nil, p.macroDef(name)
		case "let":
			return nil, p.letCommand(name)
		}
	}
	if definition, ok := p.state.pairedDelimiters[name]; ok {
		return p.invokePairedDelimiter(name, definition)
	}
	if symbol, ok := p.lookupExtensionSymbol(name); ok {
		return []*mml.Node{symbol}, nil
	}
	if symbol, ok := p.lookupMJSourceSymbol(name); ok {
		return []*mml.Node{symbol}, nil
	}
	if symbol, ok := identifierSymbols[name]; ok {
		n := token("mi", symbol.char)
		if symbol.attrs == nil {
			n.Attributes.Set("mathvariant", "normal")
		} else {
			setAttributes(n, symbol.attrs)
		}
		return []*mml.Node{n}, nil
	}
	if symbol, ok := operatorSymbols[name]; ok {
		n := p.operator(symbol.char, symbol.class, symbol.attrs)
		if symbol.attrs != nil {
			if moves, ok := symbol.attrs["movesupsub"]; ok {
				n.SetProperty("movesupsub", moves)
			}
		}
		return []*mml.Node{n}, nil
	}
	if function, ok := functionNames[name]; ok {
		if name == "injlim" || name == "projlim" {
			text := map[string]string{"injlim": "inj\u2006lim", "projlim": "proj\u2006lim"}[name]
			return []*mml.Node{p.namedOperator(text)}, nil
		}
		if name == "lim" || name == "liminf" || name == "limsup" || name == "max" || name == "min" || name == "sup" || name == "inf" {
			return []*mml.Node{p.namedOperator(function)}, nil
		}
		fn := token("mi", function)
		fn.TeXClass = mml.TeXClassOp
		fn.SetProperty("texClass", mml.TeXClassOp)
		p.commandNamedFunction = true
		return []*mml.Node{fn}, nil
	}

	switch name {
	case "not":
		p.commandNot = true
		return nil, nil
	case "dots":
		p.commandDots = p.startDots()
		return nil, nil
	case "mmlToken":
		return p.mmlToken(name)
	case ",":
		return []*mml.Node{spacer("0.167em")}, nil
	case ":":
		return []*mml.Node{spacer("0.222em")}, nil
	case ";":
		return []*mml.Node{spacer("0.278em")}, nil
	case "!":
		return []*mml.Node{spacer("-0.167em")}, nil
	case ">":
		return []*mml.Node{spacer("0.222em")}, nil
	case "quad":
		return []*mml.Node{spacer("1em")}, nil
	case "qquad":
		return []*mml.Node{spacer("2em")}, nil
	case "enspace":
		return []*mml.Node{spacer("0.5em")}, nil
	case "thinspace":
		return []*mml.Node{spacer("0.167em")}, nil
	case "negthinspace":
		return []*mml.Node{spacer("-0.167em")}, nil
	case "enskip":
		return []*mml.Node{spacer("0.5em")}, nil
	case " ", "space":
		return []*mml.Node{token("mtext", "\u00a0")}, nil
	case "{", "}", "$", "%", "#", "&", "_":
		return []*mml.Node{p.token("mo", name)}, nil
	case "backslash":
		return []*mml.Node{p.token("mo", "∖")}, nil
	case "|":
		return []*mml.Node{p.operator("‖", mml.TeXClassOrd, map[string]any{"fence": false, "stretchy": false})}, nil
	case "Vert":
		return []*mml.Node{p.operator("‖", mml.TeXClassOrd, map[string]any{"fence": false, "stretchy": false})}, nil
	case "vert":
		return []*mml.Node{p.operator("|", mml.TeXClassOrd, map[string]any{"fence": false, "stretchy": false})}, nil
	case "frac":
		return p.fraction(name, "")
	case "dfrac":
		return p.fraction(name, "D")
	case "tfrac":
		return p.fraction(name, "T")
	case "binom":
		return p.binomial(name, "")
	case "dbinom":
		return p.binomial(name, "D")
	case "tbinom":
		return p.binomial(name, "T")
	case "genfrac":
		return p.generalizedFraction(name)
	case "cfrac":
		return p.continuedFraction(name)
	case "sqrt":
		return p.sqrt(name)
	case "root":
		return p.root(name)
	case "mathchoice":
		return p.mathChoice(name)

	case "left":
		return p.leftRight(name)
	case "middle":
		delim, err := p.readDelimiter(name, false)
		if err != nil {
			return nil, err
		}
		// MathJax's Middle() brackets the relation delimiter with empty CLOSE
		// and OPEN atoms. Their classes remain observable to TeX spacing and
		// delimiter geometry even though both stringify as TeXAtom([]).
		close := texAtom(forcedRow(nil, true), mml.TeXClassClose)
		open := texAtom(forcedRow(nil, true), mml.TeXClassOpen)
		// The middle mo has no explicit texClass upstream.  In infix form the
		// delimiter dictionary classifies the selected fence (notably '|') as
		// ORD, so it must not introduce relation spacing.
		middle := p.token("mo", delim)
		middle.Attributes.Set("stretchy", true)
		return []*mml.Node{close, middle, open}, nil
	case "big", "Big", "bigg", "Bigg", "bigl", "Bigl", "biggl", "Biggl", "bigr", "Bigr", "biggr", "Biggr", "bigm", "Bigm", "biggm", "Biggm":
		return p.bigDelimiter(name)

	case "overline", "underline", "overbrace", "underbrace", "overrightarrow", "overleftarrow", "underrightarrow", "underleftarrow":
		return p.underOver(name)
	case "bar", "hat", "widehat", "tilde", "widetilde", "vec", "dot", "ddot", "dddot", "ddddot", "acute", "grave", "breve", "check", "mathring":
		return p.accent(name)
	case "overset", "stackrel":
		return p.overSet(name)
	case "underset":
		return p.underSet(name)
	case "overunderset":
		return p.overUnderSet(name)
	case "xrightarrow", "xleftarrow", "xleftrightarrow", "xLeftarrow", "xRightarrow", "xLeftrightarrow", "xhookleftarrow", "xhookrightarrow", "xmapsto", "xrightharpoondown", "xleftharpoondown", "xrightleftharpoons", "xrightharpoonup", "xleftharpoonup", "xleftrightharpoons":
		return p.xArrow(name)

	case "text", "textnormal":
		return p.hboxCommand(name, "", false)
	case "textrm", "textup":
		return p.hboxCommand(name, "normal", false)
	case "mbox", "hbox":
		return p.hboxCommand(name, "", true)
	case "textbf":
		return p.hboxCommand(name, "bold", false)
	case "textit":
		return p.hboxCommand(name, "italic", false)
	case "texttt":
		return p.hboxCommand(name, "monospace", false)
	case "textsf":
		return p.hboxCommand(name, "sans-serif", false)
	case "mathrm", "mathbf", "mathit", "mathsf", "mathtt", "mathbb", "mathcal", "mathscr", "mathfrak", "boldsymbol":
		return p.mathFont(name)
	case "operatorname":
		return p.operatorName(name)
	case "DeclareMathOperator":
		return nil, p.declareMathOperator(name)
	case "mathop", "mathrel", "mathbin", "mathord", "mathopen", "mathclose", "mathpunct", "mathinner":
		return p.mathClass(name)
	case "rank":
		fn := token("mi", "rank")
		fn.Attributes.Set("mathvariant", "normal")
		fn.TeXClass = mml.TeXClassOp
		fn.SetProperty("fnOP", true)
		p.commandNamedFunction = true
		return []*mml.Node{fn}, nil
	case "injlim", "projlim":
		text := map[string]string{"injlim": "inj\u2006lim", "projlim": "proj\u2006lim"}[name]
		op := p.operator(text, mml.TeXClassOp, map[string]any{"movablelimits": true})
		op.SetProperty("movesupsub", true)
		return []*mml.Node{op}, nil

	case "kern", "mkern", "hskip", "mskip", "hspace":
		return p.horizontalSpace(name)
	case "hspace*":
		return p.horizontalSpace(name)
	case "phantom", "hphantom", "vphantom":
		return p.phantom(name)
	case "smash":
		return p.smash(name)
	case "llap", "rlap", "clap", "mathllap", "mathrlap", "mathclap", "crampedllap", "crampedrlap", "crampedclap":
		return p.lap(name)
	case "cramped":
		return p.cramped(name)
	case "mathmbox":
		return p.mathMBox(name)
	case "mathmakebox":
		return p.mathMakeBox(name)
	case "raise", "lower":
		return p.raiseLower(name)
	case "rule":
		return p.rule(name)
	case "vcenter":
		arg, err := p.parseArgument(name)
		if err != nil {
			return nil, err
		}
		return []*mml.Node{texAtom(arg, mml.TeXClassVCenter)}, nil
	case "boxed":
		arg, err := p.parseArgument(name)
		if err != nil {
			return nil, err
		}
		styled := texAtom(setAttributes(node("mstyle", arg), map[string]any{"displaystyle": true, "scriptlevel": 0}), mml.TeXClassOrd)
		return []*mml.Node{setAttributes(node("menclose", styled), map[string]any{"notation": "box"})}, nil
	case "fbox":
		raw, _, err := p.readArgument(name, false)
		if err != nil {
			return nil, err
		}
		content, err := p.internalMath(raw, "", false)
		if err != nil {
			return nil, err
		}
		return []*mml.Node{setAttributes(node("menclose", content...), map[string]any{"notation": "box"})}, nil

	case "begin":
		return p.beginEnvironment(name)
	case "end":
		env, _, _ := p.readArgument(name, true)
		return nil, texError("ExtraEnd", "Extra \\end{%s}", env)
	case "displaylines":
		return p.displayLines(name)

	case "DeclarePairedDelimiter", "DeclarePairedDelimiters", "DeclarePairedDelimiterX", "DeclarePairedDelimitersX", "DeclarePairedDelimiterXPP", "DeclarePairedDelimitersXPP":
		return nil, p.declarePairedDelimiter(name)

	case "textcolor":
		return p.textColor(name)
	case "definecolor":
		return nil, p.defineColor(name)
	case "colorbox":
		return p.colorBox(name, false)
	case "fcolorbox":
		return p.colorBox(name, true)
	case "enclose":
		return p.enclose(name)
	case "cancel", "bcancel", "xcancel":
		return p.cancel(name)
	case "cancelto":
		return p.cancelTo(name)
	case "ce", "pu":
		raw, _, err := p.readArgument(name, false)
		if err != nil {
			return nil, err
		}
		mode := mhchem.ModeCE
		if name == "pu" {
			mode = mhchem.ModePU
		}
		expansion, err := mhchem.ToTeX(raw, mode)
		if err != nil {
			return nil, texError("MhchemParse", "%s", err)
		}
		p.source = p.source[:p.pos] + expansion + p.source[p.pos:]
		return nil, nil

	case "bra", "ket", "braket", "innerproduct", "ip", "outerproduct", "dyad", "ketbra", "op", "expectationvalue", "expval", "ev", "matrixelement", "matrixel", "mel":
		return p.physicsBraket(name)
	case "Bra", "Ket", "Braket", "Set", "set":
		return p.braket(name)
	case "qty", "quantity", "pqty", "bqty", "vqty", "absolutevalue", "abs", "norm", "evaluated", "eval", "order":
		return p.quantity(name)
	case "dd", "differential", "variation", "var", "dv", "derivative", "pdv", "pderivative", "partialderivative", "fdv", "fderivative", "functionalderivative":
		return p.derivative(name, after)
	case "diffd":
		return []*mml.Node{physicsDifferential("d")}, nil
	case "commutator", "comm", "anticommutator", "acomm", "poissonbracket", "pb":
		return p.commutator(name)
	case "vectorbold", "vb":
		return p.vectorBold(name)
	case "vectorarrow", "va", "vectorunit", "vu":
		return p.vectorAccent(name)
	case "vnabla":
		return []*mml.Node{p.physicsNabla()}, nil
	case "gradient", "grad", "laplacian":
		return p.operatorApplication(name, false)
	case "divergence", "div", "curl":
		return p.operatorApplication(name, true)
	case "qqtext", "qq", "qcc", "qif", "qthen", "qelse", "qotherwise", "qunless", "qgiven", "qusing", "qassume", "qsince", "qlet", "qfor", "qall", "qeven", "qodd", "qinteger", "qand", "qor", "qas", "qin":
		return p.quickQuadText(name)
	case "mqty", "matrixquantity", "pmqty", "Pmqty", "bmqty", "vmqty", "smqty", "smallmatrixquantity", "spmqty", "sPmqty", "sbmqty", "svmqty":
		return p.matrixQuantity(name)
	case "prescript":
		return p.prescript(name)
	case "nobreakspace":
		return []*mml.Node{token("mtext", "\u00a0")}, nil
	case "negmedspace":
		return []*mml.Node{spacer("-0.222em")}, nil
	case "negthickspace":
		return []*mml.Node{spacer("-0.278em")}, nil

	case "color":
		// parseRow handles the declaration form so that it can wrap the rest of
		// the current group. Reaching this path only happens in a one-token
		// context (e.g. a script), where an empty style is the source behavior.
		color, err := p.readColor(name)
		if err != nil {
			return nil, err
		}
		return []*mml.Node{setAttributes(node("mstyle", forcedRow(nil, true)), map[string]any{"mathcolor": color})}, nil
	case "noalign", "notag", "nonumber", "nolimits", "limits", "displaystyle", "textstyle", "scriptstyle", "scriptscriptstyle":
		return nil, nil
	}
	return nil, texError("UndefinedControlSequence", "Undefined control sequence \\%s", name)
}

func namedOperator(text string) *mml.Node {
	n := token("mo", text)
	n.Attributes.Set("movablelimits", true)
	n.SetProperty("movablelimits", true)
	n.SetProperty("movesupsub", true)
	n.Attributes.Set("form", "prefix")
	n.TeXClass = mml.TeXClassOp
	n.SetProperty("texClass", mml.TeXClassOp)
	return n
}

func lookupExtensionSymbol(name string) (*mml.Node, bool) {
	for _, definition := range gensymb.Configuration.Symbols {
		if definition.Name != name {
			continue
		}
		n := token(definition.TokenKind, definition.Character)
		for _, attribute := range definition.Attributes {
			n.Attributes.Set(attribute.Name, attribute.Value)
		}
		return n, true
	}
	return nil, false
}

func space(width string) *mml.Node {
	n := setAttributes(node("mspace"), map[string]any{"width": width})
	n.Flags.Spacelike = true
	return n
}

// BaseMethods.Spacer fixes scriptlevel at zero around explicit TeX spacing.
// That is observable under declarations such as \Huge and in scripts: the
// mspace inherits mathsize but does not shrink with the surrounding level.
func spacer(width string) *mml.Node {
	return setAttributes(node("mstyle", space(width)), map[string]any{"scriptlevel": 0})
}

func (p *parser) fraction(name, style string) ([]*mml.Node, error) {
	numerator, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	denominator, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	frac := node("mfrac", numerator, denominator)
	if style != "" {
		frac = setAttributes(node("mstyle", frac), styleAttributes(style))
	}
	return []*mml.Node{frac}, nil
}

func (p *parser) binomial(name, style string) ([]*mml.Node, error) {
	numerator, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	denominator, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	frac := setAttributes(node("mfrac", numerator, denominator), map[string]any{"linethickness": "0"})
	// AMS Genfrac uses string zero, fixedFence, then an optional style
	// around the entire fenced expression (including its MathChoice fences).
	frac.SetProperty("withDelims", true)
	content := forcedRow([]*mml.Node{
		p.amsFixedFencePalette("(", mml.TeXClassOpen),
		frac,
		p.amsFixedFencePalette(")", mml.TeXClassClose),
	}, false)
	content.SetProperty("open", "(")
	content.SetProperty("close", ")")
	content.SetProperty("texClass", mml.TeXClassOrd)
	content.TeXClass = mml.TeXClassOrd
	if style != "" {
		content = setAttributes(node("mstyle", content), styleAttributes(style))
	}
	return []*mml.Node{content}, nil
}

func (p *parser) generalizedFraction(name string) ([]*mml.Node, error) {
	left, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	right, _, err := p.readArgument(name, false)
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
	frac := node("mfrac", numerator, denominator)
	if thickness != "" {
		frac.Attributes.Set("linethickness", thickness)
	}
	content := frac
	if style != "" {
		content = setAttributes(node("mstyle", content), styleAttributes(style))
	}
	open, err := p.convertDelimiterArgument(left)
	if err != nil {
		return nil, err
	}
	close, err := p.convertDelimiterArgument(right)
	if err != nil {
		return nil, err
	}
	if open != "" || close != "" {
		content = p.fenced(open, content, close, true)
	}
	return []*mml.Node{content}, nil
}

func (p *parser) continuedFraction(name string) ([]*mml.Node, error) {
	align, _, err := p.readBrackets(nil)
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
	strutStyle := func(arg *mml.Node) *mml.Node {
		strut := setAttributes(node("mpadded", node("mrow")), map[string]any{
			"width": "0", "height": "8.6pt", "depth": "3pt",
		})
		styled := setAttributes(node("mstyle", texAtom(arg, mml.TeXClassOrd)), map[string]any{"displaystyle": false, "scriptlevel": 0})
		return node("mrow", strut, styled)
	}
	frac := node("mfrac", strutStyle(numerator), strutStyle(denominator))
	if align == "l" || align == "r" {
		value := map[string]string{"l": "left", "r": "right"}[align]
		frac.Attributes.Set("numalign", value)
		frac.Attributes.Set("denomalign", value)
	}
	return []*mml.Node{frac}, nil
}

func (p *parser) sqrt(name string) ([]*mml.Node, error) {
	index, present, err := p.readBrackets(nil)
	if err != nil {
		return nil, err
	}
	radical, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	if !present {
		return []*mml.Node{node("msqrt", radical)}, nil
	}
	root, err := p.parseRootIndex(index)
	if err != nil {
		return nil, err
	}
	return []*mml.Node{node("mroot", radical, root)}, nil
}

func (p *parser) root(name string) ([]*mml.Node, error) {
	start := p.pos
	depth := 0
	for p.pos < len(p.source) {
		if depth == 0 && strings.HasPrefix(p.source[p.pos:], "\\of") {
			index := p.source[start:p.pos]
			p.pos += len("\\of")
			radical, err := p.parseArgument(name)
			if err != nil {
				return nil, err
			}
			root, err := p.parseRootIndex(index)
			if err != nil {
				return nil, err
			}
			return []*mml.Node{node("mroot", radical, root)}, nil
		}
		switch p.source[p.pos] {
		case '{':
			depth++
		case '}':
			if depth > 0 {
				depth--
			}
		}
		p.consumeRune()
	}
	return nil, texError("TokenNotFoundForCommand", "Could not find \\of for \\root")
}

func (p *parser) mathChoice(name string) ([]*mml.Node, error) {
	choices := make([]*mml.Node, 4)
	for i := range choices {
		choice, err := p.parseArgument(name)
		if err != nil {
			return nil, err
		}
		choices[i] = choice
	}
	return []*mml.Node{node("MathChoice", choices...)}, nil
}

func styleAttributes(style string) map[string]any {
	switch strings.TrimSpace(style) {
	case "0", "D", "D'":
		return map[string]any{"displaystyle": true, "scriptlevel": 0}
	case "1", "T", "T'":
		return map[string]any{"displaystyle": false, "scriptlevel": 0}
	case "2", "S", "S'":
		return map[string]any{"displaystyle": false, "scriptlevel": 1}
	case "3", "SS", "SS'":
		return map[string]any{"displaystyle": false, "scriptlevel": 2}
	}
	return nil
}

func (p *parser) leftRight(name string) ([]*mml.Node, error) {
	open, err := p.readDelimiter(name, false)
	if err != nil {
		return nil, err
	}
	children, close, err := p.parseRow(0, true)
	if err != nil {
		return nil, err
	}
	return []*mml.Node{p.leftRightFenced(open, row(children, true), close, true)}, nil
}

func (p *parser) bigDelimiter(name string) ([]*mml.Node, error) {
	delim, err := p.readDelimiter(name, true)
	if err != nil {
		return nil, err
	}
	size := map[string]string{"big": "1.2em", "Big": "1.623em", "bigg": "2.047em", "Bigg": "2.470em"}
	base := strings.TrimRight(name, "lrm")
	class := mml.TeXClassOrd
	if strings.HasSuffix(name, "l") {
		class = mml.TeXClassOpen
	} else if strings.HasSuffix(name, "r") {
		class = mml.TeXClassClose
	} else if strings.HasSuffix(name, "m") {
		class = mml.TeXClassRel
	}
	mo := p.token("mo", delim)
	mo.Attributes.Set("minsize", size[base])
	mo.Attributes.Set("maxsize", size[base])
	mo.Attributes.Set("fence", true)
	mo.Attributes.Set("stretchy", true)
	mo.Attributes.Set("symmetric", true)
	return []*mml.Node{texAtom(mo, class)}, nil
}

var accentCharacters = map[string]string{
	"bar": "¯", "hat": "^", "widehat": "^", "tilde": "~", "widetilde": "~", "vec": "→",
	"dot": "˙", "ddot": "¨", "dddot": "⃛", "ddddot": "⃜",
	"acute": "´", "grave": "`", "breve": "˘", "check": "ˇ", "mathring": "˚",
}

func (p *parser) accent(name string) ([]*mml.Node, error) {
	base, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	wide := strings.HasPrefix(name, "wide")
	accent := p.operator(accentCharacters[name], mml.TeXClassOrd, map[string]any{"accent": true, "stretchy": wide})
	ambientFontToken(accent)
	// BaseMethods.Accent passes mathaccent through NodeUtil's property layer.
	// SVGmo uses that internal marker to zero the accent width and translate
	// the source glyph around its origin before the mover centers it.
	accent.SetProperty("mathaccent", true)
	// BaseMethods.Accent disables movable limits on its embellished core,
	// or on a nonembellished base carrying its own movablelimits property.
	core := base
	if base.Flags.Embellished {
		core = limitsCore(base)
	}
	if core != nil {
		movable, _ := core.Property("movablelimits")
		if core.Kind == "mo" || limitsTruthy(movable) {
			applySourceObject(core, mjSourceObject{{Name: "movablelimits", Value: false}})
		}
	}
	// BaseMethods.Accent leaves the parent implicit; its accent value is
	// inherited from the operator during MathML inheritance.
	result := node("mover", base, accent)
	return []*mml.Node{texAtom(result, mml.TeXClassOrd)}, nil
}

func (p *parser) underOver(name string) ([]*mml.Node, error) {
	base, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	under := strings.HasPrefix(name, "under")
	char := "¯"
	stretchy := true
	switch name {
	case "bar":
		stretchy = false
	case "overline", "underline":
		char = "―"
	case "overbrace", "underbrace":
		char = map[bool]string{true: "⏟", false: "⏞"}[under]
	case "overrightarrow", "underrightarrow":
		char = "→"
	case "overleftarrow", "underleftarrow":
		char = "←"
	}
	mark := p.operator(char, mml.TeXClassOrd, map[string]any{"stretchy": stretchy})
	if name == "overbrace" || name == "underbrace" {
		// The parser represents inline movable limits as side scripts. Restore
		// their under/over form for ParseUtil.underOver's brace normalization;
		// the renderer still chooses inline placement from movablelimits.
		if moves, _ := base.Property("movesupsub"); moves == true && base.Flags.Embellished {
			kind := map[string]string{"msub": "munder", "msup": "mover", "msubsup": "munderover"}[base.Kind]
			if kind != "" {
				replacement := node(kind, base.Children...)
				replacement.Attributes.SetList(base.Attributes.Explicit())
				replacement.Properties = base.Properties.Clone()
				base = replacement
			}
		}
		// Preserve ParseUtil.underOver's operator normalization: a brace must
		// not become a movable limit of a bare sum in inline math, and an
		// embellished under/over base needs its own unembellished row.
		movable, _ := base.Property("movablelimits")
		attribute, _ := base.Attributes.Get("movablelimits")
		moveLimits := propertyBool(movable)
		if base.Kind == "mo" {
			moveLimits = moveLimits || propertyBool(attribute)
			forms := operatorForms(base)
			if form, ok := base.Attributes.GetExplicit("form"); ok {
				forms = append([]string{propertyString(form)}, forms...)
			}
			if definition, ok := lookupOperatorDefinition(textContent(base), forms); ok {
				for _, property := range definition.Properties {
					if property.Name == "movablelimits" && propertyBool(property.Value) {
						moveLimits = true
					}
				}
			}
		}
		if moveLimits {
			applySourceObject(base, mjSourceObject{{Name: "movablelimits", Value: false}})
		}
		if (base.Kind == "munder" || base.Kind == "mover" || base.Kind == "munderover") && base.Flags.Embellished {
			// Stop at the operator: its text child is not the spacing target.
			core := base
			for core.Kind != "mo" && core.Flags.Embellished {
				next := core.Core()
				if next == nil || next == core {
					break
				}
				core = next
			}
			if core.Kind == "mo" {
				core.Attributes.Set("lspace", 0)
				core.Attributes.Set("rspace", 0)
			}
			empty := p.noteMO(node("mo"))
			empty.Attributes.Set("rspace", 0)
			base = node("mrow", empty, base)
		}
		// MathJax's UnderOver(..., stack=true) keeps the brace inside an OP
		// atom. A following script labels the whole brace instead of replacing
		// its accent or being rejected as a duplicate exponent/subscript.
		mark.Attributes.Set("accent", true)
		kind, accent := "mover", "accent"
		if under {
			kind, accent = "munder", "accentunder"
		}
		stack := texAtom(setAttributes(node(kind, base, mark), map[string]any{accent: true}), mml.TeXClassOp)
		stack.SetProperty("movesupsub", true)
		stack.SetProperty("subsupOK", true)
		return []*mml.Node{stack}, nil
	}
	checkMovableLimits(base)
	base = p.normalizeDecorationBase(base)
	if under {
		decoration := setAttributes(node("munder", base, mark), map[string]any{"accentunder": true})
		decoration.SetProperty("subsupOK", true)
		return []*mml.Node{decoration}, nil
	}
	decoration := setAttributes(node("mover", base, mark), map[string]any{"accent": true})
	decoration.SetProperty("subsupOK", true)
	return []*mml.Node{decoration}, nil
}

// normalizeDecorationBase preserves ParseUtil.underOver's embellished-base
// row. Side-script families and grouped bases are deliberately excluded.
func normalizeDecorationBase(base *mml.Node) *mml.Node {
	return normalizeDecorationBaseWithMO(base, func() *mml.Node { return node("mo") })
}

func normalizeDecorationBaseWithMO(base *mml.Node, makeMO func() *mml.Node) *mml.Node {
	if (base.Kind != "munder" && base.Kind != "mover" && base.Kind != "munderover") || !base.Flags.Embellished {
		return base
	}
	applySourceObject(limitsCore(base), mjSourceObject{{Name: "lspace", Value: 0}, {Name: "rspace", Value: 0}})
	empty := makeMO()
	empty.Attributes.Set("rspace", 0)
	return node("mrow", empty, base)
}

// normalizeStackBase restores the primary scripted-base representation before
// the direct Overset/Underset handlers apply the shared movable-limit policy.
func normalizeStackBase(base *mml.Node) *mml.Node {
	// Inline movable-limit scripts use a side-script representation in this
	// parser. The primary parser retains their under/over kind before the
	// Overset/Underset handler disables the base's movable limits.
	if moves, _ := base.Property("movesupsub"); moves == true && base.Flags.Embellished {
		kind := map[string]string{"msub": "munder", "msup": "mover", "msubsup": "munderover"}[base.Kind]
		if kind != "" {
			replacement := node(kind, base.Children...)
			replacement.Attributes.SetList(base.Attributes.Explicit())
			replacement.Properties = base.Properties.Clone()
			base = replacement
		}
	}

	checkMovableLimits(base)
	return base
}

func (p *parser) overSet(name string) ([]*mml.Node, error) {
	over, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	base, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	if name == "overset" {
		base = normalizeStackBase(base)
		if over.Kind == "mo" {
			over.Attributes.Set("accent", false)
		}
	}
	return []*mml.Node{node("mover", base, over)}, nil
}

func (p *parser) underSet(name string) ([]*mml.Node, error) {
	under, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	base, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	base = normalizeStackBase(base)
	if under.Kind == "mo" {
		under.Attributes.Set("accent", false)
	}
	return []*mml.Node{setAttributes(node("munder", base, under), map[string]any{"accentunder": false})}, nil
}

func (p *parser) overUnderSet(name string) ([]*mml.Node, error) {
	over, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	under, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	base, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	checkMovableLimits(base)
	if over.Kind == "mo" {
		over.Attributes.Set("accent", false)
	}
	if under.Kind == "mo" {
		under.Attributes.Set("accent", false)
	}
	return []*mml.Node{setAttributes(node("munderover", base, under, over), map[string]any{"accent": false, "accentunder": false})}, nil
}

// checkMovableLimits ports ParseUtil.checkMovableLimits. Only the direct base's
// property and, for an mo, its contextual dictionary entry participate. A
// grouped or scripted descendant must retain its own movable-limit policy.
func checkMovableLimits(base *mml.Node) {
	movable, _ := base.Property("movablelimits")
	moveLimits := movable != nil
	switch value := movable.(type) {
	case bool:
		moveLimits = value
	case string:
		moveLimits = value != ""
	case int:
		moveLimits = value != 0
	case int64:
		moveLimits = value != 0
	case float64:
		moveLimits = value != 0 && value == value // JavaScript NaN is falsy.
	}
	if base.Kind == "mo" {
		// NodeUtil.getForm uses getForms directly, without an explicit-form
		// override or a recursive core-operator traversal.
		if definition, ok := lookupOperatorDefinition(textContent(base), operatorForms(base)); ok {
			for _, property := range definition.Properties {
				if property.Name == "movablelimits" && propertyBool(property.Value) {
					moveLimits = true
				}
			}
		}
	}
	if moveLimits {
		applySourceObject(base, mjSourceObject{{Name: "movablelimits", Value: false}})
	}
}

var arrowCharacters = map[string]string{
	"xrightarrow": "→", "xleftarrow": "←", "xleftrightarrow": "↔",
	"xLeftarrow": "⇐", "xRightarrow": "⇒", "xLeftrightarrow": "⇔",
	"xhookleftarrow": "↩", "xhookrightarrow": "↪", "xmapsto": "↦",
	"xrightharpoondown": "⇁", "xleftharpoondown": "↽",
	"xrightleftharpoons": "⇌", "xrightharpoonup": "⇀",
	"xleftharpoonup": "↼", "xleftrightharpoons": "⇋",
}

func (p *parser) xArrow(name string) ([]*mml.Node, error) {
	belowRaw, hasBelow, err := p.readBrackets(nil)
	if err != nil {
		return nil, err
	}
	above, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	arrow := p.operator(arrowCharacters[name], mml.TeXClassRel, map[string]any{"stretchy": true})
	if hasBelow {
		below, err := p.parseString(belowRaw)
		if err != nil {
			return nil, err
		}
		return []*mml.Node{node("munderover", arrow, below, above)}, nil
	}
	return []*mml.Node{node("mover", arrow, above)}, nil
}

func (p *parser) textCommand(name, variant string) ([]*mml.Node, error) {
	raw, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	n := textRow(raw)
	if variant != "normal" {
		applyMathVariant(n, variant)
		n.Walk(func(current *mml.Node) bool {
			if current.Kind == "mtext" {
				current.Attributes.Set("mathvariant", variant)
			}
			return true
		})
	}
	return []*mml.Node{n}, nil
}

func textRow(raw string) *mml.Node {
	// ParseUtil.internalText creates a node, bypassing the token factory.
	return node("mtext", mml.NewText(strings.ReplaceAll(raw, "~", "\u00a0")))
}

var mathVariants = map[string]string{
	"mathrm": "normal", "mathbf": "bold", "mathit": "-tex-mathit", "mathsf": "sans-serif",
	"mathtt": "monospace", "mathbb": "double-struck", "mathcal": "-tex-calligraphic",
	"mathscr": "script", "mathfrak": "fraktur", "boldsymbol": "bold-italic",
}

func (p *parser) mathFont(name string) ([]*mml.Node, error) {
	raw, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	// The Go-only boldsymbol extension predates this ambient-font policy.
	// Keep its existing whole-token scope; the pinned package omits it.
	arg, err := p.parseMathFontString(raw, mathVariants[name], name != "boldsymbol")
	if err != nil {
		return nil, err
	}
	return []*mml.Node{texAtom(arg, mml.TeXClassOrd)}, nil
}

func (p *parser) operatorName(name string) ([]*mml.Node, error) {
	star := p.readStar()
	raw, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	n := p.operator(strings.TrimSpace(raw), mml.TeXClassOp, map[string]any{"mathvariant": "normal", "movablelimits": star})
	n.SetProperty("movesupsub", star)
	return []*mml.Node{n}, nil
}

func (p *parser) declareMathOperator(name string) error {
	star := p.readStar()
	cs, err := p.readCSNameArgument(name)
	if err != nil {
		return err
	}
	body, _, err := p.readArgument(name, false)
	if err != nil {
		return err
	}
	command := "\\operatorname"
	if star {
		command += "*"
	}
	p.state.macros[cs] = macroDefinition{body: command + "{" + body + "}"}
	return nil
}

func (p *parser) mathClass(name string) ([]*mml.Node, error) {
	arg, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	classes := map[string]mml.TeXClass{
		"mathop": mml.TeXClassOp, "mathrel": mml.TeXClassRel, "mathbin": mml.TeXClassBin,
		"mathord": mml.TeXClassOrd, "mathopen": mml.TeXClassOpen, "mathclose": mml.TeXClassClose,
		"mathpunct": mml.TeXClassPunct, "mathinner": mml.TeXClassInner,
	}
	n := texAtom(arg, classes[name])
	if name == "mathop" {
		n.SetProperty("movablelimits", true)
		n.SetProperty("movesupsub", true)
	}
	return []*mml.Node{n}, nil
}

func (p *parser) horizontalSpace(name string) ([]*mml.Node, error) {
	width, err := p.readDimension(name)
	if err != nil {
		return nil, err
	}
	return []*mml.Node{space(width)}, nil
}

func (p *parser) phantom(name string) ([]*mml.Node, error) {
	arg, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	phantom := node("mphantom", arg)
	if name == "hphantom" {
		return []*mml.Node{texAtom(setAttributes(node("mpadded", phantom), map[string]any{"height": 0, "depth": 0}), mml.TeXClassOrd)}, nil
	}
	if name == "vphantom" {
		return []*mml.Node{texAtom(setAttributes(node("mpadded", phantom), map[string]any{"width": 0}), mml.TeXClassOrd)}, nil
	}
	return []*mml.Node{texAtom(phantom, mml.TeXClassOrd)}, nil
}

func (p *parser) smash(name string) ([]*mml.Node, error) {
	option, _, err := p.readBrackets(nil)
	if err != nil {
		return nil, err
	}
	arg, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	attrs := map[string]any{}
	if option == "" || option == "t" {
		attrs["height"] = 0
	}
	if option == "" || option == "b" {
		attrs["depth"] = 0
	}
	return []*mml.Node{texAtom(setAttributes(node("mpadded", arg), attrs), mml.TeXClassOrd)}, nil
}

func (p *parser) lap(name string) ([]*mml.Node, error) {
	if name == "clap" {
		raw, _, err := p.readArgument(name, false)
		if err != nil {
			return nil, err
		}
		content := node("mstyle", node("mtext", mml.NewText(raw)))
		return []*mml.Node{setAttributes(node("mpadded", content), map[string]any{"width": 0, "lspace": "-.5width"})}, nil
	}
	arg, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	align := "left"
	if strings.Contains(name, "llap") || name == "llap" {
		align = "right"
	} else if strings.Contains(name, "clap") || name == "clap" {
		align = "center"
	}
	padded := node("mpadded", arg)
	padded.Attributes.Set("width", 0)
	if align == "right" {
		padded.Attributes.Set("lspace", "-1width")
	} else if align == "center" {
		padded.Attributes.Set("lspace", "-.5width")
	}
	if strings.HasPrefix(name, "math") || strings.HasPrefix(name, "cramped") {
		return []*mml.Node{texAtom(setAttributes(node("mstyle", padded), map[string]any{
			"data-cramped": strings.HasPrefix(name, "cramped"),
		}), mml.TeXClassOrd)}, nil
	}
	return []*mml.Node{texAtom(padded, mml.TeXClassOrd)}, nil
}

func (p *parser) cramped(name string) ([]*mml.Node, error) {
	style, _, err := p.readBrackets(nil)
	if err != nil {
		return nil, err
	}
	arg, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	attrs := map[string]any{"data-cramped": true}
	for key, value := range styleAttributes(style) {
		attrs[key] = value
	}
	return []*mml.Node{setAttributes(node("mstyle", arg), attrs)}, nil
}

func (p *parser) mathMBox(name string) ([]*mml.Node, error) {
	arg, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	children := []*mml.Node{arg}
	if arg.Kind == "mrow" && arg.Flags.Inferred {
		children = arg.Children
	}
	return []*mml.Node{node("mrow", children...)}, nil
}

func (p *parser) mathMakeBox(name string) ([]*mml.Node, error) {
	width, hasWidth, err := p.readBrackets(nil)
	if err != nil {
		return nil, err
	}
	defaultPosition := "c"
	position, _, err := p.readBrackets(&defaultPosition)
	if err != nil {
		return nil, err
	}
	arg, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	box := node("mpadded", arg)
	if hasWidth && width != "" {
		box.Attributes.Set("width", width)
	}
	if position == "c" {
		box.Attributes.Set("data-align", "center")
	} else if position == "r" {
		box.Attributes.Set("data-align", "right")
	}
	return []*mml.Node{box}, nil
}

func (p *parser) raiseLower(name string) ([]*mml.Node, error) {
	amount, err := p.readDimension(name)
	if err != nil {
		return nil, err
	}
	arg, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	if name == "lower" {
		amount = negateDimension(amount)
	}
	return []*mml.Node{setAttributes(node("mpadded", arg), map[string]any{"voffset": amount, "height": "+" + amount, "depth": "-" + amount})}, nil
}

func negateDimension(value string) string {
	if strings.HasPrefix(value, "-") {
		return strings.TrimPrefix(value, "-")
	}
	return "-" + strings.TrimPrefix(value, "+")
}

func (p *parser) rule(name string) ([]*mml.Node, error) {
	voffset, _, err := p.readBrackets(nil)
	if err != nil {
		return nil, err
	}
	widthRaw, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	heightRaw, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	attrs := map[string]any{"width": strings.TrimSpace(widthRaw), "height": strings.TrimSpace(heightRaw), "mathbackground": "currentColor"}
	if voffset != "" {
		attrs["voffset"] = voffset
	}
	return []*mml.Node{setAttributes(node("mspace"), attrs)}, nil
}

func (p *parser) readStar() bool {
	p.skipSpaces()
	if p.pos < len(p.source) && p.source[p.pos] == '*' {
		p.pos++
		return true
	}
	return false
}

func (p *parser) convertDelimiterArgument(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "." {
		return "", nil
	}
	if delim, ok := lookupDelimiter(raw); ok {
		return delim, nil
	}
	return "", texError("MissingOrUnrecognizedDelim", "Missing or unrecognized delimiter")
}

func (p *parser) declarePairedDelimiter(name string) error {
	cs, err := p.readCSNameArgument(name)
	if err != nil {
		return err
	}
	arguments := 1
	body := "#1"
	if strings.Contains(name, "X") {
		countRaw, _, err := p.readBrackets(nil)
		if err != nil {
			return err
		}
		if countRaw != "" {
			arguments, err = parseInteger(countRaw, "IllegalMacroParam")
			if err != nil {
				return err
			}
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
	if strings.Contains(name, "X") {
		body, _, err = p.readArgument(name, false)
		if err != nil {
			return err
		}
	}
	p.state.pairedDelimiters[cs] = pairedDelimiter{open: open, close: close, body: body, arguments: arguments}
	return nil
}

func (p *parser) invokePairedDelimiter(name string, definition pairedDelimiter) ([]*mml.Node, error) {
	star := p.readStar()
	size, hasSize, err := p.readBrackets(nil)
	if err != nil {
		return nil, err
	}
	args := make([]string, definition.arguments)
	for i := range args {
		args[i], _, err = p.readArgument(name, false)
		if err != nil {
			return nil, err
		}
	}
	body, err := substituteArguments(definition.body, args)
	if err != nil {
		return nil, err
	}
	content, err := p.parseString(body)
	if err != nil {
		return nil, err
	}
	open, err := p.convertDelimiterArgument(definition.open)
	if err != nil {
		return nil, err
	}
	close, err := p.convertDelimiterArgument(definition.close)
	if err != nil {
		return nil, err
	}
	stretchy := star
	if hasSize && size != "" {
		stretchy = true
	}
	return []*mml.Node{p.fenced(open, content, close, stretchy)}, nil
}

func (p *parser) readColor(name string) (string, error) {
	model, _, err := p.readBrackets(nil)
	if err != nil {
		return "", err
	}
	definition, _, err := p.readArgument(name, false)
	if err != nil {
		return "", err
	}
	value, err := p.state.colorModel.GetColor(model, definition)
	return value, translateColorError(err)
}

func translateColorError(err error) error {
	if err == nil {
		return nil
	}
	var colorError *texcolor.Error
	if errors.As(err, &colorError) {
		return &Error{ID: colorError.ID, Message: colorError.Message}
	}
	return err
}

func (p *parser) textColor(name string) ([]*mml.Node, error) {
	color, err := p.readColor(name)
	if err != nil {
		return nil, err
	}
	math, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	return []*mml.Node{setAttributes(node("mstyle", math), map[string]any{"mathcolor": color})}, nil
}

func (p *parser) defineColor(name string) error {
	cname, _, err := p.readArgument(name, false)
	if err != nil {
		return err
	}
	model, _, err := p.readArgument(name, false)
	if err != nil {
		return err
	}
	definition, _, err := p.readArgument(name, false)
	if err != nil {
		return err
	}
	return translateColorError(p.state.colorModel.DefineColor(model, cname, definition))
}

func (p *parser) colorBox(name string, framed bool) ([]*mml.Node, error) {
	frameColor := ""
	var err error
	if framed {
		frameColor, _, err = p.readArgument(name, false)
		if err != nil {
			return nil, err
		}
	}
	background, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	raw, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	content, err := p.internalMath(raw, "", false)
	if err != nil {
		return nil, err
	}
	background, err = p.state.colorModel.GetColor("named", background)
	if err != nil {
		return nil, translateColorError(err)
	}
	box := setAttributes(node("mpadded", content...), map[string]any{"mathbackground": background})
	for _, attribute := range texcolor.PaddingProperties("5px") {
		box.Attributes.Set(attribute.Name, attribute.Value)
	}
	if framed {
		frameColor, err = p.state.colorModel.GetColor("named", frameColor)
		if err != nil {
			return nil, translateColorError(err)
		}
		box.Attributes.Set("style", "border: 2px solid "+frameColor)
	}
	return []*mml.Node{box}, nil
}

func (p *parser) enclose(name string) ([]*mml.Node, error) {
	notation, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	options, _, err := p.readBrackets(nil)
	if err != nil {
		return nil, err
	}
	math, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	attrs, err := keyvalOptions(options, enclose.AllowedOptions)
	if err != nil {
		return nil, err
	}
	attrs["notation"] = strings.ReplaceAll(notation, ",", " ")
	return []*mml.Node{setAttributes(node("menclose", math), attrs)}, nil
}

func (p *parser) cancel(name string) ([]*mml.Node, error) {
	options, _, err := p.readBrackets(nil)
	if err != nil {
		return nil, err
	}
	math, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	attrs, err := keyvalOptions(options, enclose.AllowedOptions)
	if err != nil {
		return nil, err
	}
	attrs["notation"] = map[string]string{
		"cancel": texcancel.UpDiagonalStrike, "bcancel": texcancel.DownDiagonalStrike,
		"xcancel": texcancel.UpDiagonalStrike + " " + texcancel.DownDiagonalStrike,
	}[name]
	return []*mml.Node{setAttributes(node("menclose", math), attrs)}, nil
}

func (p *parser) cancelTo(name string) ([]*mml.Node, error) {
	options, _, err := p.readBrackets(nil)
	if err != nil {
		return nil, err
	}
	value, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	math, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	attrs, err := keyvalOptions(options, enclose.AllowedOptions)
	if err != nil {
		return nil, err
	}
	attrs["notation"] = texcancel.CancelToNotation
	padded := node("mpadded", value)
	for _, attribute := range texcancel.CancelToPadding {
		padded.Attributes.Set(attribute.Name, attribute.Value)
	}
	return []*mml.Node{node("msup", setAttributes(node("menclose", math), attrs), padded)}, nil
}

func keyvalOptions(raw string, allowed []string) (map[string]any, error) {
	result := make(map[string]any)
	if strings.TrimSpace(raw) == "" {
		return result, nil
	}
	allowedSet := make(map[string]bool, len(allowed))
	for _, key := range allowed {
		allowedSet[key] = true
	}
	for _, entry := range splitTopLevel(raw, ',') {
		parts := strings.SplitN(entry, "=", 2)
		key := strings.TrimSpace(parts[0])
		if !allowedSet[key] {
			return nil, texError("InvalidOption", "Invalid option '%s'", key)
		}
		value := any(true)
		if len(parts) == 2 {
			value = strings.TrimSpace(strings.Trim(parts[1], "{}"))
		}
		result[key] = value
	}
	return result, nil
}

func (p *parser) colorDeclaration() (mjSourceObject, error) {
	color, err := p.readColor("color")
	if err != nil {
		return nil, err
	}
	// Color reads and validates its arguments before pushing a StyleItem.
	return mjSourceObject{{Name: "mathcolor", Value: color}}, nil
}

func (p *parser) braket(name string) ([]*mml.Node, error) {
	raw, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	content, err := p.parseContinuationString(raw)
	if err != nil {
		return nil, err
	}
	// Braket's active vertical-bar handler creates the separator token itself;
	// unlike BaseConfiguration.Other it is not registered in fixStretchy's
	// observable in-lists metadata.  Keep the temporary fixStretchy marker so
	// the postfilter still materializes stretchy=false.
	content.Walk(func(n *mml.Node) bool {
		if n.Kind == "mo" && textContent(n) == "|" {
			n.RemoveProperty("in-lists")
		}
		return true
	})
	open, close := "⟨", "⟩"
	switch name {
	case "bra", "Bra":
		close = "|"
	case "ket", "Ket":
		open = "|"
	case "set", "Set":
		open, close = "{", "}"
	}
	stretchy := name == "Bra" || name == "Ket" || name == "Braket" || name == "Set"
	return []*mml.Node{p.leftRightFenced(open, content, close, stretchy)}, nil
}

func (p *parser) physicsBraket(name string) ([]*mml.Node, error) {
	switch name {
	case "bra":
		starBra := p.readStar()
		bra, _, err := p.readArgument(name, false)
		if err != nil {
			return nil, err
		}
		// PhysicsMethods.Bra consumes an immediately following \ket and emits a
		// single bra-ket expression.  Besides being the intended shorthand,
		// this shares the middle bar; parsing the commands independently adds a
		// duplicate bar and inter-atom space (observable in D2 label widths).
		ket, starKet, hasKet := p.physicsFollowingKet()
		if hasKet {
			macro := "\\left\\langle{" + bra + "}\\middle\\vert{" + ket + "}\\right\\rangle"
			if starBra || starKet {
				macro = "\\langle{" + bra + "}\\vert{" + ket + "}\\rangle"
			}
			return p.parseExpansion(macro)
		}
		macro := "\\left\\langle{" + bra + "}\\right\\vert{}"
		if starBra {
			macro = "\\langle{" + bra + "}\\vert{}"
		}
		return p.parseExpansion(macro)
	case "ket":
		star := p.readStar()
		ket, _, err := p.readArgument(name, false)
		if err != nil {
			return nil, err
		}
		macro := "\\left\\vert{" + ket + "}\\right\\rangle"
		if star {
			macro = "\\vert{" + ket + "}\\rangle"
		}
		return p.parseExpansion(macro)
	case "braket", "innerproduct", "ip":
		star := p.readStar()
		left, _, err := p.readArgument(name, false)
		if err != nil {
			return nil, err
		}
		right := ""
		p.skipSpaces()
		if p.pos < len(p.source) && p.source[p.pos] == '{' {
			right, _, err = p.readArgument(name, false)
			if err != nil {
				return nil, err
			}
		} else {
			right = left
		}
		macro := "\\left\\langle{" + left + "}\\middle\\vert{" + right + "}\\right\\rangle"
		if star {
			macro = "\\langle{" + left + "}\\vert{" + right + "}\\rangle"
		}
		nodes, err := p.parseExpansion(macro)
		if err != nil {
			return nil, err
		}
		// Physics' active vertical-bar character handler creates AutoClose
		// tokens directly, so bars occurring inside braket arguments are not
		// members of Base's fixStretchy list.  The middle \vert already lacks
		// this marker; removing it here restores that package-wide distinction.
		for _, n := range nodes {
			n.Walk(func(current *mml.Node) bool {
				if current.Kind == "mo" && textContent(current) == "|" {
					current.RemoveProperty("in-lists")
				}
				return true
			})
		}
		return nodes, nil
	case "outerproduct", "dyad", "ketbra", "op":
		ket, _, err := p.readArgument(name, false)
		if err != nil {
			return nil, err
		}
		bra, _, err := p.readArgument(name, true)
		if err != nil {
			return nil, err
		}
		if bra == "" {
			bra = ket
		}
		left, err := p.parseString(ket)
		if err != nil {
			return nil, err
		}
		right, err := p.parseString(bra)
		if err != nil {
			return nil, err
		}
		return []*mml.Node{p.fenced("|", left, "⟩", true), p.fenced("⟨", right, "|", true)}, nil
	case "expectationvalue", "expval", "ev":
		raw, _, err := p.readArgument(name, false)
		if err != nil {
			return nil, err
		}
		arg, err := p.parseString(raw)
		if err != nil {
			return nil, err
		}
		return []*mml.Node{p.fenced("⟨", texAtom(arg, mml.TeXClassOrd), "⟩", true)}, nil
	case "matrixelement", "matrixel", "mel":
		bra, _, err := p.readArgument(name, false)
		if err != nil {
			return nil, err
		}
		operatorRaw, _, err := p.readArgument(name, false)
		if err != nil {
			return nil, err
		}
		ket, _, err := p.readArgument(name, false)
		if err != nil {
			return nil, err
		}
		content, err := p.parseString(bra + "\\middle|" + operatorRaw + "\\middle|" + ket)
		if err != nil {
			return nil, err
		}
		return []*mml.Node{p.fenced("⟨", content, "⟩", true)}, nil
	}
	return nil, nil
}

func (p *parser) parseExpansion(source string) ([]*mml.Node, error) {
	return p.parseExpansionWithStack(source, nil)
}

func (p *parser) parseContinuationExpansion(source string) ([]*mml.Node, error) {
	return p.parseExpansionWithStack(source, p.ensureStackGlobal())
}

func (p *parser) parseExpansionWithStack(source string, global *parserStackGlobal) ([]*mml.Node, error) {
	parsed, err := p.parseStringWithStack(source, global)
	if err != nil {
		return nil, err
	}
	if parsed.Kind == "mrow" && parsed.Flags.Inferred {
		return parsed.Children, nil
	}
	return []*mml.Node{parsed}, nil
}

func (p *parser) quantity(name string) ([]*mml.Node, error) {
	delimiters := map[string][2]string{
		"qty": {"(", ")"}, "quantity": {"(", ")"}, "pqty": {"(", ")"},
		"bqty": {"[", "]"}, "vqty": {"|", "|"}, "absolutevalue": {"|", "|"},
		"abs": {"|", "|"}, "norm": {"‖", "‖"}, "evaluated": {".", "|"},
		"eval": {".", "|"}, "order": {"O(", ")"},
	}
	pair := delimiters[name]
	if pair[0] == "O(" {
		arg, err := p.parseArgument(name)
		if err != nil {
			return nil, err
		}
		return []*mml.Node{token("mi", "O"), p.physicsFenced("(", arg, ")", true)}, nil
	}
	return p.quantityWithDelimiters(name, pair[0], pair[1])
}

func (p *parser) quantityWithDelimiters(name, open, close string) ([]*mml.Node, error) {
	// Argument-free Quantity leaves an unsupported star for ordinary parsing.
	// Its empty fallback uses ParseUtil.fenced, not the AutoOpen row policy.
	if name == "qty" || name == "quantity" {
		p.skipSpaces()
		if p.pos < len(p.source) && p.source[p.pos] == '*' {
			return []*mml.Node{p.leftRightFenced(open, forcedRow(nil, false), close, true)}, nil
		}
	}
	star := p.readStar()
	p.skipSpaces()
	if p.pos >= len(p.source) {
		return []*mml.Node{p.physicsFenced(open, forcedRow(nil, true), close, !star)}, nil
	}
	var raw string
	var err error
	continuation := false
	if closing, ok := map[byte]byte{'(': ')', '[': ']', '|': '|'}[p.source[p.pos]]; ok {
		continuation = true
		p.pos++
		raw, err = p.readUpToByte(closing)
	} else {
		raw, _, err = p.readArgument(name, false)
	}
	if err != nil {
		return nil, err
	}
	// Quantity's raw fences use AutoOpen on the caller; its braced argument
	// is parsed by a genuine child TexParser.
	var content *mml.Node
	if continuation {
		content, err = p.parseContinuationString(raw)
	} else {
		content, err = p.parseString(raw)
	}
	if err != nil {
		return nil, err
	}
	return []*mml.Node{p.physicsFenced(open, content, close, !star)}, nil
}

func physicsFenced(open string, content *mml.Node, close string, stretchy bool) *mml.Node {
	row := fenced(open, content, close, stretchy)
	// Physics AutoOpen reduces through its stack item to an INNER-class row,
	// but does not install an explicit texClass property on the MML node.
	row.TeXClass = mml.TeXClassInner
	return row
}

func (p *parser) readUpToByte(close byte) (string, error) {
	start := p.pos
	braceDepth := 0
	fenceDepth := 0
	open := byte(0)
	switch close {
	case ')':
		open = '('
	case ']':
		open = '['
	case '|':
		open = '|'
	}
	for p.pos < len(p.source) {
		c := p.source[p.pos]
		if c == '\\' {
			p.pos++
			p.readControlSequence()
			continue
		}
		if c == '{' {
			braceDepth++
		} else if c == '}' && braceDepth > 0 {
			braceDepth--
		} else if braceDepth == 0 {
			if c == open && open != close {
				fenceDepth++
			} else if c == close {
				if fenceDepth == 0 {
					result := p.source[start:p.pos]
					p.pos++
					return result, nil
				}
				fenceDepth--
			}
		}
		p.consumeRune()
	}
	return "", texError("TokenNotFoundForCommand", "Could not find closing delimiter")
}

// physicsFollowingKet ports the lookahead in PhysicsMethods.Bra.  The source
// only commits the lookahead when the ket has a braced argument; otherwise the
// parser position is restored so the ordinary \ket handler sees it.
func (p *parser) physicsFollowingKet() (ket string, star bool, ok bool) {
	start := p.pos
	p.skipSpaces()
	if p.pos >= len(p.source) || p.source[p.pos] != '\\' {
		p.pos = start
		return "", false, false
	}
	p.pos++
	if p.readControlSequence() != "ket" {
		p.pos = start
		return "", false, false
	}
	star = p.readStar()
	p.skipSpaces()
	if p.pos >= len(p.source) || p.source[p.pos] != '{' {
		p.pos = start
		return "", false, false
	}
	ket, _, err := p.readArgument("ket", true)
	if err != nil {
		p.pos = start
		return "", false, false
	}
	return ket, star, true
}

func physicsDifferential(character string) *mml.Node {
	d := token("mi", character)
	if character != "d" {
		// PhysicsMethods.Differential reparses its operator string.  The
		// variation operator is \delta, which is an ordinary mi rather than a
		// nested TeXAtom; \diffd below expands through {\rm d} and does retain
		// that atom wrapper.
		return d
	}
	d.Attributes.Set("mathvariant", "normal")
	return texAtom(d, mml.TeXClassOrd)
}

func physicsNabla() *mml.Node {
	return physicsNablaWithOperator(operator)
}

func physicsNablaWithOperator(makeOperator func(string, mml.TeXClass, map[string]any) *mml.Node) *mml.Node {
	inner := texAtom(makeOperator("∇", mml.TeXClassOrd, map[string]any{"mathvariant": "bold"}), mml.TeXClassOrd)
	return texAtom(inner, mml.TeXClassOrd)
}

func (p *parser) commutator(name string) ([]*mml.Node, error) {
	star := p.readStar()
	p.skipSpaces()
	big := ""
	if p.pos < len(p.source) && p.source[p.pos] == '\\' {
		p.pos++
		big = p.readControlSequence()
		switch big {
		case "big", "Big", "bigg", "Bigg":
		default:
			return nil, texError("MissingArgFor", "Missing argument for %s", "\\"+name)
		}
		p.skipSpaces()
	}
	if p.pos >= len(p.source) || p.source[p.pos] != '{' {
		return nil, texError("MissingArgFor", "Missing argument for %s", "\\"+name)
	}
	leftRaw, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	rightRaw, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	open, close := "[", "]"
	switch name {
	case "anticommutator", "acomm", "poissonbracket", "pb":
		open, close = "\\{", "\\}"
	}
	// PhysicsMethods.Commutator parses both raw arguments in one child
	// expression. Their comma is a real token in the same lexical row.
	argument := leftRaw + "," + rightRaw
	switch {
	case star:
		argument = open + " " + argument + " " + close
	case big != "":
		argument = "\\" + big + "l" + open + " " + argument + " " + "\\" + big + "r" + close
	default:
		argument = "\\left" + open + " " + argument + " " + "\\right" + close
	}
	parsed, err := p.parseChild(argument)
	if err != nil {
		return nil, err
	}
	return unwrapInferred(parsed), nil
}

func (p *parser) vectorBold(name string) ([]*mml.Node, error) {
	star := p.readStar()
	raw, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	arg, err := p.parseVectorString(raw, star)
	if err != nil {
		return nil, err
	}
	return unwrapInferred(arg), nil
}

func (p *parser) vectorAccent(name string) ([]*mml.Node, error) {
	star := p.readStar()
	raw, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	accent := "vec"
	if name == "vectorunit" || name == "vu" {
		accent = "hat"
	}
	bold := "vb"
	if star {
		bold += "*"
	}
	expansion, err := substituteMacroArguments("\\"+accent+"{\\"+bold+"{#1}}", []string{raw})
	if err != nil {
		return nil, err
	}
	// StarMacro checks the complete replacement plus unparsed input before
	// charging. The existing synthetic expansion still owns a separate source
	// and cursor; this validates the joined buffer without installing it.
	if _, err := macroAddArgs(expansion, p.source[p.pos:], maxMacroBuffer); err != nil {
		return nil, err
	}
	p.state.macroCount++
	if p.state.macroCount > maxMacros {
		return nil, texError("MaxMacroSub1", "MathJax maximum macro substitution count exceeded; is here a recursive macro call?")
	}
	// PhysicsMethods.StarMacro preserves the ordinary accent handler around
	// the vector argument; the accent is outside VectorBold's font reset.
	// This synthetic continuation carries the charged caller budget. Only its
	// genuine ParseArg and VectorBold children receive a fresh count.
	sub := &parser{source: expansion, state: p.state, stackGlobal: p.ensureStackGlobal(), display: p.display,
		multiLetterFont: p.multiLetterFont, activeFont: p.activeFont,
		identifierPattern: p.identifierPattern, operatorLetters: p.operatorLetters,
		noAutoOP: p.noAutoOP, fontExplicitEmpty: p.fontExplicitEmpty,
		vectorFactory: p.vectorFactory, vectorFont: p.vectorFont, vectorStar: p.vectorStar, vectorAlias: true,
		genfracPalette: p.genfracPalette, starMacroChildren: true,
		derivativeChildren: p.derivativeChildren}
	children, _, err := sub.parseRow(0, false)
	if err != nil {
		return nil, err
	}
	return children, nil
}

func (p *parser) operatorApplication(name string, vector bool) ([]*mml.Node, error) {
	var operatorNode *mml.Node
	var prefix []*mml.Node
	if vector {
		operatorNode = p.physicsNabla()
		char := "⋅"
		attributes := map[string]any{"mathvariant": "bold"}
		if name == "curl" {
			char = "×"
			attributes = nil
		}
		// PhysicsMappings maps \vdot to a bold U+22C5, whereas
		// \crossproduct is an unmodified U+00D7.  The distinction changes the
		// SVG glyph width (and therefore the fenced operand's x position).
		prefix = []*mml.Node{operatorNode, p.operator(char, operatorClass(char), attributes)}
	} else if name == "laplacian" {
		operatorNode = node("msup", setAttributes(token("mi", "∇"), map[string]any{"mathvariant": "normal"}), token("mn", "2"))
	} else {
		operatorNode = p.physicsNabla()
	}
	if prefix == nil {
		prefix = []*mml.Node{operatorNode}
	}
	apply := p.operator("\u2061", mml.TeXClassNone, nil)
	p.skipSpaces()
	if p.pos >= len(p.source) {
		return prefix, nil
	}
	var raw string
	var err error
	if p.source[p.pos] == '(' {
		p.pos++
		raw, err = p.readUpToByte(')')
	} else if p.source[p.pos] == '[' {
		p.pos++
		raw, err = p.readUpToByte(']')
	} else {
		raw, _, err = p.readArgument(name, false)
	}
	if err != nil {
		return nil, err
	}
	arg, err := p.parseContinuationString(raw)
	if err != nil {
		return nil, err
	}
	if vector {
		return append(prefix, p.fenced("(", arg, ")", true)), nil
	}
	return append(prefix, apply, p.fenced("(", arg, ")", true)), nil
}

func (p *parser) quickQuadText(name string) ([]*mml.Node, error) {
	star := p.readStar()
	_ = star
	defaults := map[string]string{
		"qcc": "c.c.", "qif": "if", "qthen": "then", "qelse": "else", "qotherwise": "otherwise",
		"qunless": "unless", "qgiven": "given", "qusing": "using", "qassume": "assume", "qsince": "since",
		"qlet": "let", "qfor": "for", "qall": "all", "qeven": "even", "qodd": "odd", "qinteger": "integer",
		"qand": "and", "qor": "or", "qas": "as", "qin": "in",
	}
	text := defaults[name]
	if text == "" {
		var err error
		text, _, err = p.readArgument(name, false)
		if err != nil {
			return nil, err
		}
	}
	return []*mml.Node{spacer("1em"), node("mtext", mml.NewText(text)), spacer("1em")}, nil
}

func (p *parser) prescript(name string) ([]*mml.Node, error) {
	sup, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	sub, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	base, err := p.parseArgument(name)
	if err != nil {
		return nil, err
	}
	result := node("mmultiscripts", base, node("mprescripts"), sub, sup)
	result.SetProperty("fixPrescript", true)
	return []*mml.Node{result}, nil
}

func unwrapInferred(n *mml.Node) []*mml.Node {
	if n != nil && n.Kind == "mrow" && n.Flags.Inferred {
		return n.Children
	}
	return []*mml.Node{n}
}

func (p *parser) derivative(name string, after **derivativeAutoOpen) ([]*mml.Node, error) {
	if name == "dd" || name == "differential" || name == "variation" || name == "var" {
		power, hasPower, err := p.readBrackets(nil)
		if err != nil {
			return nil, err
		}
		diffText := map[bool]string{true: "δ", false: "d"}[name == "variation" || name == "var"]
		diff := physicsDifferential(diffText)
		if hasPower {
			exponent, err := p.parseString(power)
			if err != nil {
				return nil, err
			}
			diff = node("msup", diff, exponent)
		}
		p.skipSpaces()
		if p.pos < len(p.source) && p.source[p.pos] == '(' {
			p.pos++
			raw, err := p.readUpToByte(')')
			if err != nil {
				return nil, err
			}
			arg, err := p.parseContinuationString(raw)
			if err != nil {
				return nil, err
			}
			parens := p.fenced("(", arg, ")", true)
			// Physics' AutoOpen item reduces to an INNER row through TeX-class
			// processing, without materializing a texClass node property.
			parens.TeXClass = mml.TeXClassInner
			return []*mml.Node{diff, parens}, nil
		}
		raw, braced, err := p.readArgument(name, true)
		if err != nil {
			return nil, err
		}
		if raw == "" {
			return []*mml.Node{diff}, nil
		}
		arg, err := p.parseString(raw)
		if err != nil {
			return nil, err
		}
		// Differential builds one temporary TeX parser over op+argument.  Its
		// top-level inferred row therefore contains the argument's children
		// directly; keeping the nested inferred row adds an extra atom/spacing
		// layer (most visibly for \var{...}).
		combinedNodes := []*mml.Node{diff}
		combinedNodes = append(combinedNodes, unwrapInferred(arg)...)
		combined := row(combinedNodes, true)
		if braced {
			return []*mml.Node{texAtom(combined, mml.TeXClassOp)}, nil
		}
		return []*mml.Node{combined}, nil
	}
	star := p.readStar()
	order, hasOrder, err := p.readBrackets(nil)
	if err != nil {
		var failure *Error
		if errors.As(err, &failure) && failure.ID == "MissingCloseBracket" {
			return nil, texError(failure.ID, "Could not find closing ']' for argument to %s", "\\"+name)
		}
		return nil, err
	}
	firstRaw, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	args := []string{firstRaw}
	argMax := 2
	op := "\\diffd"
	switch name {
	case "pdv", "pderivative", "partialderivative":
		argMax = 3
		op = "\\partial"
	case "fdv", "fderivative", "functionalderivative":
		op = "\\delta"
	}
	for {
		p.skipSpaces()
		if p.pos >= len(p.source) || p.source[p.pos] != '{' || len(args) == argMax {
			break
		}
		arg, _, err := p.readArgument(name, false)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}
	ignore := false
	power1, power2 := " ", " "
	if argMax > 2 && len(args) > 2 {
		power1 = "^{" + strconv.Itoa(len(args)-1) + "}"
		ignore = true
	} else if hasOrder {
		ignore = argMax > 2 && len(args) > 1
		power1 = "^{" + order + "}"
		power2 = power1
	}
	frac := "\\frac"
	if star {
		frac = "\\flatfrac"
	}
	first, second := "", args[0]
	if len(args) > 1 {
		first, second = args[0], args[1]
	}
	rest := ""
	for i := 2; i < len(args) && args[i] != ""; i++ {
		rest += op + " " + args[i]
	}
	// PhysicsMethods.Derivative reparses this literal source. In particular,
	// the denominator power belongs after the raw variable, and each order
	// occurrence is parsed independently by the actual fraction handler.
	expansion := frac + "{" + op + power1 + first + "}" +
		"{" + op + " " + second + power2 + " " + rest + "}"
	parsed, err := p.parseChild(expansion)
	if err != nil {
		return nil, err
	}
	// Push/PushAll forwards the actual parser result, including zero or many
	// children from registered frac/flatfrac overrides. Only then may its
	// recipient activate AutoOpen on the original parser.
	*after = &derivativeAutoOpen{ignore: ignore}
	return unwrapInferred(parsed), nil
}

func (p *parser) matrixQuantity(name string) ([]*mml.Node, error) {
	star := p.readStar()
	p.skipSpaces()
	var raw string
	var err error
	autoOpen, autoClose := "", ""
	if p.pos < len(p.source) && p.source[p.pos] == '(' {
		p.pos++
		raw, err = p.readUpToByte(')')
		autoOpen, autoClose = "(", ")"
	} else {
		raw, _, err = p.readArgument(name, false)
	}
	if err != nil {
		return nil, err
	}
	small := strings.HasPrefix(name, "s") || strings.Contains(name, "small")
	table, err := p.parseTable(raw, map[bool]string{true: "S", false: "T"}[small])
	if err != nil {
		return nil, err
	}
	open, close := autoOpen, autoClose
	if strings.Contains(name, "pmqty") || name == "Pmqty" || name == "spmqty" || name == "sPmqty" {
		open, close = "(", ")"
	} else if strings.Contains(name, "bmqty") {
		open, close = "[", "]"
	} else if strings.Contains(name, "vmqty") {
		open, close = "|", "|"
	} else if star {
		open, close = "‖", "‖"
	}
	if open != "" {
		table = p.fenced(open, table, close, true)
	}
	return []*mml.Node{table}, nil
}

func keyvalString(raw string) map[string]string {
	result := make(map[string]string)
	for _, entry := range splitTopLevel(raw, ',') {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) == 2 {
			result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return result
}

func numberString(value any) string { return fmt.Sprint(value) }

func parseFloat(value string) (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(value), 64)
}
