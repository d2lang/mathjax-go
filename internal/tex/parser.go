// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
//
// This file is a Go translation and modification of MathJax 3.2.2.
// Sources: ts/input/tex/{TexParser,Stack,StackItem,StackItemFactory,
// ParseMethods,ParseUtil,NodeFactory,NodeUtil}.ts and
// base/{BaseItems,BaseMappings,BaseMethods}.ts.

package tex

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/d2lang/mathjax-go/internal/mml"
	texcolor "github.com/d2lang/mathjax-go/internal/tex/extensions/color"
)

type macroDefinition struct {
	body            string
	arguments       int
	optionalDefault *string
	prefix          string
	delimiters      []string
}

type pairedDelimiter struct {
	open, close string
	body        string
	arguments   int
}

type environmentDefinition struct {
	begin, end      string
	arguments       int
	optionalDefault *string
}

type parseState struct {
	macros            map[string]macroDefinition
	pairedDelimiters  map[string]pairedDelimiter
	environments      map[string]environmentDefinition
	amsTags           *amsTagState
	augmentedPackages bool
	macroCount        int
	colorModel        *texcolor.Model
}

func newParseState() *parseState {
	s := &parseState{
		macros:           make(map[string]macroDefinition),
		pairedDelimiters: make(map[string]pairedDelimiter),
		environments:     make(map[string]environmentDefinition),
		colorModel:       texcolor.NewModel(),
	}
	for name, body := range simpleMacros {
		s.macros[name] = macroDefinition{body: body, arguments: macroArguments(body)}
	}
	return s
}

type parser struct {
	source               string
	pos                  int
	state                *parseState
	display              bool
	commandNamedFunction bool
	multiLetterFont      string
	activeFont           string
	vectorFactory        bool
	vectorFont           string
	vectorStar           bool
	vectorAlias          bool
}

// parseRow corresponds to TexParser.Parse plus the base Stack reduction.  A
// non-zero terminator is consumed.  stopRight lets a \left subparse return the
// delimiter consumed by its matching \right.
func (p *parser) parseRow(terminator byte, stopRight bool) ([]*mml.Node, string, error) {
	return p.parseRowWithInfix(terminator, stopRight, false)
}

