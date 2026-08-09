// Copyright (c) 2015-2021 Martin Hensel
// SPDX-License-Identifier: Apache-2.0

// Package mhchem is a native Go port of mhchemparser 4.1.1, the exact parser
// used by MathJax 3.2.2.  It expands chemistry and physical-unit notation into
// ordinary TeX for a host TeX parser to consume.
package mhchem

// Mode selects one of mhchemparser's three source state machines.
type Mode string

const (
	// ModeTeX scans ordinary TeX and expands nested \ce and \pu commands.
	ModeTeX Mode = "tex"
	// ModeCE parses the argument of \ce.
	ModeCE Mode = "ce"
	// ModePU parses the argument of \pu.
	ModePU Mode = "pu"
)

// ToTeX returns the exact TeX expansion produced by mhchemparser 4.1.1.
func ToTeX(input string, mode Mode) (string, error) {
	switch mode {
	case ModeTeX, ModeCE, ModePU:
	default:
		return "", errMsg("unknown mode %q (expected tex, ce, or pu)", mode)
	}
	data, err := LoadData()
	if err != nil {
		return "", err
	}
	context := &parserCtx{data: data}
	parsed, err := goMachine(context, input, string(mode))
	if err != nil {
		return "", err
	}
	return texify(parsed, mode != ModeTeX)
}
