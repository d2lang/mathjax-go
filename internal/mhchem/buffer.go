// Copyright (c) 2015-2021 Martin Hensel
// Copyright (c) 2026 luo-studio
// Copyright (c) 2026 erweixin (upstream RaTeX)
// SPDX-License-Identifier: Apache-2.0 AND MIT
//
// Go translation and modification of mhchemparser 4.1.1. Initial Go
// translation structure adapted from github.com/Luo-Studio/go-tex (MIT).

// Package mhchem ports mhchemparser 4.1.1, the parser used by MathJax 3.2.2.
// ToTeX converts \ce and \pu source into a TeX expansion for a host parser.
package mhchem

// Buffer mirrors the mhchemparser buffer object used by the action
// handlers in actions.go. Slots a..rm are typed as *string so we can
// distinguish "unset" from "empty string".
type Buffer struct {
	ParenthesisLevel int
	BeginsWithBond   bool
	Sb               bool
	A                *string
	B                *string
	P                *string
	O                *string
	Q                *string
	D                *string
	DType            *string
	R                *string
	Rdt              *string
	Rd               *string
	Rqt              *string
	Rq               *string
	Text             *string
	Rm               *string
}

// NewBuffer returns a fresh, zeroed Buffer.
func NewBuffer() *Buffer { return &Buffer{} }

// ClearSoft clears all slots except ParenthesisLevel/BeginsWithBond.
func (b *Buffer) ClearSoft() {
	b.Sb = false
	b.A, b.B, b.P, b.O, b.Q = nil, nil, nil, nil, nil
	b.D, b.DType, b.R = nil, nil, nil
	b.Rdt, b.Rd, b.Rqt, b.Rq = nil, nil, nil, nil
	b.Text, b.Rm = nil, nil
}

// ClearAll resets the buffer to its zero value.
func (b *Buffer) ClearAll() { *b = Buffer{} }

// IsSlotEmpty reports whether s is nil or points to an empty string.
func IsSlotEmpty(s *string) bool { return s == nil || *s == "" }

// SetSlot writes v to *slot, allocating if nil.
func SetSlot(slot **string, v string) {
	if *slot == nil {
		s := v
		*slot = &s
		return
	}
	**slot = v
}
