// Copyright (c) 2018-2022 MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.
//
// Source: ts/input/tex/physics/PhysicsMethods.ts and PhysicsMappings.ts.

package tex

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// physicsCommand is the narrow Physics command hook for the central parser.
// It must run before built-in function-name handling, since Physics overrides
// names such as \sin with its optional-exponent and automatic-fence method.
func (p *parser) physicsCommand(name string) (nodes []*mml.Node, handled bool, err error) {
	switch name {
	case "sin", "sinh", "arcsin", "asin", "cos", "cosh", "arccos", "acos", "tan", "tanh", "arctan", "atan",
		"csc", "csch", "arccsc", "acsc", "sec", "sech", "arcsec", "asec", "cot", "coth", "arccot", "acot",
		"exp", "log", "ln", "det", "Pr", "tr", "trace", "Tr", "Trace", "erf":
		nodes, err = p.physicsExpression(name)
	case "evaluated", "eval":
		nodes, err = p.physicsEval(name)
	case "outerproduct", "dyad", "ketbra", "op":
		nodes, err = p.physicsKetBra(name)
	case "matrixelement", "matrixel", "mel":
		nodes, err = p.physicsMatrixElement(name)
	case "identitymatrix", "imat", "xmatrix", "xmat", "zeromatrix", "zmat", "paulimatrix", "pmat", "diagonalmatrix", "dmat", "antidiagonalmatrix", "admat":
		var expansion string
		expansion, err = p.physicsMatrixExpansion(name)
		if err == nil && expansion != "" {
			p.source = p.source[:p.pos] + expansion + p.source[p.pos:]
		}
	default:
		return nil, false, nil
	}
	return nodes, true, err
}

// physicsEnvironment owns only matrix environments whose body contains a
// Physics matrix generator.  The hook is required because the local table
// parser splits rows and cells before parsing commands, while MathJax's stack
// parser lets \imat, \xmat, \pmat, and \dmat inject alignment separators.
func (p *parser) physicsEnvironment(name string) (nodes []*mml.Node, handled bool, err error) {
	if !physicsMatrixEnvironment(name) || !physicsBodyHasMatrixGenerator(p.source[p.pos:]) {
		return nil, false, nil
	}
	columnSpec := ""
	if strings.HasSuffix(name, "*") {
		columnSpec, _, err = p.readBrackets(nil)
		if err != nil {
			return nil, true, err
		}
	}
	body, err := p.captureEnvironment(name)
	if err != nil {
		return nil, true, err
	}
	body, err = p.physicsRewriteMatrixBody(body)
	if err != nil {
		return nil, true, err
	}
	style := "T"
	if strings.Contains(name, "small") {
		style = "S"
	}
	table, err := p.parseTable(body, style)
	if err != nil {
		return nil, true, err
	}
	applyColumnSpec(table, columnSpec)
	open, close := matrixDelimiters(name)
	if open != "" || close != "" {
		table = p.leftRightFenced(open, table, close, true)
	}
	return []*mml.Node{table}, true, nil
}

func (p *parser) physicsExpression(name string) ([]*mml.Node, error) {
	id := name
	if name == "trace" {
		id = "tr"
	} else if name == "Trace" {
		id = "Tr"
	}
	allowExponent := true
	switch name {
	case "exp", "det", "Pr", "tr", "trace", "Tr", "Trace", "erf":
		allowExponent = false
	}
	exponent := ""
	if allowExponent {
		var err error
		exponent, _, err = p.readBrackets(nil)
		if err != nil {
			return nil, err
		}
	}
	function := token("mi", id)
	function.TeXClass = mml.TeXClassOp
	function.SetProperty("texClass", mml.TeXClassOp)
	var base *mml.Node = function
	if exponent != "" {
		sup, err := p.parseString(exponent)
		if err != nil {
			return nil, err
		}
		base = node("msup", function, sup)
	}
	p.skipSpaces()
	if p.pos >= len(p.source) || p.source[p.pos] != '(' {
		// PhysicsMethods.Expression pushes the operator as an FnItem.  The
		// surrounding parser row decides whether a following item requires
		// ApplyFunction; a group boundary finalizes the function without it.
		p.commandNamedFunction = true
		return []*mml.Node{base}, nil
	}
	p.pos++
	raw, err := p.readUpToByte(')')
	if err != nil {
		return nil, err
	}
	content, err := p.parseString(raw)
	if err != nil {
		return nil, err
	}
	// Physics' AutoOpen item emits fixed fence tokens around the argument; it
	// does not mark the resulting mrow as a \left...\right INNER atom.  That
	// keeps the function-to-opening-delimiter spacing at zero.
	return []*mml.Node{base, p.operator("\u2061", mml.TeXClassNone, nil), p.fenced("(", content, ")", true)}, nil
}

