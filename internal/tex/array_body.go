// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Source: BaseItems.ArrayItem.checkItem/EndEntry/EndRow/EndTable.
package tex

import (
	"math"
	"strings"

	"github.com/d2lang/mathjax-go/internal/jscompat"
	"github.com/d2lang/mathjax-go/internal/mml"
)

// Each array kind keeps its own callbacks. Entry and CR are emitted by the
// parser, after command argument readers and dynamic maps have had ownership.
// Fresh cell parsers implement ArrayItem.clearEnv without clearing Stack.global.
type arrayBodyOwner struct {
	live         bool
	environment  *environmentFrame
	ordinary     *ordinaryArrayItem
	requireClose bool
	rules        *arrayRules
	configure    func(*parser)
	prepare      func(*parser) ([]*mml.Node, error)
	prepareEntry func(*parser, *cellItem) ([]*mml.Node, error)
	hasEntries   func() bool
	endEntry     func([]*mml.Node, *arrayCellState) error
	endRow       func() error
	addSpacing   func(string)
}

func (p *parser) parseArrayBody(source string, owner arrayBodyOwner) error {
	terminator := byte(0)
	if owner.requireClose {
		terminator = '}'
		if !owner.live {
			source += "}"
		}
	}
	position := 0
	if owner.live || owner.environment != nil {
		source, position = p.source, p.pos
	}
	var current *parser
	defer func() {
		if (owner.live || owner.environment != nil) && current != nil {
			p.source, p.pos = current.source, current.pos
		}
	}()
	var previous *cellItem
	for {
		sub := p.matrixCellParser(source)
		current = sub
		sub.pos = position
		if owner.environment != nil {
			sub.environmentRow = owner.environment
		} else if owner.live {
			// A standalone Matrix Array is above the enclosing recipient;
			// it does not install an environment Begin of its own.
			sub.environmentRow = nil
		}
		if owner.live {
			sub.liveMatrix = true
		}
		sub.ordinaryArray = owner.ordinary
		sub.matrixClose, sub.cdArrayEntry = owner.requireClose, true
		sub.arrayCell.rules = owner.rules
		if owner.configure != nil {
			owner.configure(sub)
		}
		var prefix []*mml.Node
		var err error
		if owner.prepare != nil {
			prefix, err = owner.prepare(sub)
			if err != nil {
				return err
			}
		}
		if owner.prepareEntry != nil {
			entryPrefix, err := owner.prepareEntry(sub, previous)
			if err != nil {
				return err
			}
			prefix = append(prefix, entryPrefix...)
		}
		children, _, err := sub.parseRowContinuation(terminator, false, false, nil, "", prefix)
		if err != nil {
			return err
		}
		if owner.ordinary != nil && owner.ordinary.popped {
			// Pop.toMml uses only current nodes, not completed row/table
			// buffers. Preserve their creation-time font before resuming
			// the surrounding Begin's original lexical environment.
			matrixCellContent(children)
			owner.ordinary.nodes = children
			return nil
		}
		item := sub.stoppedCell
		// EndTable omits only a pending empty final row. Explicit Entry and
		// CR always deliver their empty entry, and CR always ends its row.
		if item == nil && len(children) == 0 && !owner.hasEntries() {
			return nil
		}
		if err := owner.endEntry(children, sub.arrayCell); err != nil {
			return err
		}
		if item == nil || item.cr {
			if err := owner.endRow(); err != nil {
				return err
			}
			if item != nil && item.spacing != "" && owner.addSpacing != nil {
				owner.addSpacing(item.spacing)
			}
		}
		if item == nil {
			return nil
		}
		// Macro expansion may replace the current source. Do not resume a
		// substring of the original captured program or a pre-split row.
		if owner.live || owner.environment != nil {
			source, position = sub.source, sub.pos
		} else {
			source, position = sub.source[sub.pos:], 0
		}
		previous = item
	}
}

type arrayRowSpacing struct {
	rows        int
	adjustments map[int]string
}

func (spacing *arrayRowSpacing) add(adjust string) {
	if spacing.adjustments == nil {
		spacing.adjustments = make(map[int]string)
	}
	spacing.adjustments[spacing.rows-1] = adjust
}

func (spacing *arrayRowSpacing) apply(table *mml.Node, initial string) {
	if spacing.adjustments == nil {
		table.Attributes.Set("rowspacing", initial)
		return
	}
	base := matrixDimensionEm(initial)
	values := make([]string, spacing.rows)
	for i := range values {
		values[i] = arraySpacingEm(math.Max(0, base+matrixDimensionEm(spacing.adjustments[i])))
	}
	table.Attributes.Set("rowspacing", strings.Join(values, " "))
}

// ArrayItem.addRowSpacing uses ParseUtil.Em, whose zero cutoff and zero
// spelling differ from the output renderer's length formatter.
func arraySpacingEm(value float64) string {
	if math.Abs(value) < .0006 {
		return "0em"
	}
	return strings.TrimSuffix(strings.TrimRight(jscompat.ToFixed(value, 3), "0"), ".") + "em"
}

func tableRowSpacing(table *mml.Node) any {
	spacing, _ := table.Attributes.Get("rowspacing")
	return spacing
}

func (p *parser) parseMultlineBody(source string, ams bool) (*mml.Node, *arrayRowSpacing, error) {
	table := node("mtable")
	spacing := &arrayRowSpacing{}
	var entries []*mml.Node
	err := p.parseArrayBody(source, arrayBodyOwner{
		rules:      newArrayRules(table),
		configure:  func(sub *parser) { sub.arrayCell.multline = ams },
		hasEntries: func() bool { return len(entries) != 0 },
		endEntry: func(children []*mml.Node, state *arrayCellState) error {
			if len(table.Children) != 0 {
				children = fixInitialMO(children)
			}
			cell := node("mtd", matrixCellContent(children))
			if state.shove != "" {
				cell.Attributes.Set("columnalign", state.shove)
			}
			entries = append(entries, cell)
			return nil
		},
		endRow: func() error {
			if len(entries) != 1 {
				return texError("MultlineRowsOneCol", "The rows within the %s environment must have exactly one column", "multline")
			}
			table.AppendChild(node("mtr", entries...))
			entries = nil
			spacing.rows++
			return nil
		},
		addSpacing: spacing.add,
	})
	return table, spacing, err
}