// parseRowWithInfix retains an OverItem across denominator and declaration
// continuations of the same logical row. Real groups and subparsers enter via
// parseRow, so they cannot inherit an outer row's pending fraction.
func (p *parser) parseRowWithInfix(terminator byte, stopRight, infixPending bool) ([]*mml.Node, string, error) {
	vectorFont, vectorStar, activeFont := p.vectorFont, p.vectorStar, p.activeFont
	defer func() { p.vectorFont, p.vectorStar, p.activeFont = vectorFont, vectorStar, activeFont }()
	var nodes []*mml.Node
	var pending *pendingPrime
	var pendingFont string
	finishPrime := func() {
		if pending != nil {
			nodes = append(nodes, pending.finish())
			pending = nil
		}
	}
	// BaseMethods.NamedFn and PhysicsMethods.Expression push an FnItem.  It
	// retains the function until the next stack item determines whether an
	// ApplyFunction operator belongs between them.  Keeping this state local to
	// one parseRow call is important: closing a group finalizes the function
	// without letting it act on the first item outside that group.
	pendingFunction := false
	appendNodes := func(created []*mml.Node, namedFunction bool) {
		if len(created) == 0 {
			return
		}
		finishPrime()
		for _, n := range created {
			if pendingFont != "" {
				applyScopedMathVariant(n, pendingFont)
			}
			p.applyVectorFactory(n)
		}
		if pendingFunction {
			if !suppressesFunctionApplication(created[0]) {
				nodes = append(nodes, operator("\u2061", mml.TeXClassNone, nil))
			}
			pendingFunction = false
		}
		nodes = append(nodes, created...)
		if namedFunction {
			pendingFunction = true
		}
	}
	for p.pos < len(p.source) {
		c := p.source[p.pos]
		if terminator != 0 && c == terminator {
			finishPrime()
			p.pos++
			return nodes, "", nil
		}
		if c == '}' {
			return nil, "", texError("ExtraCloseMissingOpen", "Extra close brace or missing open brace")
		}
		if c == '%' {
			p.skipComment()
			continue
		}
		if unicode.IsSpace(p.peekRune()) {
			p.consumeRune()
			continue
		}

		if c == '\\' {
			p.pos++
			name := p.readControlSequence()
			if name == "right" {
				if !stopRight {
					return nil, "", texError("ExtraRight", "Extra \\right")
				}
				finishPrime()
				delim, err := p.readDelimiter(false)
				return nodes, delim, err
			}
			if _, registered := p.state.macros[name]; !registered && (name == "over" || name == "atop" || name == "above" || name == "choose" || name == "brace" || name == "brack") {
				finishPrime()
				fraction, right, err := p.infixFraction(name, nodes, terminator, stopRight, infixPending)
				if err != nil {
					return nil, "", err
				}
				return []*mml.Node{fraction}, right, nil
			}
			if name == "color" {
				finishPrime()
				colored, right, err := p.colorDeclaration(terminator, stopRight, infixPending)
				if err != nil {
					return nil, "", err
				}
				appendNodes(colored, false)
				return nodes, right, nil
			}
			if style, ok := styleDeclarations[name]; ok {
				finishPrime()
				rest, right, err := p.parseRowWithInfix(terminator, stopRight, infixPending)
				if err != nil {
					return nil, "", err
				}
				styled := node("mstyle", row(rest, true))
				// BaseMethods.SetStyle writes displaystyle before scriptlevel.
				// Preserve that order rather than ranging over the Go map.
				styled.Attributes.Set("displaystyle", style["displaystyle"])
				styled.Attributes.Set("scriptlevel", style["scriptlevel"])
				appendNodes([]*mml.Node{styled}, false)
				return nodes, right, nil
			}
			if size, ok := sizeDeclarations[name]; ok {
				finishPrime()
				rest, right, err := p.parseRowWithInfix(terminator, stopRight, infixPending)
				if err != nil {
					return nil, "", err
				}
				// BaseMethods.SetSize pushes a style item that scopes the
				// declaration over the remainder of the current group.
				styled := setAttributes(node("mstyle", row(rest, true)), map[string]any{"mathsize": emLength(size)})
				appendNodes([]*mml.Node{styled}, false)
				return nodes, right, nil
			}
			if variant, ok := fontDeclarations[name]; ok {
				// SetFont changes the current environment without pushing a
				// stack item, so a pending PrimeItem remains available to ^/_.
				if pending != nil {
					p.activeFont, pendingFont = variant, variant
					continue
				}
				oldFont := p.activeFont
				p.activeFont = variant
				rest, right, err := p.parseRowWithInfix(terminator, stopRight, infixPending)
				p.activeFont = oldFont
				if err != nil {
					return nil, "", err
				}
				applyScopedMathVariant(row(rest, true), variant)
				appendNodes(rest, false)
				return nodes, right, nil
			}
			if name == "limits" || name == "nolimits" {
				if pending != nil {
					return nil, "", texError("MisplacedLimits", "%s is allowed only on operators", "\\"+name)
				}
				var err error
				nodes, err = p.parseLimits(nodes, name)
				if err != nil {
					return nil, "", err
				}
				continue
			}
			p.commandNamedFunction = false
			created, err := p.command(name)
			if err != nil {
				return nil, "", err
			}
			namedFunction := p.commandNamedFunction
			p.commandNamedFunction = false
			appendNodes(created, namedFunction)
			continue
		}

		switch c {
		case '{':
			finishPrime()
			p.pos++
			contents, _, err := p.parseRow('}', false)
			if err != nil {
				return nil, "", err
			}
			appendNodes([]*mml.Node{texAtom(row(contents, true), mml.TeXClassOrd)}, false)
		case '^', '_':
			p.pos++
			var err error
			if pending == nil {
				nodes, pendingFont, err = p.attachScriptWithFont(nodes, c, pendingFont)
			} else {
				var script *mml.Node
				p.scriptInitialLookahead()
				moves, _ := pending.base.Property("movesupsub")
				var attachment *scriptAttachment
				attachment, err = prepareScriptAttachment(pending.base, c, limitsTruthy(moves))
				if err == nil {
					script, pendingFont, err = p.parseScriptArgument(attachment, pendingFont)
				}
				if err == nil {
					var result *mml.Node
					result, err = pending.attach(script, c)
					if err == nil {
						nodes = append(nodes, result)
						pending = nil
					}
				}
			}
			if err != nil {
				return nil, "", err
			}
		case '\'', 0xE2: // Only ASCII apostrophe and U+2019 enter Prime.
			if isPrimeRune(p.peekRune()) {
				p.consumeRune()
				var err error
				finishPrime()
				var base *mml.Node
				if len(nodes) == 0 {
					base = node("mi")
				} else {
					base = nodes[len(nodes)-1]
					nodes = nodes[:len(nodes)-1]
				}
				pending, err = p.startPrime(base)
				if err != nil {
					return nil, "", err
				}
			} else {
				appendNodes([]*mml.Node{p.parseCharacter()}, false)
			}
		case '&':
			return nil, "", texError("Misplaced", "Misplaced alignment tab character &")
		case '#':
			return nil, "", texError("CantUseHash1", "You can't use macro parameter character # in math mode")
		default:
			appendNodes([]*mml.Node{p.parseCharacter()}, false)
		}
	}
	finishPrime()
	if terminator != 0 {
		return nil, "", texError("ExtraOpenMissingClose", "Extra open brace or missing close brace")
	}
	if stopRight {
		return nil, "", texError("ExtraLeftMissingRight", "Extra \\left or missing \\right")
	}
	return nodes, "", nil
}

