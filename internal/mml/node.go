// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

// Package mml defines MathJax's internal MathML tree representation.
package mml

import (
	"fmt"

	"github.com/d2lang/mathjax-go/internal/ordered"
)

// TeXClass is a TeX atom class used by MathJax's spacing rules.
type TeXClass int8

const (
	TeXClassNone    TeXClass = -1
	TeXClassOrd     TeXClass = 0
	TeXClassOp      TeXClass = 1
	TeXClassBin     TeXClass = 2
	TeXClassRel     TeXClass = 3
	TeXClassOpen    TeXClass = 4
	TeXClassClose   TeXClass = 5
	TeXClassPunct   TeXClass = 6
	TeXClassInner   TeXClass = 7
	TeXClassVCenter TeXClass = 8
)

var TeXClassNames = [...]string{
	"ORD", "OP", "BIN", "REL", "OPEN", "CLOSE", "PUNCT", "INNER", "VCENTER",
}

var texSpaceLength = [...]string{"", "thinmathspace", "mediummathspace", "thickmathspace"}

var texSpace = [8][8]int8{
	{0, -1, 2, 3, 0, 0, 0, 1},
	{-1, -1, 0, 3, 0, 0, 0, 1},
	{2, 2, 0, 0, 2, 0, 0, 2},
	{3, 3, 0, 0, 3, 0, 0, 3},
	{0, 0, 0, 0, 0, 0, 0, 0},
	{0, -1, 2, 3, 0, 0, 0, 1},
	{1, 1, 0, 1, 1, 1, 1, 1},
	{1, -1, 2, 3, 1, 0, 1, 1},
}

// TeXSpacing returns MathJax's named space between two TeX atom classes.
// script suppresses entries marked as display/text-style-only. Negative table
// entries are the ones retained in scripts in MathJax's TEXSPACE encoding.
func TeXSpacing(left, right TeXClass, script bool) string {
	if left < TeXClassOrd || left > TeXClassInner || right < TeXClassOrd || right > TeXClassInner {
		return ""
	}
	space := texSpace[left][right]
	if script && space >= 0 {
		return ""
	}
	if space < 0 {
		space = -space
	}
	return texSpaceLength[space]
}

// Flags contains node-kind behavior that MathJax expresses through its MML
// node subclasses.
type Flags struct {
	Token              bool
	Embellished        bool
	Spacelike          bool
	LinebreakContainer bool
	HasNewline         bool
	Inferred           bool
	NotParent          bool
	Arity              int
	CoreIndex          int
}

// Node is one node in MathJax's internal MathML tree. Kind-specific state that
// is not an attribute lives in Properties.
type Node struct {
	Kind       string
	Parent     *Node
	Children   []*Node
	Attributes *Attributes
	Properties *ordered.Map[Property]
	Text       string
	Flags      Flags
	TeXClass   TeXClass
	PrevClass  TeXClass
	PrevLevel  int
}

// NewNode creates a node with empty attribute and property maps.
func NewNode(kind string, defaults, global *ordered.Map[Property], children ...*Node) *Node {
	n := &Node{
		Kind:       kind,
		Attributes: NewAttributes(defaults, global),
		Properties: ordered.New[Property](),
		TeXClass:   TeXClassNone,
		PrevClass:  TeXClassNone,
	}
	n.SetChildren(children)
	return n
}

// NewText creates an internal text node.
func NewText(text string) *Node {
	n := NewNode("text", nil, nil)
	n.Text = text
	return n
}

// IsKind reports whether n has kind.
func (n *Node) IsKind(kind string) bool { return n != nil && n.Kind == kind }

// SetProperty sets internal node state.
func (n *Node) SetProperty(name string, value Property) { n.Properties.Set(name, value) }

// Property returns internal node state.
func (n *Node) Property(name string) (Property, bool) { return n.Properties.Get(name) }

