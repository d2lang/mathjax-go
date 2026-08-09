// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package mml

import "github.com/d2lang/mathjax-go/internal/ordered"

// Property is a MathML attribute or internal-property value.
type Property = any

type inheritValue struct{}

// Inherit marks a default that resolves directly from the math node's global
// attributes, matching MathJax's _inherit_ sentinel.
var Inherit Property = inheritValue{}

func isInherit(value Property) bool {
	_, ok := value.(inheritValue)
	return ok
}

// IsInherit reports whether value is the Inherit sentinel.
func IsInherit(value Property) bool { return isInherit(value) }

// Attributes preserves MathJax's four attribute layers: explicit, inherited,
// per-node defaults, and global math attributes.
type Attributes struct {
	explicit  *ordered.Map[Property]
	inherited *ordered.Map[Property]
	defaults  *ordered.Map[Property]
	global    *ordered.Map[Property]
}

// NewAttributes constructs an attribute set. Inputs are cloned; nil inputs are
// treated as empty maps.
func NewAttributes(defaults, global *ordered.Map[Property]) *Attributes {
	return &Attributes{
		explicit:  ordered.New[Property](),
		inherited: ordered.New[Property](),
		defaults:  defaults.Clone(),
		global:    global.Clone(),
	}
}

// Set sets an explicit attribute.
func (a *Attributes) Set(name string, value Property) { a.explicit.Set(name, value) }

// SetList sets explicit attributes in the source map's insertion order.
func (a *Attributes) SetList(list *ordered.Map[Property]) {
	list.Range(func(name string, value Property) bool {
		a.Set(name, value)
		return true
	})
}

func lookup(layers []*ordered.Map[Property], name string) (Property, bool) {
	for _, layer := range layers {
		if value, ok := layer.Get(name); ok {
			return value, true
		}
	}
	return nil, false
}

// Get resolves explicit, inherited, default, and global layers in that order.
func (a *Attributes) Get(name string) (Property, bool) {
	value, ok := lookup([]*ordered.Map[Property]{a.explicit, a.inherited, a.defaults, a.global}, name)
	if ok && isInherit(value) {
		return a.global.Get(name)
	}
	return value, ok
}

// GetExplicit returns only an explicitly set value.
func (a *Attributes) GetExplicit(name string) (Property, bool) { return a.explicit.Get(name) }

// SetInherited sets an inherited value.
func (a *Attributes) SetInherited(name string, value Property) { a.inherited.Set(name, value) }

// GetInherited resolves the inherited layer and its default/global prototypes.
func (a *Attributes) GetInherited(name string) (Property, bool) {
	return lookup([]*ordered.Map[Property]{a.inherited, a.defaults, a.global}, name)
}

// GetDefault resolves the per-kind default and global layers.
func (a *Attributes) GetDefault(name string) (Property, bool) {
	return lookup([]*ordered.Map[Property]{a.defaults, a.global}, name)
}

// IsSet reports whether name is explicit or directly inherited.
func (a *Attributes) IsSet(name string) bool {
	return a.explicit.Has(name) || a.inherited.Has(name)
}

// HasDefault reports whether name has a kind or global default.
func (a *Attributes) HasDefault(name string) bool {
	return a.defaults.Has(name) || a.global.Has(name)
}

func (a *Attributes) ExplicitNames() []string  { return a.explicit.Keys() }
func (a *Attributes) InheritedNames() []string { return a.inherited.Keys() }
func (a *Attributes) DefaultNames() []string   { return a.defaults.Keys() }
func (a *Attributes) GlobalNames() []string    { return a.global.Keys() }

// Explicit returns the mutable explicit-attribute map.
func (a *Attributes) Explicit() *ordered.Map[Property] { return a.explicit }

// Inherited returns the mutable inherited-attribute map.
func (a *Attributes) Inherited() *ordered.Map[Property] { return a.inherited }

// Defaults returns the per-kind default map.
func (a *Attributes) Defaults() *ordered.Map[Property] { return a.defaults }

// Globals returns the global math-attribute map.
func (a *Attributes) Globals() *ordered.Map[Property] { return a.global }

// Clone returns an independent shallow copy.
func (a *Attributes) Clone() *Attributes {
	if a == nil {
		return NewAttributes(nil, nil)
	}
	return &Attributes{
		explicit:  a.explicit.Clone(),
		inherited: a.inherited.Clone(),
		defaults:  a.defaults.Clone(),
		global:    a.global.Clone(),
	}
}