func (p *parser) parseCharacter() *mml.Node {
	r := p.consumeRune()
	if unicode.IsLetter(r) {
		text := string(r)
		if p.multiLetterFont != "" && isASCIILetter(r) {
			start := p.pos - utf8.RuneLen(r)
			for p.pos < len(p.source) && isASCIILetter(p.peekRune()) {
				p.consumeRune()
			}
			text = p.source[start:p.pos]
		}
		identifier := token("mi", text)
		if p.multiLetterFont != "" {
			if p.multiLetterFont == "normal" && len(text) > 1 {
				// ParseMethods.variable records noAutoOP as the internal autoOP
				// property, rather than allowing MmlMi to promote a roman run to
				// an operator during TeX-class assignment.
				identifier.SetProperty("autoOP", false)
			}
		}
		return ambientLiteralToken(identifier, r)
	}
	if unicode.IsDigit(r) || ((r == '.' || r == ',') && p.pos < len(p.source) && unicode.IsDigit(p.peekRune())) {
		start := p.pos - utf8.RuneLen(r)
		for p.pos < len(p.source) {
			next := p.peekRune()
			if !unicode.IsDigit(next) && next != '.' && next != ',' {
				break
			}
			p.consumeRune()
		}
		return ambientLiteralToken(token("mn", p.source[start:p.pos]), r)
	}
	// The selected package closure installs two character-map overrides ahead
	// of BaseConfiguration.Other.  Mathtools' centered-colon handler creates a
	// plain mo (the default centercolon option is false), and Physics' AutoClose
	// creates a non-stretchy closer.  Neither node is placed on Base's
	// fixStretchy list, so retaining the marker is observable in copied trees.
	if r == ':' {
		return token("mo", ":")
	}
	if r == ')' || r == ']' || r == '|' {
		return setAttributes(token("mo", string(r)), map[string]any{"stretchy": false})
	}
	text := string(r)
	if r == '-' {
		text = "−"
	} else if r == '*' {
		text = "∗"
	} else if r == '`' {
		text = "‘"
	}
	mo := ambientLiteralToken(token("mo", text), r)
	// BaseConfiguration.Other records raw operators for the fixStretchy
	// postfilter.  addNode() leaves the observable in-lists marker even after
	// the temporary fixStretchy property is removed.
	mo.SetProperty("fixStretchy", true)
	mo.SetProperty("in-lists", "fixStretchy")
	return mo
}

func (p *parser) attachScript(nodes []*mml.Node, marker byte) ([]*mml.Node, error) {
	result, _, err := p.attachScriptWithFont(nodes, marker, "")
	return result, err
}

