// Copyright (c) 2021-2022 MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.
//
// Source: ts/input/tex/empheq/EmpheqConfiguration.ts and EmpheqUtil.ts.

package tex

import (
	"strings"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// empheqMO contains the fixed argument passed to EmpheqMethods.EmpheqMO by
// the empheq-macros CommandMap.  The "big" spellings deliberately create the
// same <mo> as their non-big counterparts, exactly as in the upstream map.
var empheqMO = map[string]string{
	"empheqlbrace": "{", "empheqrbrace": "}",
	"empheqlbrack": "[", "empheqrbrack": "]",
	"empheqlangle": "⟨", "empheqrangle": "⟩",
	"empheqlparen": "(", "empheqrparen": ")",
	"empheqlvert": "|", "empheqrvert": "|",
	"empheqlVert": "‖", "empheqrVert": "‖",
	"empheqlfloor": "⌊", "empheqrfloor": "⌋",
	"empheqlceil": "⌈", "empheqrceil": "⌉",

	"empheqbiglbrace": "{", "empheqbigrbrace": "}",
	"empheqbiglbrack": "[", "empheqbigrbrack": "]",
	"empheqbiglangle": "⟨", "empheqbigrangle": "⟩",
	"empheqbiglparen": "(", "empheqbigrparen": ")",
	"empheqbiglvert": "|", "empheqbigrvert": "|",
	"empheqbiglVert": "‖", "empheqbigrVert": "‖",
	"empheqbiglfloor": "⌊", "empheqbigrfloor": "⌋",
	"empheqbiglceil": "⌈", "empheqbigrceil": "⌉",
}

var empheqDelimiters = map[string]bool{
	"empheql": true, "empheqr": true,
	"empheqbigl": true, "empheqbigr": true,
}

// empheqCommand is the single command-dispatch hook for Empheq's delimiter
// helpers.  It must run before the generic source-symbol fallback.
func (p *parser) empheqCommand(name string) (nodes []*mml.Node, handled bool, err error) {
	if character, ok := empheqMO[name]; ok {
		return []*mml.Node{token("mo", character)}, true, nil
	}
	if !empheqDelimiters[name] {
		return nil, false, nil
	}
	delimiter, err := p.readDelimiter(false)
	if err != nil {
		// TexParser.GetDelimiter reports the calling control sequence, rather
		// than the invalid delimiter token, in this diagnostic.
		if texErr, ok := err.(*Error); ok && texErr.ID == "MissingOrUnrecognizedDelim" {
			err = texError("MissingOrUnrecognizedDelim", "Missing or unrecognized delimiter for \\%s", name)
		}
		return nil, true, err
	}
	mo := token("mo", delimiter)
	mo.Attributes.Set("stretchy", true)
	mo.Attributes.Set("symmetric", true)
	return []*mml.Node{mo}, true, nil
}

// empheqEnvironment is the single begin-environment hook.  p.pos must be
// immediately after \begin{empheq}; it consumes options, the inner environment
// argument, and the matching \end{empheq}.
func (p *parser) empheqEnvironment(name string) (nodes []*mml.Node, handled bool, err error) {
	if name != "empheq" {
		return nil, false, nil
	}

	options, _, err := p.readBrackets(nil)
	if err != nil {
		return nil, true, err
	}
	innerArgument, _, err := p.readArgument("begin{"+name+"}", false)
	if err != nil {
		return nil, true, err
	}
	parts := strings.Split(innerArgument, "=")
	innerName := parts[0]
	innerArgumentN := ""
	if len(parts) > 1 {
		innerArgumentN = parts[1]
	}
	if !empheqAllowedEnvironment(innerName) {
		return nil, true, texError("UnknownEnv", "Unknown environment %q", innerName)
	}

	settings := map[string]any{}
	if options != "" {
		settings, err = empheqSplitOptions(options)
		if err != nil {
			return nil, true, err
		}
	}
	body, err := p.captureEnvironment(name)
	if err != nil {
		return nil, true, err
	}

	// The local parser has the same table structure for flalign as align but
	// does not expose flalign as a separate public environment.  Parsing it as
	// align retains Empheq's observable MML construction (including rows and
	// labels) while the Empheq adjustment below supplies its added columns.
	parseName := innerName
	if strings.TrimSuffix(innerName, "*") == "flalign" {
		parseName = "align"
		if strings.HasSuffix(innerName, "*") {
			parseName += "*"
		}
	}
	innerSource := "\\begin{" + parseName + "}"
	if innerArgumentN != "" {
		innerSource += "{" + innerArgumentN + "}"
	}
	innerSource += body + "\\end{" + parseName + "}"
	result, err := p.parseString(innerSource)
	if err != nil {
		return nil, true, err
	}

	left, hasLeft := settings["left"]
	right, hasRight := settings["right"]
	if (hasLeft && empheqTruthy(left)) || (hasRight && empheqTruthy(right)) {
		original := result.Clone()
		if hasLeft && empheqTruthy(left) {
			if err := p.empheqAddLeft(result, original, empheqOptionTeX(left)); err != nil {
				return nil, true, err
			}
		}
		if hasRight && empheqTruthy(right) {
			if err := p.empheqAddRight(result, original, empheqOptionTeX(right)); err != nil {
				return nil, true, err
			}
		}
	}
	return []*mml.Node{result}, true, nil
}

func empheqAllowedEnvironment(name string) bool {
	switch strings.TrimSuffix(name, "*") {
	case "equation", "align", "gather", "flalign", "alignat", "multline":
		return true
	}
	return false
}

func empheqTruthy(value any) bool {
	switch value := value.(type) {
	case bool:
		return value
	case string:
		return value != ""
	}
	return value != nil
}

func empheqOptionTeX(value any) string {
	if value, ok := value.(string); ok {
		return value
	}
	// readKeyval's bare-key and empty-braces forms produce boolean true.
	// Passing it to TexParser in JavaScript yields an empty parse (the boolean
	// has no string length), leaving only the sizing phantom in the cell.
	return ""
}

// empheqSplitOptions is ParseUtil.keyvalOptions(text, {left: 1, right: 1},
// true).  readKeyval trims outer spaces, strips balanced outer braces,
// converts the two boolean literals, and treats a bare key as true.
func empheqSplitOptions(text string) (map[string]any, error) {
	result := make(map[string]any)
	var order []string
	rest := text
	for rest != "" {
		key, end, next, err := empheqReadOptionValue(rest, "=,")
		if err != nil {
			return nil, err
		}
		rest = next
		var value any = true
		if end == '=' {
			var raw string
			raw, _, rest, err = empheqReadOptionValue(rest, ",")
			if err != nil {
				return nil, err
			}
			switch raw {
			case "true":
				value = true
			case "false":
				value = false
			default:
				value = raw
			}
		} else if key == "" {
			continue
		}
		if _, exists := result[key]; !exists {
			order = append(order, key)
		}
		result[key] = value
	}
	// ParseUtil.keyvalOptions validates only after readKeyval has parsed the
	// complete list, so a malformed later value takes precedence over an
	// earlier unknown key.
	for _, key := range order {
		if key != "left" && key != "right" {
			return nil, texError("InvalidOption", "Invalid option: %s", key)
		}
	}
	return result, nil
}

func empheqReadOptionValue(text, endings string) (value string, ending byte, rest string, err error) {
	braces := 0
	start := 0
	startCount := true
	stopCount := false
	var b strings.Builder
	for index := 0; index < len(text); index++ {
		character := text[index]
		switch character {
		case ' ':
			// Spaces do not affect brace-state tracking.  They remain in the
			// collected value and are removed only at its outer edges below.
		case '{':
			if startCount {
				start++
			} else {
				stopCount = false
				if start > braces {
					start = braces
				}
			}
			braces++
		case '}':
			if braces > 0 {
				braces--
			}
			if startCount || stopCount {
				start--
				stopCount = true
			}
			startCount = false
		default:
			if braces == 0 && strings.IndexByte(endings, character) >= 0 {
				cleaned := empheqRemoveOptionBraces(b.String(), start)
				if stopCount {
					cleaned = "true"
				}
				return cleaned, character, text[index+1:], nil
			}
			startCount = false
			stopCount = false
		}
		b.WriteByte(character)
	}
	if braces != 0 {
		return "", 0, "", texError("ExtraOpenMissingClose", "Extra open brace or missing close brace")
	}
	cleaned := empheqRemoveOptionBraces(b.String(), start)
	if stopCount {
		cleaned = "true"
	}
	return cleaned, 0, "", nil
}

func empheqRemoveOptionBraces(text string, count int) string {
	for count > 0 {
		text = strings.TrimSpace(text)
		if len(text) >= 2 {
			text = text[1 : len(text)-1]
		}
		count--
	}
	return strings.TrimSpace(text)
}

func empheqColumnCount(table *mml.Node) int {
	count := 0
	for _, row := range table.Children {
		columns := len(row.Children)
		if row.Kind == "mlabeledtr" {
			columns--
		}
		if columns > count {
			count = columns
		}
	}
	return count
}

func (p *parser) empheqCellBlock(tex string, table *mml.Node) (*mml.Node, error) {
	block := node("mpadded")
	block.Attributes.Set("height", 0)
	block.Attributes.Set("depth", 0)
	block.Attributes.Set("voffset", "-1height")
	contents, err := p.parseString(tex)
	if err != nil {
		return nil, err
	}
	children := []*mml.Node{contents}
	if contents.Kind == "mrow" && contents.Flags.Inferred {
		children = contents.Children
	}
	for _, child := range children {
		block.Children[0].AppendChild(child)
	}
	tableSize := node("mpadded", table)
	tableSize.Attributes.Set("width", 0)
	block.Children[0].AppendChild(node("mphantom", tableSize))
	refreshDynamicFlags(block.Children[0])
	refreshDynamicFlags(block)
	return block, nil
}

func empheqTopRowTable(original *mml.Node) *mml.Node {
	table := original.Clone()
	if len(table.Children) > 1 {
		table.SetChildren(table.Children[:1])
	}
	table.Attributes.Set("align", "baseline 1")
	padded := node("mpadded", table)
	padded.Attributes.Set("width", 0)
	return node("mphantom", padded)
}

func (p *parser) empheqRowspanCell(cell *mml.Node, tex string, table *mml.Node) error {
	block, err := p.empheqCellBlock(tex, table.Clone())
	if err != nil {
		return err
	}
	content := node("mpadded", block, empheqTopRowTable(table))
	content.Attributes.Set("height", 0)
	content.Attributes.Set("depth", 0)
	content.Attributes.Set("voffset", "height")
	cell.Children[0].AppendChild(content)
	refreshDynamicFlags(cell.Children[0])
	refreshDynamicFlags(cell)
	return nil
}

func (p *parser) empheqAddLeft(table, original *mml.Node, left string) error {
	columnAlign, _ := table.Attributes.Get("columnalign")
	columnSpacing, _ := table.Attributes.Get("columnspacing")
	table.Attributes.Set("columnalign", "right "+sourceValueString(columnAlign))
	table.Attributes.Set("columnspacing", "0em "+sourceValueString(columnSpacing))
	var topCell *mml.Node
	for rowIndex := len(table.Children) - 1; rowIndex >= 0; rowIndex-- {
		row := table.Children[rowIndex]
		cell := node("mtd")
		children := append([]*mml.Node{cell}, row.Children...)
		if row.Kind == "mlabeledtr" && len(children) > 1 {
			children[0], children[1] = children[1], children[0]
		}
		row.SetChildren(children)
		topCell = cell
	}
	if topCell == nil {
		return nil
	}
	return p.empheqRowspanCell(topCell, left, original)
}

func (p *parser) empheqAddRight(table, original *mml.Node, right string) error {
	if len(table.Children) == 0 {
		table.AppendChild(node("mtr"))
	}
	columns := empheqColumnCount(table)
	first := table.Children[0]
	for len(first.Children) < columns {
		first.AppendChild(node("mtd"))
	}
	cell := first.AppendChild(node("mtd"))
	if err := p.empheqRowspanCell(cell, right, original); err != nil {
		return err
	}
	columnAlign, _ := table.Attributes.Get("columnalign")
	aligns := strings.Split(sourceValueString(columnAlign), " ")
	if columns < len(aligns) {
		aligns = aligns[:columns]
	}
	table.Attributes.Set("columnalign", strings.Join(aligns, " ")+" left")
	columnSpacing, _ := table.Attributes.Get("columnspacing")
	spacings := strings.Split(sourceValueString(columnSpacing), " ")
	limit := columns - 1
	if limit < 0 {
		limit = 0
	}
	if limit < len(spacings) {
		spacings = spacings[:limit]
	}
	table.Attributes.Set("columnspacing", strings.Join(spacings, " ")+" 0em")
	return nil
}
