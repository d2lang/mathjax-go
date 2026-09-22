// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Based on FilterUtil.moveLimits and NodeUtil.copyAttributes.
package tex

import "github.com/d2lang/mathjax-go/internal/mml"

var movableScriptKinds = [][2]string{{"munderover", "msubsup"}, {"munder", "msub"}, {"mover", "msup"}}

// The compiler has a completed live tree rather than ParseOptions creation
// lists. Same-kind replacements commute for the live topology: they share
// attributes, copy properties and preserve current children/core delegation.
// Detached old-wrapper traces are deliberately not a compiler API contract.
func moveMathLimits(root *mml.Node) *mml.Node {
	lists := map[string][]*mml.Node{}
	root.Walk(func(n *mml.Node) bool {
		switch n.Kind {
		case "munderover", "munder", "mover":
			lists[n.Kind] = append(lists[n.Kind], n)
		}
		return true
	})
	return moveMathLimitsFromLists(root, lists)
}

// The ordered-list entry point follows the actual primary private method,
// including fixed family order and filtering detached registered nodes.
func moveMathLimitsFromLists(root *mml.Node, lists map[string][]*mml.Node) *mml.Node {
	for _, pair := range movableScriptKinds {
		retained := []*mml.Node{}
		for _, old := range lists[pair[0]] {
			live := old
			for live != nil && live != root {
				live = live.Parent
			}
			if live == nil {
				continue
			}
			display, _ := old.Attributes.Get("displaystyle")
			if limitsTruthy(display) || len(old.Children) == 0 || old.Children[0] == nil {
				retained = append(retained, old)
				continue
			}
			base := old.Children[0]
			core := limitsCore(base)
			if core == nil {
				retained = append(retained, old)
				continue
			}
			movable, _ := base.Property("movablelimits")
			explicit, _ := core.Attributes.GetExplicit("movablelimits")
			if !limitsTruthy(movable) || limitsTruthy(explicit) {
				retained = append(retained, old)
				continue
			}
			replacement := texMMLFactory.Create(pair[1])
			replacement.SetChildren(old.Children)
			replacement.Attributes = old.Attributes
			properties := mjSourceObject{}
			for _, key := range old.Properties.Keys() {
				value, _ := old.Property(key)
				properties = append(properties, mjSourceProperty{Name: key, Value: value})
			}
			applySourceObject(replacement, properties)
			refreshDynamicFlags(replacement)
			if old.Parent != nil {
				_ = old.Parent.ReplaceChild(replacement, old)
			} else {
				root = replacement
			}
		}
		lists[pair[0]] = retained
	}
	return root
}
