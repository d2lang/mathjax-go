// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Sources: ts/input/tex/ams/{AmsMethods,AmsItems}.ts, FlalignItem and XalignAt.

package tex

import (
	"strconv"
	"strings"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// amsFlalignLayout retains the source distinction between authored entries,
// declared equation pairs, and the fit columns inserted by EndRow.
type amsFlalignLayout struct {
	name       string
	entries    float64
	padded     bool
	center     bool
	zeroLabel  bool
	maxColumns int
}

func newAMSFlalignLayout(name, count string) *amsFlalignLayout {
	layout := &amsFlalignLayout{name: name, center: !isAMSXAlignAt(name), zeroLabel: isAMSXAlignAt(name)}
	if isAMSXAlignAt(name) {
		layout.padded = name != "xxalignat"
		// The scanner has already validated decimal digits. Empty and zero
		// counts disable the constraint; other counts use JavaScript's numeric
		// representation rather than a machine-width integer conversion.
		if count != "" {
			value, _ := strconv.ParseFloat(count, 64)
			layout.entries = 2 * value
		}
	}
	return layout
}

func (layout *amsFlalignLayout) endEntry(entries int) error {
	if layout.entries != 0 && float64(entries) > layout.entries {
		return texError("XalignOverflow", "Extra & in row of %s", layout.name)
	}
	return nil
}

func (layout *amsFlalignLayout) endRow(row *mml.Node) {
	entries := append([]*mml.Node(nil), row.Children...)
	for float64(len(entries)) < layout.entries {
		entries = append(entries, node("mtd"))
	}
	cells := make([]*mml.Node, 0, len(entries)+len(entries)/2+2)
	if layout.padded {
		cells = append(cells, node("mtd"))
	}
	for len(entries) != 0 {
		cells = append(cells, entries[0])
		entries = entries[1:]
		if len(entries) != 0 {
			cells = append(cells, entries[0])
			entries = entries[1:]
		}
		if len(entries) != 0 || layout.padded {
			cells = append(cells, node("mtd"))
		}
	}
	row.SetChildren(cells)
	if len(cells) > layout.maxColumns {
		layout.maxColumns = len(cells)
	}
}

func (layout *amsFlalignLayout) label(tag *mml.Node) *mml.Node {
	if layout.zeroLabel {
		padded := node("mpadded", tag.Children...)
		padded.Attributes.Set("width", 0)
		padded.Attributes.Set("lspace", "-1width")
		tag.SetChildren([]*mml.Node{forcedRow([]*mml.Node{padded}, true)})
	}
	return tag
}

func (layout *amsFlalignLayout) endTable(table *mml.Node) {
	align, width, labelSpacing := "right left center", "auto auto fit", "0.8em"
	if layout.padded {
		align, width = "center right left", "fit auto auto"
	}
	if layout.zeroLabel {
		labelSpacing = "0"
	}
	resetTableAttributes(table,
		"width", "100%",
		"displaystyle", true,
		"columnalign", repeatAMSEqnArrayDefinition(align, layout.maxColumns),
		"columnspacing", "0em",
		"columnwidth", repeatAMSEqnArrayDefinition(width, layout.maxColumns),
		"rowspacing", "3pt",
		"side", "right",
		"minlabelspacing", labelSpacing,
		"data-width-includes-label", true,
	)
	if layout.center && layout.maxColumns <= 2 {
		table.Attributes.Explicit().Delete("width")
	}
}

func (p *parser) amsFlalignEnvironment(name string) ([]*mml.Node, error) {
	count := ""
	if isAMSXAlignAt(name) {
		var err error
		count, err = p.readEquationPairCount(name)
		if err != nil {
			return nil, err
		}
	}
	if err := p.checkEquationEnvironment(); err != nil {
		return nil, err
	}
	layout := newAMSFlalignLayout(name, count)
	tags := p.amsTags()
	numbered := name == "flalign" || name == "xalignat"
	tags.start(name, numbered, numbered)
	defer tags.end()
	body, err := p.captureEnvironment(name)
	if err != nil {
		return nil, err
	}
	rows := splitTable(body)
	table := node("mtable")
	for rowIndex, cells := range rows {
		entries, err := p.parseFlalignEntries(cells, rowIndex == len(rows)-1, layout)
		if err != nil {
			return nil, err
		}
		if entries == nil {
			continue
		}
		row := node("mtr", entries...)
		layout.endRow(row)
		tag, err := tags.getTag(p)
		if err != nil {
			return nil, err
		}
		if tag != nil {
			row = node("mlabeledtr", append([]*mml.Node{layout.label(tag)}, row.Children...)...)
		}
		tags.clearTag()
		table.AppendChild(row)
	}
	layout.endTable(table)
	return []*mml.Node{table}, nil
}

// FlalignItem owns Entry tokens after the TeX parser has had the opportunity
// to consume an ampersand as a command argument. It has no EqnArray kind, so
// the Mathtools commands guarded by checkAlignment remain unavailable here.
func (p *parser) parseFlalignEntries(cells []string, final bool, layout *amsFlalignLayout) ([]*mml.Node, error) {
	source := strings.Join(cells, "&")
	var entries []*mml.Node
	for {
		sub := p.matrixCellParser(source)
		sub.matrixClose = false
		sub.cdArrayEntry = true
		children, _, err := sub.parseRowWithInfix(0, false, false)
		if err != nil {
			return nil, err
		}
		if !sub.cdEntryStopped && final && len(children) == 0 && len(entries) == 0 {
			return nil, nil
		}
		if len(entries) != 0 {
			children = fixInitialMO(children)
		}
		entries = append(entries, node("mtd", matrixCellContent(children)))
		// The declared-count check precedes parsing the following cell.
		if err := layout.endEntry(len(entries)); err != nil {
			return nil, err
		}
		if !sub.cdEntryStopped {
			return entries, nil
		}
		source = sub.source[sub.pos:]
	}
}
