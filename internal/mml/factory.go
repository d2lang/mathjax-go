// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package mml

import "github.com/d2lang/mathjax-go/internal/ordered"

// Definition describes behavior shared by all nodes of one kind.
type Definition struct {
	Defaults *ordered.Map[Property]
	Flags    Flags
}

// Factory constructs nodes with registered kind definitions and shared global
// math attributes. It is intentionally small so parser and layout packages can
// evolve independently.
type Factory struct {
	definitions map[string]Definition
	global      *ordered.Map[Property]
}

// NewFactory returns an empty factory.
func NewFactory() *Factory {
	return &Factory{
		definitions: make(map[string]Definition),
		global:      ordered.New[Property](),
	}
}

// Globals returns the shared global math-attribute map.
func (f *Factory) Globals() *ordered.Map[Property] { return f.global }

// Register registers or replaces a node definition.
func (f *Factory) Register(kind string, definition Definition) {
	definition.Defaults = definition.Defaults.Clone()
	f.definitions[kind] = definition
}

// Definition returns a registered definition.
func (f *Factory) Definition(kind string) (Definition, bool) {
	definition, ok := f.definitions[kind]
	return definition, ok
}

// Create constructs kind and links children.
func (f *Factory) Create(kind string, children ...*Node) *Node {
	definition := f.definitions[kind]
	node := NewNode(kind, definition.Defaults, f.global, children...)
	node.Flags = definition.Flags
	return node
}

// Text constructs an internal text node.
func (f *Factory) Text(text string) *Node { return NewText(text) }
