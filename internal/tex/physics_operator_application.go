// Copyright (c) 2018-2022 MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Source: physics/PhysicsMethods.ts vectorApplication and OperatorApplication.

package tex

import (
	"strings"

	"github.com/d2lang/mathjax-go/internal/mml"
)

type physicsOperatorApplication struct {
	operator string
	fences   string
}

type physicsApplicationArgument struct {
	name, fences string
}

// Only expression-map OperatorApplication registrations belong here. Vector
// operators and NamedFn/Expression use their own source methods and hooks.
var physicsExpressionOperators = func() map[string]physicsOperatorApplication {
	result := make(map[string]physicsOperatorApplication)
	for _, table := range mjSourcePhysicsMaps {
		if table.Name != "Physics-expressions-macros" {
			continue
		}
		for _, entry := range table.Entries {
			args, ok := entry.Value.(mjSourceList)
			if !ok || len(args) < 2 || args[0] != "OperatorApplication" {
				continue
			}
			application := physicsOperatorApplication{operator: args[1].(string)}
			for _, fence := range args[2:] {
				application.fences += fence.(string)
			}
			result[entry.Name] = application
		}
	}
	return result
}()

func (p *parser) physicsOperatorApplication(name string, application physicsOperatorApplication, after **derivativeAutoOpen) ([]*mml.Node, error) {
	// The operator is a fresh TexParser result, not a fixed token. Its inner
	// commands may have been redefined, and errors occur before GetArgument.
	op, err := p.parseChild(application.operator)
	if err != nil {
		return nil, err
	}
	p.commandNamedFunction = true
	// Push the FnItem before inspecting the caller's argument. A recipient
	// such as SubsupItem can reject it before a malformed brace is scanned.
	*after = &derivativeAutoOpen{application: &physicsApplicationArgument{name, application.fences}}
	return []*mml.Node{op}, nil
}

func (a *physicsApplicationArgument) prepare(p *parser, after *derivativeAutoOpen) (bool, error) {
	for p.pos < len(p.source) && isPrimeSpace(p.peekRune()) {
		p.consumeRune()
	}
	if p.pos == len(p.source) {
		return false, nil
	}
	left := p.source[p.pos]
	enlarge := strings.ContainsRune(a.fences, rune(left))
	if left == '{' {
		arg, _, err := p.readArgument(a.name, false)
		if err != nil {
			return false, err
		}
		lfence, rfence := "", ""
		if enlarge {
			lfence, rfence = "\\left\\{", "\\right\\}"
		}
		// The braces are consumed, not retained as a lexical group. Parse the
		// replacement and caller suffix together so fonts and macros continue.
		p.source, p.pos = lfence+" "+arg+" "+rfence+p.source[p.pos:], 0
	} else if enlarge {
		closer := byte(')')
		if left == '[' {
			closer = ']'
		}
		after.open, after.closer = left, closer
		return true, nil
	}
	return false, nil
}