// RemoveProperty removes internal node state.
func (n *Node) RemoveProperty(name string) bool { return n.Properties.Delete(name) }

// SetChildren replaces n's children and maintains parent pointers.
func (n *Node) SetChildren(children []*Node) {
	for _, child := range n.Children {
		if child != nil && child.Parent == n {
			child.Parent = nil
		}
	}
	n.Children = append(n.Children[:0], children...)
	for _, child := range n.Children {
		if child != nil {
			child.Parent = n
		}
	}
}

// AppendChild appends child and returns it.
func (n *Node) AppendChild(child *Node) *Node {
	if child != nil {
		child.Parent = n
	}
	n.Children = append(n.Children, child)
	return child
}

// ChildIndex returns child's index, or -1 if child is absent.
func (n *Node) ChildIndex(child *Node) int {
	for i, candidate := range n.Children {
		if candidate == child {
			return i
		}
	}
	return -1
}

// ReplaceChild replaces oldChild and returns an error when it is absent.
func (n *Node) ReplaceChild(newChild, oldChild *Node) error {
	i := n.ChildIndex(oldChild)
	if i < 0 {
		return fmt.Errorf("mml: child is not present in %s", n.Kind)
	}
	if oldChild != nil && oldChild.Parent == n {
		oldChild.Parent = nil
	}
	if newChild != nil {
		newChild.Parent = n
	}
	n.Children[i] = newChild
	return nil
}

// RemoveChild removes child and returns an error when it is absent.
func (n *Node) RemoveChild(child *Node) error {
	i := n.ChildIndex(child)
	if i < 0 {
		return fmt.Errorf("mml: child is not present in %s", n.Kind)
	}
	if child != nil && child.Parent == n {
		child.Parent = nil
	}
	copy(n.Children[i:], n.Children[i+1:])
	n.Children[len(n.Children)-1] = nil
	n.Children = n.Children[:len(n.Children)-1]
	return nil
}

// ParentNode returns the semantic parent, skipping inferred/not-parent nodes.
func (n *Node) ParentNode() *Node {
	parent := n.Parent
	for parent != nil && (parent.Flags.Inferred || parent.Flags.NotParent) {
		parent = parent.Parent
	}
	return parent
}

// Core returns the child containing an embellished operator, or n.
func (n *Node) Core() *Node {
	if n == nil || !n.Flags.Embellished || n.Flags.CoreIndex < 0 || n.Flags.CoreIndex >= len(n.Children) {
		return n
	}
	return n.Children[n.Flags.CoreIndex]
}

// CoreMO follows embellished children to their core mo node.
func (n *Node) CoreMO() *Node {
	current := n
	for current != nil && current.Flags.Embellished {
		next := current.Core()
		if next == current {
			break
		}
		current = next
	}
	return current
}

// Walk visits n and its descendants in pre-order.
func (n *Node) Walk(visit func(*Node) bool) {
	if n == nil || !visit(n) {
		return
	}
	for _, child := range n.Children {
		child.Walk(visit)
	}
}

// Find returns all descendants, including n, whose kind matches.
func (n *Node) Find(kind string) []*Node {
	var found []*Node
	n.Walk(func(node *Node) bool {
		if node.Kind == kind {
			found = append(found, node)
		}
		return true
	})
	return found
}

// Clone deep-copies n without retaining a parent pointer.
func (n *Node) Clone() *Node {
	if n == nil {
		return nil
	}
	clone := &Node{
		Kind:       n.Kind,
		Attributes: n.Attributes.Clone(),
		Properties: n.Properties.Clone(),
		Text:       n.Text,
		Flags:      n.Flags,
		TeXClass:   n.TeXClass,
		PrevClass:  n.PrevClass,
		PrevLevel:  n.PrevLevel,
	}
	children := make([]*Node, len(n.Children))
	for i, child := range n.Children {
		children[i] = child.Clone()
	}
	clone.SetChildren(children)
	return clone
}
