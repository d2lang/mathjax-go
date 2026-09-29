// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Sources: PhysicsMappings, PhysicsMethods.Quantity and PhysicsItems.AutoOpen.
package tex

import "github.com/d2lang/mathjax-go/internal/mml"

type physicsQuantitySpec struct {
	open, close    string
	argument       bool
	named, variant string
}

var physicsQuantities = func() map[string]physicsQuantitySpec {
	result := make(map[string]physicsQuantitySpec)
	for _, table := range mjSourcePhysicsMaps {
		if table.Name != "Physics-automatic-bracing-macros" {
			continue
		}
		for _, entry := range table.Entries {
			q := physicsQuantitySpec{open: "(", close: ")"}
			if entry.Value == "Quantity" {
				result[entry.Name] = q
				continue
			}
			args, ok := entry.Value.(mjSourceList)
			if !ok || len(args) < 4 || args[0] != "Quantity" {
				continue
			}
			q.open, q.close, q.argument = args[1].(string), args[2].(string), args[3].(bool)
			if len(args) > 4 {
				q.named, q.variant = args[4].(string), args[5].(string)
			}
			result[entry.Name] = q
		}
	}
	return result
}()

// The named Quantity pushes its FnItem before GetArgument and the fresh child
// parse. Its recipient may reject that item before either operation executes.
type physicsQuantityArgument struct {
	name, open, close, big string
	star                   bool
}

func (a *physicsQuantityArgument) parse(p *parser) ([]*mml.Node, error) {
	argument, _, err := p.readArgumentAtCursor(a.name, false)
	if err != nil {
		return nil, err
	}
	source := a.open + " " + argument + " " + a.close
	if !a.star {
		if a.big != "" {
			source = "\\" + a.big + "l" + a.open + " " + argument + " \\" + a.big + "r" + a.close
		} else {
			source = "\\left" + a.open + " " + argument + " \\right" + a.close
		}
	}
	result, err := p.parseChild(source)
	if err != nil {
		return nil, err
	}
	return unwrapInferred(result), nil
}

func (p *parser) physicsQuantity(name string, q physicsQuantitySpec, after **derivativeAutoOpen) ([]*mml.Node, error) {
	star := q.argument && p.readStarSkipping(internalTextSpace)
	getNext := func() rune {
		for p.pos < len(p.source) && internalTextSpace(p.peekRune()) {
			p.consumeRune()
		}
		return p.peekRune()
	}
	next := getNext()
	position := p.pos
	empty := func() ([]*mml.Node, error) {
		p.pos = position
		return []*mml.Node{p.leftRightFenced(q.open, forcedRow(nil, false), q.close, true)}, nil
	}
	big := ""
	if next == '\\' {
		p.pos++
		big = p.readControlSequence()
		switch big {
		case "big", "Big", "bigg", "Bigg":
			next = getNext()
		default:
			return empty()
		}
	}
	if q.argument && next != '{' {
		return nil, texError("MissingArgFor", "Missing argument for \\%s", name)
	}
	close, ok := map[rune]byte{'(': ')', '[': ']', '{': '}', '|': '|'}[next]
	if !ok {
		return empty()
	}
	if next == '{' {
		open, right := q.open, q.close
		if !q.argument {
			open, right = "\\{", "\\}"
		}
		argument := &physicsQuantityArgument{name: name, open: open, close: right, big: big, star: star}
		if q.named != "" {
			function := p.token("mi", q.named)
			function.TeXClass = mml.TeXClassOp
			function.SetProperty("texClass", mml.TeXClassOp)
			// NodeFactory token overrides run before Quantity sets its variant.
			p.applyVectorFactory(function)
			function.Attributes.Set("mathvariant", q.variant)
			p.commandNamedFunction = true
			*after = &derivativeAutoOpen{quantity: argument}
			return []*mml.Node{function}, nil
		}
		return argument.parse(p)
	}
	// Quantity consumes the physical opening and then pushes AutoOpen. The
	// content remains on the caller's input, font scope and macro budget.
	p.pos++
	*after = &derivativeAutoOpen{open: byte(next), closer: close, openingConsumed: true, big: big}
	return nil, nil
}
