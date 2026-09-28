// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Sources: BaseMethods.HLine and BaseItems.ArrayItem.checkLines/createMml.
package tex

import (
	"strings"

	"github.com/d2lang/mathjax-go/internal/mml"
)

const arrayRulesProperty = "arrayRules"

// Each real ArrayItem owns its rules. Reading the owner's completed-row count
// also observes direct EndRow calls, such as ArrowBetweenLines, without
// confusing them with an Entry or an ordinary no-item command.
type arrayRules struct {
	rows   func() int
	frame  []string
	lines  []string
	dashed bool
}

func newArrayRules(table *mml.Node) *arrayRules {
	rules := &arrayRules{rows: func() int { return len(table.Children) }}
	table.SetProperty(arrayRulesProperty, rules)
	return rules
}

func isHLine(name string) bool { return name == "hline" || name == "hdashline" }

func misplacedHLine(name string) error {
	return texError("Misplaced", "Misplaced %s", "\\"+name)
}

func (rules *arrayRules) add(name string) {
	rows := rules.rows()
	if rows == 0 {
		rules.frame = append(rules.frame, "top")
		return
	}
	for len(rules.lines) < rows {
		rules.lines = append(rules.lines, "none")
	}
	style := "solid"
	if name == "hdashline" {
		style = "dashed"
	}
	rules.lines[rows-1] = style
}

func tableArrayRules(table *mml.Node) *arrayRules {
	if value, ok := table.Property(arrayRulesProperty); ok {
		if rules, ok := value.(*arrayRules); ok {
			return rules
		}
	}
	return newArrayRules(table)
}

// Call after the owner has finished its entries/rows and applied its arraydef,
// but before adding the owner's delimiters or delivering its final MML item.
func finishArrayRules(table *mml.Node) *mml.Node {
	value, ok := table.Property(arrayRulesProperty)
	if !ok {
		return table
	}
	rules := value.(*arrayRules)
	table.RemoveProperty(arrayRulesProperty)
	if len(rules.lines) != 0 {
		if len(rules.lines) == len(table.Children) {
			rules.frame = append(rules.frame, "bottom")
			rules.lines = rules.lines[:len(rules.lines)-1]
		} else if len(rules.lines) < len(table.Children)-1 {
			rules.lines = append(rules.lines, "none")
		}
		table.Attributes.Set("rowlines", strings.Join(rules.lines, " "))
	}
	if len(rules.frame) == 4 {
		style := "solid"
		if rules.dashed {
			style = "dashed"
		}
		table.Attributes.Set("frame", style)
		return table
	}
	if len(rules.frame) == 0 {
		return table
	}
	// The original constructs mtable before trimming the arraydef's trailing
	// repeated none entries. The trimmed spelling controls partial-frame padding;
	// it does not rewrite the already-created table's rowlines attribute.
	rowLines := append([]string(nil), rules.lines...)
	for len(rowLines) > 1 && rowLines[len(rowLines)-1] == "none" && rowLines[len(rowLines)-2] == "none" {
		rowLines = rowLines[:len(rowLines)-1]
	}
	columnLines, _ := table.Attributes.Get("columnlines")
	columnText, rowText := sourceValueString(columnLines), strings.Join(rowLines, " ")
	table.Attributes.Set("frame", "")
	enclosure := setAttributes(node("menclose", table), map[string]any{"notation": strings.Join(rules.frame, " ")})
	if (columnText != "" && columnText != "none") || (rowText != "" && rowText != "none") {
		enclosure.Attributes.Set("data-padding", 0)
	}
	return enclosure
}
