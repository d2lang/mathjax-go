// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Sources: BaseMethods.Array/AlignedArray, BaseItems.ArrayItem/BeginItem,
// MathtoolsMethods.MtMatrix/MtSmallMatrix/SpreadLines, and Stack.Pop.
package tex

import (
	"strings"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// Array and its enclosing Begin are separate stack items. EndTable removes
// Array and replays the same End to Begin; a direct Pop removes only Array
// and leaves Begin executing with the pending cell's raw nodes.
type ordinaryArrayItem struct {
	begin  *environmentFrame
	end    *environmentEndItem
	popped bool
	nodes  []*mml.Node
}

// These are precisely the pinned Array/AlignedArray registrations and the
// MtMatrix/MtSmallMatrix wrappers that delegate to Array. Other table owners
// retain their own scanners, EndTable callbacks, and closing-item contracts.
func isOrdinaryArrayEnvironment(name string) bool {
	switch name {
	case "array", "matrix", "pmatrix", "bmatrix", "Bmatrix", "vmatrix", "Vmatrix",
		"cases", "subarray", "crampedsubarray", "dcases", "rcases", "drcases",
		"smallmatrix", "matrix*", "pmatrix*", "bmatrix*", "Bmatrix*", "vmatrix*", "Vmatrix*",
		"smallmatrix*", "psmallmatrix", "psmallmatrix*", "bsmallmatrix", "bsmallmatrix*",
		"Bsmallmatrix", "Bsmallmatrix*", "vsmallmatrix", "vsmallmatrix*", "Vsmallmatrix", "Vsmallmatrix*":
		return true
	}
	return false
}

// Begin copies the caller's lexical environment before Array clears its own.
// Keep the full environment for resuming Begin after Array is directly popped.
type ordinaryBeginEnvironment struct {
	multiLetterFont, activeFont, activeColor string
	identifierPattern                       identifierPattern
	operatorLetters, noAutoOP                bool
	fontExplicitEmpty, inRoot                bool
	vectorFont                              string
	vectorStar                              bool
}

func (p *parser) ordinaryBeginEnvironment() ordinaryBeginEnvironment {
	return ordinaryBeginEnvironment{
		multiLetterFont: p.multiLetterFont, activeFont: p.activeFont, activeColor: p.activeColor,
		identifierPattern: p.identifierPattern, operatorLetters: p.operatorLetters, noAutoOP: p.noAutoOP,
		fontExplicitEmpty: p.fontExplicitEmpty, inRoot: p.inRoot,
		vectorFont: p.vectorFont, vectorStar: p.vectorStar,
	}
}

func (env ordinaryBeginEnvironment) restore(p *parser) {
	p.multiLetterFont, p.activeFont, p.activeColor = env.multiLetterFont, env.activeFont, env.activeColor
	p.identifierPattern, p.operatorLetters, p.noAutoOP = env.identifierPattern, env.operatorLetters, env.noAutoOP
	p.fontExplicitEmpty, p.inRoot = env.fontExplicitEmpty, env.inRoot
	p.vectorFont, p.vectorStar = env.vectorFont, env.vectorStar
}

func (p *parser) ordinaryArrayEnvironment(name string, begin *environmentFrame) ([]*mml.Node, error) {
	lexical := p.ordinaryBeginEnvironment()
	defer lexical.restore(p)
	// Retain the existing Array setup scope; individual cells additionally
	// clear font and identifier state in matrixCellParser.
	p.inRoot, p.activeColor = false, ""
	// Ordinary Array owns its EOF check independently of legacy streaming
	// frames, which can be inherited by captured table child parsers.
	begin.ordinaryArray = true
	owner := &ordinaryArrayItem{begin: begin}

	columns := "c"
	var err error
	if name == "array" || name == "subarray" || name == "crampedsubarray" {
		// AlignedArray's optional-position setup is a separate existing gap.
		// This ownership change retains the existing argument reader/order.
		columns, _, err = p.readArgument("begin{"+name+"}", false)
		if err != nil {
			return nil, err
		}
	}
	small := strings.Contains(name, "smallmatrix")
	if strings.HasSuffix(name, "*") {
		alignment := "c"
		if small {
			alignment = p.mathtoolsOption("smallmatrix-align")
		}
		columns, _, err = p.readBrackets(&alignment)
		if err != nil {
			return nil, err
		}
	}
	if small || strings.HasSuffix(name, "*") {
		columns, err = p.completeArrayAlignment(columns)
		if err != nil {
			return nil, err
		}
	}
	style, spacing := "T", "4pt"
	subarray := name == "subarray" || name == "crampedsubarray"
	cases := name == "cases" || name == "dcases" || name == "rcases" || name == "drcases"
	if small || subarray {
		style = "S"
	}
	if small || cases {
		spacing = ".2em"
	}
	if subarray {
		spacing = "0.1em"
	}
	if cases && strings.HasPrefix(name, "d") {
		style = "D"
	}
	table, err := p.parseTableWithOrdinaryArray("", style, true, owner, spacing)
	if err != nil {
		return nil, err
	}
	if owner.popped {
		lexical.restore(p)
		return p.continueOrdinaryBegin(begin, owner.nodes)
	}

	// Only a real EndTable result receives table metadata, rules and fences.
	// Keep each registration's existing setup while replacing raw capture.
	if small {
		resetTableAttributes(table,
			"columnalign", "center",
			"columnspacing", ".333em",
			"rowspacing", tableRowSpacing(table),
			"displaystyle", false,
		)
		table.SetProperty("useHeight", false)
		table.SetProperty("scriptlevel", 1)
	} else if subarray {
		resetTableAttributes(table,
			"columnspacing", "0em",
			"rowspacing", tableRowSpacing(table),
		)
		table.SetProperty("useHeight", false)
		table.SetProperty("scriptlevel", 1)
		table.Attributes.Set("displaystyle", false)
		if name == "crampedsubarray" {
			table.Attributes.Set("data-cramped", true)
		}
	} else if strings.Contains(name, "matrix") {
		table.Attributes.Set("columnalign", "center")
	}
	if cases {
		resetTableAttributes(table,
			"columnalign", "left left",
			"columnspacing", "1em",
			"rowspacing", tableRowSpacing(table),
		)
		table.Attributes.Set("displaystyle", false)
		open, close := "{", ""
		if strings.Contains(name, "rcases") {
			open, close = "", "}"
		}
		table = p.leftRightFenced(open, finishArrayRules(table), close, true)
	} else {
		table = applyColumnSpec(table, columns)
		open, close := matrixDelimiters(name)
		if open != "" || close != "" {
			table = p.leftRightFenced(open, table, close, true)
		}
		if small {
			table = finishArrayRules(table)
		}
	}
	if owner.end == nil {
		// Stop first reduces pending items, then Array, and finally Begin.
		return nil, begin.missing()
	}
	if err := p.closeEnvironment(begin, owner.end); err != nil {
		return nil, err
	}
	return []*mml.Node{table}, nil
}

func (p *parser) continueOrdinaryBegin(begin *environmentFrame, prefix []*mml.Node) ([]*mml.Node, error) {
	previous, array := p.environmentRow, p.ordinaryArray
	cell, matrixClose, entry, braket := p.arrayCell, p.matrixClose, p.cdArrayEntry, p.braketOwner
	p.environmentRow, p.ordinaryArray = begin, nil
	p.arrayCell, p.matrixClose, p.cdArrayEntry, p.braketOwner = nil, false, false, nil
	defer func() {
		p.environmentRow, p.ordinaryArray = previous, array
		p.arrayCell, p.matrixClose, p.cdArrayEntry, p.braketOwner = cell, matrixClose, entry, braket
	}()
	children, _, err := p.parseRowContinuation(0, false, false, nil, "", prefix)
	if err != nil {
		return nil, err
	}
	if !begin.closed {
		return nil, begin.missing()
	}
	return children, nil
}