func (p *parser) attachScriptWithFont(nodes []*mml.Node, marker byte, font string) ([]*mml.Node, string, error) {
	var base *mml.Node
	if len(nodes) == 0 {
		base = token("mi", "")
	} else {
		base = nodes[len(nodes)-1]
		nodes = nodes[:len(nodes)-1]
	}
	moves, hasMoves := base.Property("movesupsub")
	moveLimits, _ := moves.(bool)
	if value, ok := base.Attributes.Get("movesupsub"); ok {
		moveLimits, _ = value.(bool)
	}
	p.scriptInitialLookahead()
	attachment, err := prepareScriptAttachment(base, marker, moveLimits)
	if err != nil {
		return nil, font, err
	}
	script, font, err := p.parseScriptArgument(attachment, font)
	if err != nil {
		return nil, font, err
	}
	result := attachment.fill(script)
	result.Flags.Embellished = base.Flags.Embellished
	result.Flags.CoreIndex = 0
	result.SetProperty(limitsScriptOrigin, true)
	if hasMoves {
		result.SetProperty("movesupsub", moves)
	}
	return append(nodes, result), font, nil
}

func (p *parser) parseOneToken() (created []*mml.Node, err error) {
	defer func() {
		if err == nil {
			for _, n := range created {
				p.applyVectorFactory(n)
			}
		}
	}()
	p.skipSpaces()
	if p.pos >= len(p.source) {
		return nil, texError("MissingArgFor", "Missing argument")
	}
	if p.source[p.pos] == '\\' {
		p.pos++
		// A one-token argument has its own stack boundary.  A named function is
		// finalized there, so its pending FnItem must not escape into the row
		// that owns the argument.
		p.commandNamedFunction = false
		nodes, err := p.command(p.readControlSequence())
		p.commandNamedFunction = false
		return nodes, err
	}
	if p.source[p.pos] == '{' {
		p.pos++
		children, _, err := p.parseRow('}', false)
		return children, err
	}
	return []*mml.Node{p.parseCharacter()}, nil
}

func (p *parser) parseArgument(name string) (*mml.Node, error) {
	raw, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	return p.parseString(raw)
}

func (p *parser) parseString(source string) (*mml.Node, error) {
	sub := &parser{source: source, state: p.state, display: p.display,
		activeFont: p.activeFont, vectorFactory: p.vectorFactory,
		vectorFont: p.vectorFont, vectorStar: p.vectorStar, vectorAlias: p.vectorAlias}
	if p.vectorFactory || p.vectorAlias {
		sub.multiLetterFont = p.multiLetterFont
	}
	children, _, err := sub.parseRow(0, false)
	if err != nil {
		return nil, err
	}
	return row(children, true), nil
}

// parseMathFontString ports BaseMethods.MathFont's nested TexParser
// environment.  Its multiLetterIdentifiers expression consumes an ASCII
// letter run into one mi and noAutoOP prevents a multi-letter roman identifier
// from being reclassified as a named operator.
func (p *parser) parseMathFontString(source, variant string, ambientOnly bool) (*mml.Node, error) {
	sub := &parser{
		source:          source,
		state:           p.state,
		display:         p.display,
		multiLetterFont: variant,
		activeFont:      variant,
		vectorFactory:   p.vectorFactory,
		vectorFont:      p.vectorFont,
		vectorStar:      p.vectorStar,
		vectorAlias:     p.vectorAlias,
	}
	children, _, err := sub.parseRow(0, false)
	if err != nil {
		return nil, err
	}
	result := row(children, true)
	// ParseUtil.getFontDef applies the selected font to variables, digits,
	// mapped math characters, and operators created by the nested parser.
	applyFontScope(result, variant, ambientOnly)
	return result, nil
}

// suppressesFunctionApplication is BaseItems.FnItem's look-ahead decision.
// Explicit TeX spacers and BIN/REL/CLOSE/PUNCT operators terminate the FnItem
// without U+2061; every other following MML item receives ApplyFunction.
func suppressesFunctionApplication(next *mml.Node) bool {
	if next == nil {
		return false
	}
	if next.Kind == "mspace" {
		return true
	}
	if next.Kind == "mstyle" && len(next.Children) != 0 {
		contents := next.Children[0]
		if contents != nil && len(contents.Children) != 0 && contents.Children[0] != nil && contents.Children[0].Kind == "mspace" {
			return true
		}
	}
	core := next
	for core != nil && core.Kind != "mo" && core.Flags.Embellished {
		index := core.Flags.CoreIndex
		if index < 0 || index >= len(core.Children) {
			break
		}
		child := core.Children[index]
		if child == core {
			break
		}
		core = child
	}
	if core == nil || core.Kind != "mo" {
		return false
	}
	class := core.TeXClass
	return class == mml.TeXClassBin || class == mml.TeXClassRel || class == mml.TeXClassClose || class == mml.TeXClassPunct
}

