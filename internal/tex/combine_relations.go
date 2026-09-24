// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Source: ts/input/tex/FilterUtil.ts (combineRelations).

package tex

import (
	"reflect"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// relationUndefined distinguishes an absent JavaScript field (or an own field
// explicitly containing undefined) from an own null value in method inputs.
type relationUndefined struct{}

func relationField(value any, exists bool) any {
	if !exists {
		return relationUndefined{}
	}
	return value
}

func relationNullish(value any, exists bool) bool {
	_, undefined := value.(relationUndefined)
	return !exists || undefined || value == nil
}

func relationTruthy(value any) bool {
	if _, undefined := value.(relationUndefined); undefined {
		return false
	}
	return limitsTruthy(value)
}

// Attribute values are JS scalars. Go's numeric aliases represent one JS
// Number type; booleans, strings, null and undefined remain distinct. Internal
// object-valued properties compare by identity, never by structural equality.
func relationStrictEqual(a, b any) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	x, y := reflect.ValueOf(a), reflect.ValueOf(b)
	number := func(v reflect.Value) (float64, bool) {
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return float64(v.Int()), true
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return float64(v.Uint()), true
		case reflect.Float32, reflect.Float64:
			return v.Float(), true
		}
		return 0, false
	}
	if xn, ok := number(x); ok {
		yn, ok := number(y)
		return ok && xn == yn
	}
	if x.Type() != y.Type() {
		return false
	}
	if x.Kind() == reflect.Map {
		return x.UnsafePointer() == y.UnsafePointer()
	}
	return x.Comparable() && a == b
}

func relationCompareExplicit(left, right *mml.Node) bool {
	names := func(n *mml.Node, excluded string) []string {
		result := []string{}
		for _, name := range n.Attributes.ExplicitNames() {
			value, _ := n.Attributes.GetExplicit(name)
			if name != excluded && (name != "stretchy" || relationTruthy(value)) {
				result = append(result, name)
			}
		}
		return result
	}
	a, b := names(left, "lspace"), names(right, "rspace")
	if len(a) != len(b) {
		return false
	}
	// This asymmetry is intentional: the source compares only the left names,
	// after checking the counts, including a right name excluded above.
	for _, name := range a {
		if !relationStrictEqual(relationField(left.Attributes.GetExplicit(name)), relationField(right.Attributes.GetExplicit(name))) {
			return false
		}
	}
	return true
}

// combineRelations consumes the actual per-parse creation list. Parent-chain
// membership matches ParseOptions.getList; final-tree order is not equivalent.
func combineRelations(root *mml.Node, operators []*mml.Node) []*mml.Node {
	live := make([]*mml.Node, 0, len(operators))
	for _, op := range operators {
		for n := op; n != nil; n = n.Parent {
			if n == root {
				live = append(live, op)
				break
			}
		}
	}
	removed := make(map[*mml.Node]bool)
	for _, op := range live {
		combined, _ := op.Property("relationsCombined")
		parent := op.Parent
		if relationTruthy(combined) || parent == nil || parent.Kind != "mrow" || op.TeXClass != mml.TeXClassRel {
			continue
		}
		next := 0
		for i, child := range parent.Children {
			if child == op {
				next = i + 1
				break
			}
		}
		variant := relationField(op.Property("variantForm"))
		for next < len(parent.Children) {
			other := parent.Children[next]
			if other == nil || other.Kind != "mo" || other.TeXClass != mml.TeXClassRel {
				break
			}
			if !relationStrictEqual(variant, relationField(other.Property("variantForm"))) || !relationCompareExplicit(op, other) {
				if relationNullish(op.Attributes.GetExplicit("rspace")) {
					op.Attributes.Set("rspace", "0pt")
				}
				if relationNullish(other.Attributes.GetExplicit("lspace")) {
					other.Attributes.Set("lspace", "0pt")
				}
				break
			}
			for _, child := range other.Children {
				op.AppendChild(child)
			}
			for _, name := range []string{"stretchy", "rspace"} {
				value, exists := other.Attributes.GetExplicit(name)
				if !relationNullish(value, exists) {
					op.Attributes.Set(name, value)
				}
			}
			other.Properties.Range(func(name string, value any) bool {
				op.SetProperty(name, value)
				return true
			})
			parent.Children = append(parent.Children[:next], parent.Children[next+1:]...)
			other.Parent = nil
			// MathJax reads these child-dependent flags through dynamic
			// getters. Refresh the Go caches before operatorForms or a later
			// filter iteration observes the newly combined row.
			for ancestor := parent; ancestor != nil; ancestor = ancestor.Parent {
				refreshDynamicFlags(ancestor)
			}
			other.SetProperty("relationsCombined", true)
			removed[other] = true
		}
		op.Attributes.SetInherited("form", operatorForms(op)[0])
	}
	result := live[:0]
	for _, op := range live {
		if !removed[op] {
			result = append(result, op)
		}
	}
	return result
}