func (p *parser) physicsEval(name string) ([]*mml.Node, error) {
	star := p.readStar()
	p.skipSpaces()
	if p.pos >= len(p.source) {
		return nil, texError("MissingArgFor", "Missing argument for \\%s", name)
	}
	var raw string
	var err error
	switch p.source[p.pos] {
	case '{':
		raw, _, err = p.readArgument(name, false)
	case '(':
		p.pos++
		raw, err = p.readUpToByte(')')
	case '[':
		p.pos++
		raw, err = p.readUpToByte(']')
	default:
		return nil, texError("MissingArgFor", "Missing argument for \\%s", name)
	}
	if err != nil {
		return nil, err
	}
	if star {
		raw = "\\smash{" + raw + "}"
	}
	return p.parseExpansion("\\left. " + raw + " \\vphantom{\\int}\\right|")
}

func (p *parser) physicsKetBra(name string) ([]*mml.Node, error) {
	star := p.readStar()
	ket, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	bra, present, err := p.readArgument(name, true)
	if err != nil {
		return nil, err
	}
	if !present {
		bra = ket
	}
	macro := "\\left\\vert{" + ket + "}\\middle\\rangle\\!\\middle\\langle{" + bra + "}\\right\\vert"
	if star {
		macro = "\\vert{" + ket + "}\\rangle\\!\\langle{" + bra + "}\\vert"
	}
	return p.parseExpansion(macro)
}

func (p *parser) physicsMatrixElement(name string) ([]*mml.Node, error) {
	star1 := p.readStar()
	star2 := star1 && p.readStar()
	args := make([]string, 3)
	for i := range args {
		var err error
		args[i], _, err = p.readArgument(name, false)
		if err != nil {
			return nil, err
		}
	}
	macro := "\\left\\langle{" + args[0] + "}\\right\\vert{" + args[1] + "}\\left\\vert{" + args[2] + "}\\right\\rangle"
	if star1 && star2 {
		macro = "\\left\\langle{" + args[0] + "}\\middle\\vert{" + args[1] + "}\\middle\\vert{" + args[2] + "}\\right\\rangle"
	} else if star1 {
		macro = "\\langle{" + args[0] + "}\\vert{" + args[1] + "}\\vert{" + args[2] + "}\\rangle"
	}
	return p.parseExpansion(macro)
}

func physicsMatrixEnvironment(name string) bool {
	name = strings.TrimSuffix(name, "*")
	switch name {
	case "matrix", "smallmatrix", "pmatrix", "bmatrix", "Bmatrix", "vmatrix", "Vmatrix",
		"psmallmatrix", "bsmallmatrix", "Bsmallmatrix", "vsmallmatrix", "Vsmallmatrix":
		return true
	}
	return false
}

func physicsBodyHasMatrixGenerator(source string) bool {
	for _, name := range []string{"imat", "identitymatrix", "xmat", "xmatrix", "zmat", "zeromatrix", "pmat", "paulimatrix", "dmat", "diagonalmatrix", "admat", "antidiagonalmatrix"} {
		if strings.Contains(source, "\\"+name) {
			return true
		}
	}
	return false
}

func (p *parser) physicsRewriteMatrixBody(body string) (string, error) {
	var rewritten strings.Builder
	start := 0
	for index := 0; index < len(body); {
		if body[index] != '\\' {
			index++
			continue
		}
		scanner := &parser{source: body, pos: index + 1, state: p.state, display: p.display}
		name := scanner.readControlSequence()
		if !physicsMatrixGenerator(name) {
			index = scanner.pos
			continue
		}
		expansion, err := scanner.physicsMatrixExpansion(name)
		if err != nil {
			return "", err
		}
		rewritten.WriteString(body[start:index])
		rewritten.WriteString(expansion)
		start = scanner.pos
		index = scanner.pos
	}
	rewritten.WriteString(body[start:])
	return rewritten.String(), nil
}