func (p *parser) readArgument(name string, noneOK bool) (string, bool, error) {
	p.skipSpaces()
	if p.pos >= len(p.source) {
		if noneOK {
			return "", false, nil
		}
		return "", false, texError("MissingArgFor", "Missing argument for %s", "\\"+name)
	}
	if p.source[p.pos] == '}' {
		if noneOK {
			return "", false, nil
		}
		return "", false, texError("ExtraCloseMissingOpen", "Extra close brace or missing open brace")
	}
	if p.source[p.pos] == '{' {
		start := p.pos + 1
		end, err := p.scanBalanced(p.pos, '{', '}')
		if err != nil {
			return "", false, err
		}
		p.pos = end + 1
		return p.source[start:end], true, nil
	}
	if p.source[p.pos] == '\\' {
		start := p.pos
		p.pos++
		p.readControlSequence()
		return p.source[start:p.pos], false, nil
	}
	start := p.pos
	p.consumeRune()
	return p.source[start:p.pos], false, nil
}

func (p *parser) readBrackets(defaultValue *string) (string, bool, error) {
	p.skipSpaces()
	if p.pos >= len(p.source) || p.source[p.pos] != '[' {
		if defaultValue == nil {
			return "", false, nil
		}
		return *defaultValue, false, nil
	}
	start := p.pos + 1
	depth := 0
	for i := start; i < len(p.source); {
		switch p.source[i] {
		case '\\':
			i++
			if i < len(p.source) {
				_, n := utf8.DecodeRuneInString(p.source[i:])
				i += n
			}
		case '{':
			depth++
			i++
		case '}':
			if depth == 0 {
				return "", false, texError("ExtraCloseLooking", "Extra close brace while looking for ']'")
			}
			depth--
			i++
		case ']':
			if depth == 0 {
				p.pos = i + 1
				return p.source[start:i], true, nil
			}
			i++
		default:
			_, n := utf8.DecodeRuneInString(p.source[i:])
			i += n
		}
	}
	return "", false, texError("MissingCloseBracket", "Could not find closing ']' for argument")
}

func (p *parser) scanBalanced(at int, open, close byte) (int, error) {
	depth := 0
	for i := at; i < len(p.source); i++ {
		switch p.source[i] {
		case '\\':
			i++
		case open:
			depth++
		case close:
			depth--
			if depth == 0 {
				return i, nil
			}
		}
	}
	return 0, texError("MissingCloseBrace", "Missing close brace")
}

func (p *parser) readControlSequence() string {
	if p.pos >= len(p.source) {
		return " "
	}
	r, size := utf8.DecodeRuneInString(p.source[p.pos:])
	if !isASCIILetter(r) {
		p.pos += size
		return string(r)
	}
	start := p.pos
	for p.pos < len(p.source) {
		r, size = utf8.DecodeRuneInString(p.source[p.pos:])
		if !isASCIILetter(r) {
			break
		}
		p.pos += size
	}
	name := p.source[start:p.pos]
	if p.pos < len(p.source) && p.source[p.pos] == ' ' {
		p.pos++
	}
	return name
}

func (p *parser) readDelimiter(braceOK bool) (string, error) {
	p.skipSpaces()
	if p.pos >= len(p.source) {
		return "", texError("MissingOrUnrecognizedDelim", "Missing or unrecognized delimiter")
	}
	var name string
	escaped := false
	if p.source[p.pos] == '\\' {
		escaped = true
		p.pos++
		name = p.readControlSequence()
	} else if p.source[p.pos] == '{' && braceOK {
		raw, _, err := p.readArgument("delimiter", false)
		if err != nil {
			return "", err
		}
		name = strings.TrimSpace(raw)
	} else {
		r := p.consumeRune()
		name = string(r)
	}
	if !escaped && name == "|" {
		return "|", nil
	}
	if delim, ok := delimiterSymbols[name]; ok {
		return delim, nil
	}
	return "", texError("MissingOrUnrecognizedDelim", "Missing or unrecognized delimiter for \\%s", name)
}

