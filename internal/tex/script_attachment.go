// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Based on BaseMethods.Subscript and BaseMethods.Superscript (MathJax 3.2.2).
package tex

import "github.com/d2lang/mathjax-go/internal/mml"

// attachScriptBase preserves authored annotation ownership. The source tests
// family/subtype and occupied slots before choosing whether to reuse a script
// wrapper or wrap the entire base. This parser's own eager compact subtypes
// represent an unfinished generic family; their transient origin bridges that
// representation only. Pending-prime behavior is retained as a separate parser
// boundary, rather than changing PrimeItem handling here.
func attachScriptBase(base, script *mml.Node, marker byte, moves bool) (*mml.Node, error) {
	origin, _ := base.Property(limitsScriptOrigin)
	eager := origin == true || origin == "prime"
	side := base.Kind == "msub" || base.Kind == "msup" || base.Kind == "msubsup"
	limits := base.Kind == "munder" || base.Kind == "mover" || base.Kind == "munderover"
	supOnly := base.Kind == "msup" && !eager
	overOnly := base.Kind == "mover" && !eager
	var under, over *mml.Node
	if side || limits {
		if len(base.Children) > 1 {
			if base.Kind == "msup" || base.Kind == "mover" {
				over = base.Children[1]
			} else {
				under = base.Children[1]
			}
		}
		if len(base.Children) > 2 {
			over = base.Children[2]
		}
	}
	occupied := under
	if marker == '^' {
		occupied = over
	}
	allowed, _ := base.Property("subsupOK")
	if (side && !supOnly && occupied != nil) || (limits && !overOnly && occupied != nil && !limitsTruthy(allowed)) {
		if marker == '_' {
			return nil, texError("DoubleSubscripts", "Double subscripts: use braces to clarify")
		}
		return nil, texError("DoubleExponent", "Double exponent: use braces to clarify")
	}
	reuse := side && !supOnly
	if !reuse && moves {
		reuse = limits && !overOnly && occupied == nil
	}
	if !reuse {
		kind := "msub"
		if marker == '^' {
			kind = "msup"
		}
		if moves {
			kind = "munder"
			if marker == '^' {
				kind = "mover"
			}
		}
		result := node(kind, base, script)
		if moves {
			result.SetProperty("movesupsub", true)
		}
		return result, nil
	}
	if marker == '_' {
		under = script
	} else {
		over = script
	}
	children := []*mml.Node{base.Children[0]}
	kind := "msubsup"
	if limits {
		kind = "munderover"
	}
	if under != nil && over != nil {
		children = append(children, under, over)
	} else if under != nil {
		children = append(children, under)
		kind = map[bool]string{false: "msub", true: "munder"}[limits]
	} else {
		children = append(children, over)
		kind = map[bool]string{false: "msup", true: "mover"}[limits]
	}
	// Keep the accepted pending-prime representation boundary: consuming its
	// subscript creates a new wrapper, leaving the old prime's child slots
	// intact for the existing Limits lifetime/identity contract (D060).
	if origin == "prime" {
		return node(kind, children...), nil
	}
	base.Kind = kind
	base.Flags.Arity = len(children)
	base.SetChildren(children)
	refreshDynamicFlags(base)
	return base, nil
}
