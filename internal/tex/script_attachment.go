// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Based on BaseMethods.Subscript and BaseMethods.Superscript (MathJax 3.2.2).
package tex

import "github.com/d2lang/mathjax-go/internal/mml"

// attachScriptBase preserves authored annotation ownership. The source tests
// family/subtype and occupied slots before choosing whether to reuse a script
// wrapper or wrap the entire base. This parser's own eager compact subtypes
// represent an unfinished generic family; their transient origin bridges that
// representation only. Pending PrimeItems pass their original base here.
func attachScriptBase(base, script *mml.Node, marker byte, moves bool) (*mml.Node, error) {
	attachment, err := prepareScriptAttachment(base, marker, moves)
	if err != nil {
		return nil, err
	}
	return attachment.fill(script), nil
}

// scriptAttachment is a read-only decision until a real argument is available.
// The pending argument uses the same family/occupied-slot policy as final fill.
type scriptAttachment struct {
	base, under, over    *mml.Node
	marker               byte
	moves, limits, reuse bool
}

func prepareScriptAttachment(base *mml.Node, marker byte, moves bool) (*scriptAttachment, error) {
	origin, _ := base.Property(limitsScriptOrigin)
	eager := origin == true
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
	return &scriptAttachment{base, under, over, marker, moves, limits, reuse}, nil
}

// pendingBase exposes the source's unfinished generic family only to read-only
// handler checks. It shares the actual children and own maps, without changing
// their parent pointers or publishing a temporary wrapper into the MML tree.
func (a *scriptAttachment) pendingBase() *mml.Node {
	kind := "msubsup"
	if a.limits && a.reuse || a.moves && !a.reuse {
		kind = "munderover"
	}
	if a.reuse {
		view := *a.base
		if origin, _ := a.base.Property(limitsScriptOrigin); origin == true {
			view.Kind = kind
			if a.base.Kind == "msup" || a.base.Kind == "mover" {
				view.Children = []*mml.Node{a.base.Children[0], nil, a.over}
			}
		}
		return &view
	}
	view := node(kind)
	view.Children = []*mml.Node{a.base}
	if a.moves {
		view.SetProperty("movesupsub", true)
	}
	return view
}

func (a *scriptAttachment) fill(script *mml.Node) *mml.Node {
	base, under, over := a.base, a.under, a.over
	marker, moves, limits, reuse := a.marker, a.moves, a.limits, a.reuse
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
		return result
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
	base.Kind = kind
	base.Flags.Arity = len(children)
	base.SetChildren(children)
	refreshDynamicFlags(base)
	return base
}
