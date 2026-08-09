// Copyright (c) 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Go registration model derived from MathJax 3.2.2's TeX configurations.

// Package spec defines dependency-free TeX extension registration data.
// Extension packages depend only on this package and the standard library;
// the core TeX parser can therefore import them without an import cycle.
package spec

// Package describes one MathJax TeX package and its single source symbol map.
// Entries in every slice retain their order in the MathJax 3.2.2 source.
type Package struct {
	Name           string
	MacroMap       string
	Commands       []Command
	Symbols        []Symbol
	Options        []Option
	AllowedOptions []string
}

// Command is one CommandMap entry. Handler names the extension operation;
// Arguments contains the fixed trailing arguments from an array-valued source
// entry, such as cancel's notation.
type Command struct {
	Name      string
	Handler   string
	Arguments []string
}

// Symbol is one CharacterMap entry and the token recipe its map handler uses.
type Symbol struct {
	Name       string
	Character  string
	TokenKind  string
	Attributes []Attribute
}

// Attribute is an ordered token or node attribute.
type Attribute struct {
	Name  string
	Value string
}

// Option is one package option and its source default value.
type Option struct {
	Name  string
	Value string
}
