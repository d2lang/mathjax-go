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
}

// parseRow corresponds to TexParser.Parse plus the base Stack reduction.  A
// non-zero terminator is consumed.  stopRight lets a \left subparse return the
// delimiter consumed by its matching \right.
func (p *parser) parseRow(terminator byte, stopRight bool) ([]*mml.Node, string, error) {
	var nodes []*mml.Node
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
				delim, err := p.readDelimiter(false)
				return nodes, delim, err
			}
			if name == "over" || name == "atop" || name == "above" || name == "choose" || name == "brace" || name == "brack" {
				fraction, right, err := p.infixFraction(name, nodes, terminator, stopRight)
				if err != nil {
					return nil, "", err
				}
				return []*mml.Node{fraction}, right, nil
			}
			if name == "color" {
				colored, right, err := p.colorDeclaration(terminator, stopRight)
				if err != nil {
					return nil, "", err
				}
				appendNodes(colored, false)
				return nodes, right, nil
			}
			if style, ok := styleDeclarations[name]; ok {
				rest, right, err := p.parseRow(terminator, stopRight)
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
				rest, right, err := p.parseRow(terminator, stopRight)
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
				rest, right, err := p.parseRow(terminator, stopRight)
				if err != nil {
					return nil, "", err
				}
				applyMathVariant(row(rest, true), variant)
				appendNodes(rest, false)
				return nodes, right, nil
			}
			if name == "limits" || name == "nolimits" {
				if len(nodes) == 0 {
					return nil, "", texError("MisplacedLimits", "Misplaced %s", "\\"+name)
				}
				nodes[len(nodes)-1].SetProperty("movesupsub", name == "limits")
				nodes[len(nodes)-1].Attributes.Set("movablelimits", name == "limits")
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
			p.pos++
			contents, _, err := p.parseRow('}', false)
			if err != nil {
				return nil, "", err
			}
			appendNodes([]*mml.Node{texAtom(row(contents, true), mml.TeXClassOrd)}, false)
		case '^', '_':
			p.pos++
			var err error
			nodes, err = p.attachScript(nodes, c)
			if err != nil {
				return nil, "", err
			}
		case '\'', 0xE2: // ASCII prime; U+2019 is handled by parseCharacter.
			if c == '\'' {
				p.pos++
				var err error
				nodes, err = p.attachPrimes(nodes)
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
	if terminator != 0 {
		return nil, "", texError("MissingCloseBrace", "Missing close brace")
	}
	if stopRight {
		return nil, "", texError("MissingRight", "Missing \\right")
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
			identifier.Attributes.Set("mathvariant", p.multiLetterFont)
			if p.multiLetterFont == "normal" && len(text) > 1 {
				// ParseMethods.variable records noAutoOP as the internal autoOP
				// property, rather than allowing MmlMi to promote a roman run to
				// an operator during TeX-class assignment.
				identifier.SetProperty("autoOP", false)
			}
		}
		return identifier
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
		return token("mn", p.source[start:p.pos])
	}
	// The selected package closure installs two character-map overrides ahead
	// of BaseConfiguration.Other.  Mathtools' centered-colon handler creates a
	// plain mo (the default centercolon option is false), and Physics' AutoClose
	// creates a non-stretchy closer.  Neither node is placed on Base's
	// fixStretchy list, so retaining the marker is observable in copied trees.
	if r == ':' {
		return token("mo", ":")
	}
	if r == ')' || r == ']' {
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
	mo := token("mo", text)
	// BaseConfiguration.Other records raw operators for the fixStretchy
	// postfilter.  addNode() leaves the observable in-lists marker even after
	// the temporary fixStretchy property is removed.
	mo.SetProperty("fixStretchy", true)
	mo.SetProperty("in-lists", "fixStretchy")
	return mo
}

func (p *parser) attachScript(nodes []*mml.Node, marker byte) ([]*mml.Node, error) {
	var base *mml.Node
	if len(nodes) == 0 {
		base = token("mi", "")
	} else {
		base = nodes[len(nodes)-1]
		nodes = nodes[:len(nodes)-1]
	}
	script, err := p.parseScriptArgument()
	if err != nil {
		return nil, err
	}
	moves, hasMoves := base.Property("movesupsub")
	moveLimits, _ := moves.(bool)
	if value, ok := base.Attributes.Get("movesupsub"); ok {
		moveLimits, _ = value.(bool)
	}
	// Braces and brackets are stacked operators even in inline math. Ordinary
	// movable-limit operators still use side scripts in that mode.
	stacked, _ := base.Property("subsupOK")
	underOver := moveLimits && (p.display || (base.Kind == "TeXAtom" && stacked == true))
	var result *mml.Node
	if marker == '_' {
		switch base.Kind {
		case "msup":
			result = node("msubsup", base.Children[0], script, base.Children[1])
		case "mover":
			result = node("munderover", base.Children[0], script, base.Children[1])
		case "msub", "msubsup", "munder", "munderover":
			return nil, texError("DoubleSubscripts", "Double subscripts: use braces to clarify")
		default:
			if underOver {
				result = node("munder", base, script)
			} else {
				result = node("msub", base, script)
			}
		}
	} else {
		switch base.Kind {
		case "msub":
			result = node("msubsup", base.Children[0], base.Children[1], script)
		case "munder":
			result = node("munderover", base.Children[0], base.Children[1], script)
		case "msup", "msubsup", "mover", "munderover":
			return nil, texError("DoubleExponent", "Double exponent: use braces to clarify")
		default:
			if underOver {
				result = node("mover", base, script)
			} else {
				result = node("msup", base, script)
			}
		}
	}
	result.Flags.Embellished = base.Flags.Embellished
	result.Flags.CoreIndex = 0
	if hasMoves {
		result.SetProperty("movesupsub", moves)
	}
	return append(nodes, result), nil
}

func (p *parser) attachPrimes(nodes []*mml.Node) ([]*mml.Node, error) {
	count := 1
	for p.pos < len(p.source) {
		p.skipSpaces()
		if p.pos >= len(p.source) || p.source[p.pos] != '\'' {
			break
		}
		p.pos++
		count++
	}
	primes := []string{"", "′", "″", "‴", "⁗"}
	text := strings.Repeat("′", count)
	if count < len(primes) {
		text = primes[count]
	}
	sup := operator(text, mml.TeXClassOrd, map[string]any{"variantForm": true})
	var base *mml.Node
	if len(nodes) == 0 {
		base = token("mi", "")
	} else {
		base = nodes[len(nodes)-1]
		nodes = nodes[:len(nodes)-1]
	}
	if base.Kind == "msup" || base.Kind == "msubsup" {
		return nil, texError("DoubleExponentPrime", "Prime causes double exponent: use braces to clarify")
	}
	if base.Kind == "msub" {
		return append(nodes, node("msubsup", base.Children[0], base.Children[1], sup)), nil
	}
	return append(nodes, node("msup", base, sup)), nil
}

func (p *parser) parseScriptArgument() (*mml.Node, error) {
	p.skipSpaces()
	if p.pos >= len(p.source) {
		return nil, texError("MissingScript", "Missing superscript or subscript argument")
	}
	if p.source[p.pos] == '{' {
		p.pos++
		children, _, err := p.parseRow('}', false)
		if err != nil {
			return nil, err
		}
		return texAtom(row(children, true), mml.TeXClassOrd), nil
	}
	created, err := p.parseOneToken()
	if err != nil {
		return nil, err
	}
	return row(created, true), nil
}

func (p *parser) parseOneToken() ([]*mml.Node, error) {
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
	sub := &parser{source: source, state: p.state, display: p.display}
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
func (p *parser) parseMathFontString(source, variant string) (*mml.Node, error) {
	sub := &parser{
		source:          source,
		state:           p.state,
		display:         p.display,
		multiLetterFont: variant,
	}
	children, _, err := sub.parseRow(0, false)
	if err != nil {
		return nil, err
	}
	result := row(children, true)
	// ParseUtil.getFontDef applies the selected font to variables, digits,
	// mapped math characters, and operators created by the nested parser.
	applyMathVariant(result, variant)
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
	"rm": "normal", "bf": "bold", "it": "italic", "cal": "-tex-calligraphic",
	"tt": "monospace", "sf": "sans-serif",
}

func (p *parser) infixFraction(name string, left []*mml.Node, terminator byte, stopRight bool) (*mml.Node, string, error) {
	attributes := map[string]any{}
	if name == "atop" || name == "choose" || name == "brace" || name == "brack" {
		attributes["linethickness"] = "0"
	}
	if name == "above" {
		p.skipSpaces()
		start := p.pos
		for p.pos < len(p.source) {
			r := p.peekRune()
			if unicode.IsSpace(r) || r == '{' || r == '}' || r == '\\' {
				break
			}
			p.consumeRune()
		}
		attributes["linethickness"] = p.source[start:p.pos]
	}
	rightNodes, right, err := p.parseRow(terminator, stopRight)
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
		frac = fenced(open, frac, close, true)
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