func (p *parser) skipSpaces() {
	for p.pos < len(p.source) && unicode.IsSpace(p.peekRune()) {
		p.consumeRune()
	}
}

func (p *parser) skipComment() {
	for p.pos < len(p.source) {
		r := p.consumeRune()
		if r == '\n' || r == '\r' {
			return
		}
	}
}

func (p *parser) peekRune() rune {
	r, _ := utf8.DecodeRuneInString(p.source[p.pos:])
	return r
}

func (p *parser) consumeRune() rune {
	r, size := utf8.DecodeRuneInString(p.source[p.pos:])
	p.pos += size
	return r
}

func isASCIILetter(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
}

func macroArguments(body string) int {
	max := 0
	for i := 0; i+1 < len(body); i++ {
		if body[i] != '#' || body[i+1] < '1' || body[i+1] > '9' {
			continue
		}
		n := int(body[i+1] - '0')
		if n > max {
			max = n
		}
	}
	return max
}

func parseInteger(value string, id string) (int, error) {
	n, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, texError(id, "Invalid number %q", value)
	}
	return n, nil
}

func applyMathVariant(n *mml.Node, variant string) {
	n.Walk(func(current *mml.Node) bool {
		if current.Kind == "mi" || current.Kind == "mn" || current.Kind == "mo" {
			// MmlToken's kept explicit attributes override the surrounding TeX
			// font environment, as in the pinned NodeFactory/MmlToken path.
			if keep, ok := current.Attributes.GetExplicit("mjx-keep-attrs"); ok {
				if names, ok := keep.(string); ok && strings.Contains(" "+names+" ", " mathvariant ") {
					return true
				}
			}
			current.Attributes.Set("mathvariant", variant)
		}
		return true
	})
}

// MathFont and SetFont select the current lexical environment in MathJax.
// This parser lowers those environments after parsing their contents, so an
// inner environment has already resolved its tokens when an outer one runs.
// Keep that choice until compilation finishes instead of overwriting it.
// The transient property survives the parser's existing node clones and is
// removed before any MathML is returned to the renderer or caller.
const resolvedFontScope = "go-resolved-font-scope"

func applyScopedMathVariant(n *mml.Node, variant string) {
	applyFontScope(n, variant, true)
}

func applyFontScope(n *mml.Node, variant string, ambientOnly bool) {
	n.Walk(func(current *mml.Node) bool {
		if current.Kind != "mi" && current.Kind != "mn" && current.Kind != "mo" {
			return true
		}
		if eligible, _ := current.Property(ambientFontSource); ambientOnly && eligible != true {
			return true
		}
		if _, resolved := current.Property(resolvedFontScope); resolved {
			return true
		}
		if keep, ok := current.Attributes.GetExplicit("mjx-keep-attrs"); ok {
			if names, ok := keep.(string); ok && strings.Contains(" "+names+" ", " mathvariant ") {
				return true
			}
		}
		current.Attributes.Set("mathvariant", variant)
		current.SetProperty(resolvedFontScope, true)
		return true
	})
}

// Font environments apply at selected token-creation sites in ParseMethods,
// BaseConfiguration.Other and BaseMethods.Accent, not at the generic factory.
// Keep that provenance while the parser lowers MathFont/SetFont scopes.
const ambientFontSource = "go-ambient-font-source"

func ambientFontToken(n *mml.Node) *mml.Node {
	n.SetProperty(ambientFontSource, true)
	return n
}

func ambientLiteralToken(n *mml.Node, character rune) *mml.Node {
	// Other's Unicode-range variant overrides the current font. Reuse the
	// pinned table without changing token-kind or character classification.
	for _, interval := range mjOperatorRanges {
		if int(character) < interval.First {
			break
		}
		if int(character) <= interval.Last {
			if interval.HasVariant {
				n.Attributes.Set("mathvariant", interval.Variant)
				return n
			}
			break
		}
	}
	return ambientFontToken(n)
}

var styleDeclarations = map[string]map[string]any{
	"displaystyle":      {"displaystyle": true, "scriptlevel": 0},
	"textstyle":         {"displaystyle": false, "scriptlevel": 0},
	"scriptstyle":       {"displaystyle": false, "scriptlevel": 1},
	"scriptscriptstyle": {"displaystyle": false, "scriptlevel": 2},
}

