// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Source: BaseMethods.BeginEnd, BeginItem/EndItem, CasesMethods.NumCases,
// and MathtoolsMethods.SpreadLines.
package tex

import "github.com/d2lang/mathjax-go/internal/mml"

// Environment ownership belongs to one executing input, not shared macro
// configuration. A genuine TexParser child starts without an outer owner.
// Captured legacy owners are retained as barriers, but never charge at EOF.
type environmentFrame struct {
	name           string
	parent         *environmentFrame
	stream, closed bool
	ordinaryArray  bool
	casesBegin     *casesBeginItem
}

// Only the builtin BeginEnd handler can produce this item. Dynamic command
// maps and argument readers run first; neither raw capture nor EOF creates it.
type environmentEndItem struct {
	name   string
	spread bool
}

func (item *environmentEndItem) extra() error {
	return texError("MissingBeginExtraEnd", "Missing \\begin{%s} or extra \\end{%s}", item.name, item.name)
}

func (frame *environmentFrame) missing() error {
	return texError("EnvMissingEnd", "Missing \\end{%s}", frame.name)
}

func (p *parser) endEnvironment(name string) ([]*mml.Node, error) {
	environment, err := p.readEnvironmentName(name)
	if err != nil {
		return nil, err
	}
	registered, item := p.sourceEndDefinition(environment)
	if !item {
		// BeginEnd charges before its environment handler touches the stack.
		if err := p.countEnvironment(); err != nil {
			return nil, err
		}
	}
	if !registered {
		return nil, texError("UnknownEnv", "Unknown environment '%s'", environment)
	}
	if p.environmentOwner == nil && environment == "spreadlines" && !p.liveMatrix {
		// An orphan SpreadLines executes Pop rather than emitting EndItem.
		// Its source runtime failure remains a separate bounded diagnostic.
		return nil, texError("ExtraEnd", "Extra \\end{%s}", environment)
	}
	p.commandEnvironmentEnd = &environmentEndItem{name: environment, spread: environment == "spreadlines"}
	return nil, nil
}

func (p *parser) closeEnvironment(frame *environmentFrame, item *environmentEndItem) error {
	if frame == nil {
		return item.extra()
	}
	if item.name != frame.name {
		return texError("EnvBadEnd", "\\begin{%s} ended with \\end{%s}", frame.name, item.name)
	}
	if cases := frame.casesBegin; cases != nil && cases.end {
		// Cases pushes one Begin object twice. The first End removes
		// its upper entry and flips the shared end property; decoration
		// executes while the remaining entry still owns these nodes.
		cases.end = false
		cases.closing = item
	} else {
		frame.closed = true
	}
	p.pendingEnvironmentEnd = nil
	return nil
}

// SpreadLines is an open BeginItem on this same input. Parse from the current
// cursor until its executing handler pops that item; return the updated input
// and cursor unchanged by any artificial body capture or EOF callback.
func (p *parser) parseEnvironmentContinuation(frame *environmentFrame) (*mml.Node, error) {
	previous, ordinary := p.environmentRow, p.ordinaryArray
	arrayCell, matrixClose, cdArrayEntry, braket := p.arrayCell, p.matrixClose, p.cdArrayEntry, p.braketOwner
	p.environmentRow, p.ordinaryArray = frame, nil
	p.arrayCell, p.matrixClose, p.cdArrayEntry, p.braketOwner = nil, false, false, nil
	defer func() {
		p.environmentRow, p.ordinaryArray = previous, ordinary
		p.arrayCell, p.matrixClose, p.cdArrayEntry, p.braketOwner = arrayCell, matrixClose, cdArrayEntry, braket
	}()
	children, _, err := p.parseRowContinuation(0, false, false, nil, "", nil)
	if err != nil {
		return nil, err
	}
	if !frame.closed {
		return nil, frame.missing()
	}
	return row(children, true), nil
}
