// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Based on BaseMethods.Not and BaseItems.NotItem (MathJax 3.2.2).
package tex

import (
	"unicode/utf16"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// This is an explicit data-map consumer. It does not change executable-map
// eligibility or route arbitrary names through a parser-less map.
func notRemap(text string) (string, bool) {
	for i := len(mjSourceMaps) - 1; i >= 0; i-- {
		m := mjSourceMaps[i]
		if m.Name != "not_remap" || m.Kind != mjSourceCharacterMap {
			continue
		}
		for j := len(m.Entries) - 1; j >= 0; j-- {
			entry := m.Entries[j]
			if entry.Name == text {
				value, ok := entry.Value.(string)
				return value, ok
			}
		}
		return "", false
	}
	return "", false
}

func notFallback() *mml.Node {
	// Primary NotItem uses the node factory, not the token factory.
	text := node("mtext", mml.NewText("\u29F8"))
	padded := node("mpadded", text)
	padded.Attributes.Set("width", 0)
	return texAtom(padded, mml.TeXClassRel)
}

// A true result means the original eligible token was changed in place.
// The caller delivers that same token and consumes the pending Not.
func applyNotToken(n *mml.Node) bool {
	if n == nil || (n.Kind != "mo" && n.Kind != "mi" && n.Kind != "mtext") || len(n.Children) != 1 {
		return false
	}
	text := textContent(n)
	moves, _ := n.Property("movesupsub")
	if len(utf16.Encode([]rune(text))) != 1 || limitsTruthy(moves) {
		return false
	}
	if replacement, ok := notRemap(text); ok {
		child := mml.NewText(replacement)
		// NodeUtil.setChild overwrites only the slot and new parent. It does
		// not clear the removed text node's parent pointer.
		n.Children[0] = child
		child.Parent = n
	} else {
		n.AppendChild(mml.NewText("\u0338"))
	}
	return true
}

// A NotItem belongs to one logical row, not the shared macro configuration.
// Empty command results leave it pending until a real stack item arrives.
type pendingNot bool

func (pending *pendingNot) finish() []*mml.Node {
	if !*pending {
		return nil
	}
	*pending = false
	return []*mml.Node{notFallback()}
}

func (pending *pendingNot) start() []*mml.Node {
	nodes := pending.finish()
	*pending = true
	return nodes
}

func (pending *pendingNot) apply(nodes []*mml.Node) []*mml.Node {
	if !*pending || len(nodes) == 0 {
		return nodes
	}
	*pending = false
	if applyNotToken(nodes[0]) {
		return nodes
	}
	return append([]*mml.Node{notFallback()}, nodes...)
}