var sizeDeclarations = map[string]float64{
	"tiny": 0.5, "Tiny": 0.6, "scriptsize": 0.7, "small": 0.85,
	"normalsize": 1, "large": 1.2, "Large": 1.44, "LARGE": 1.73,
	"huge": 2.07, "Huge": 2.49,
}

func emLength(value float64) string {
	if value > -0.001 && value < 0.001 {
		return "0"
	}
	formatted := strconv.FormatFloat(value, 'f', 3, 64)
	formatted = strings.TrimRight(strings.TrimRight(formatted, "0"), ".")
	return formatted + "em"
}

var fontDeclarations = map[string]string{
	"rm": "normal", "bf": "bold", "it": "-tex-mathit", "cal": "-tex-calligraphic",
	"tt": "monospace", "sf": "sans-serif",
}

func (p *parser) infixFraction(name string, left []*mml.Node, terminator byte, stopRight, infixPending bool) (*mml.Node, string, error) {
	attributes := map[string]any{}
	if name == "atop" || name == "choose" || name == "brace" || name == "brack" {
		attributes["linethickness"] = "0"
	}
	if name == "above" {
		thickness, err := p.readDimension(name)
		if err != nil {
			return nil, "", err
		}
		attributes["linethickness"] = thickness
	}
	// BaseMethods.Over reads its arguments before OverItem checks ambiguity.
	// In particular, malformed incoming above dimensions keep their own error.
	if infixPending {
		return nil, "", texError("AmbiguousUseOf", "Ambiguous use of \\%s", name)
	}
	rightNodes, right, err := p.parseRowWithInfix(terminator, stopRight, true)
	if err != nil {
		return nil, "", err
	}
	frac := setAttributes(node("mfrac", row(left, true), row(rightNodes, true)), attributes)
	if name == "choose" || name == "brace" || name == "brack" {
		open, close := "(", ")"
		if name == "brace" {
			open, close = "{", "}"
		} else if name == "brack" {
			open, close = "[", "]"
		}
		// OverItem uses numeric zero and ParseUtil.fixedFence: the fraction
		// has no null-delimiter padding, and its fences select the existing
		// bigg/big palette after inherited display/script style is known.
		frac.Attributes.Set("linethickness", 0)
		frac.SetProperty("withDelims", true)
		frac = forcedRow([]*mml.Node{
			amsFixedFencePalette(open, mml.TeXClassOpen),
			frac,
			amsFixedFencePalette(close, mml.TeXClassClose),
		}, false)
		frac.SetProperty("open", open)
		frac.SetProperty("close", close)
		frac.SetProperty("texClass", mml.TeXClassOrd)
		frac.TeXClass = mml.TeXClassOrd
	}
	return frac, right, nil
}

func fenced(open string, content *mml.Node, close string, stretchy bool) *mml.Node {
	children := make([]*mml.Node, 0, 3)
	openNode := operator(open, mml.TeXClassOpen, nil)
	openNode.Attributes.Set("fence", true)
	openNode.Attributes.Set("stretchy", stretchy)
	openNode.Attributes.Set("symmetric", true)
	children = append(children, openNode)
	if content != nil && content.Kind == "mrow" && content.Flags.Inferred {
		children = append(children, content.Children...)
	} else {
		children = append(children, content)
	}
	closeNode := operator(close, mml.TeXClassClose, nil)
	closeNode.Attributes.Set("fence", true)
	closeNode.Attributes.Set("stretchy", stretchy)
	closeNode.Attributes.Set("symmetric", true)
	children = append(children, closeNode)
	return forcedRow(children, false)
}

// leftRightFenced records the LeftItem properties that distinguish a
// \left...\right (and Matrix ArrayItem fence) row from macros that merely
// emit delimiter tokens, such as physics \qty.
func leftRightFenced(open string, content *mml.Node, close string, stretchy bool) *mml.Node {
	n := fenced(open, content, close, stretchy)
	n.SetProperty("open", open)
	n.SetProperty("close", close)
	n.SetProperty("texClass", mml.TeXClassInner)
	n.TeXClass = mml.TeXClassInner
	return n
}

func debugNode(n *mml.Node) string {
	return fmt.Sprintf("%s(%q)", n.Kind, textContent(n))
}