func physicsMatrixGenerator(name string) bool {
	switch name {
	case "identitymatrix", "imat", "xmatrix", "xmat", "zeromatrix", "zmat", "paulimatrix", "pmat", "diagonalmatrix", "dmat", "antidiagonalmatrix", "admat":
		return true
	}
	return false
}

func (p *parser) physicsMatrixExpansion(name string) (string, error) {
	switch name {
	case "identitymatrix", "imat":
		raw, _, err := p.readArgument(name, false)
		if err != nil {
			return "", err
		}
		size, ok := physicsParseInt(raw)
		if !ok {
			return "", texError("InvalidNumber", "Invalid number")
		}
		if size <= 1 {
			return "1", nil
		}
		rows := make([]string, size)
		for i := 0; i < size; i++ {
			cells := make([]string, size)
			for j := range cells {
				cells[j] = "0"
			}
			cells[i] = "1"
			rows[i] = strings.Join(cells, " & ")
		}
		return strings.Join(rows, "\\\\ "), nil

	case "xmatrix", "xmat", "zeromatrix", "zmat":
		star := p.readStar()
		value := "0"
		var err error
		if name == "xmatrix" || name == "xmat" {
			value, _, err = p.readArgument(name, false)
			if err != nil {
				return "", err
			}
		}
		rawRows, _, err := p.readArgument(name, false)
		if err != nil {
			return "", err
		}
		rawColumns, _, err := p.readArgument(name, false)
		if err != nil {
			return "", err
		}
		rows, rowOK := physicsParseInt(rawRows)
		columns, columnOK := physicsParseInt(rawColumns)
		if !rowOK || !columnOK || strconv.Itoa(rows) != rawRows || strconv.Itoa(columns) != rawColumns {
			return "", texError("InvalidNumber", "Invalid number")
		}
		if rows < 1 {
			rows = 1
		}
		if columns < 1 {
			columns = 1
		}
		matrixRows := make([]string, rows)
		for i := 1; i <= rows; i++ {
			cells := make([]string, columns)
			for j := 1; j <= columns; j++ {
				cell := value
				if star && (rows > 1 || columns > 1) {
					if rows == 1 {
						cell += "_{" + strconv.Itoa(j) + "}"
					} else if columns == 1 {
						cell += "_{" + strconv.Itoa(i) + "}"
					} else {
						cell += "_{{" + strconv.Itoa(i) + "}{" + strconv.Itoa(j) + "}}"
					}
				}
				cells[j-1] = cell
			}
			matrixRows[i-1] = strings.Join(cells, " & ")
		}
		return strings.Join(matrixRows, "\\\\ "), nil

	case "paulimatrix", "pmat":
		raw, _, err := p.readArgument(name, false)
		if err != nil {
			return "", err
		}
		if raw == "" {
			return "", nil
		}
		matrix := raw[1:]
		switch raw[0] {
		case '0':
			matrix += " 1 & 0\\\\ 0 & 1"
		case '1', 'x':
			matrix += " 0 & 1\\\\ 1 & 0"
		case '2', 'y':
			matrix += " 0 & -i\\\\ i & 0"
		case '3', 'z':
			matrix += " 1 & 0\\\\ 0 & -1"
		}
		return matrix, nil

	case "diagonalmatrix", "dmat", "antidiagonalmatrix", "admat":
		p.skipSpaces()
		if p.pos >= len(p.source) || p.source[p.pos] != '{' {
			return "", nil
		}
		raw, _, err := p.readArgument(name, false)
		if err != nil {
			return "", err
		}
		elements := splitTopLevel(raw, ',')
		anti := name == "antidiagonalmatrix" || name == "admat"
		rows := make([]string, len(elements))
		for i, element := range elements {
			ampersands := i
			if anti {
				ampersands = len(elements) - i - 1
			}
			rows[i] = strings.Repeat("&", ampersands) + "\\mqty{" + element + "}"
		}
		return strings.Join(rows, "\\\\ "), nil
	}
	return "", nil
}

func physicsParseInt(raw string) (int, bool) {
	raw = strings.TrimLeftFunc(raw, unicode.IsSpace)
	if raw == "" {
		return 0, false
	}
	end := 0
	if raw[0] == '+' || raw[0] == '-' {
		end++
	}
	startDigits := end
	for end < len(raw) && raw[end] >= '0' && raw[end] <= '9' {
		end++
	}
	if end == startDigits {
		return 0, false
	}
	value, err := strconv.Atoi(raw[:end])
	return value, err == nil
}
