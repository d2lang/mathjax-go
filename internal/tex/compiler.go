// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

// Package tex is a source-shaped Go port of MathJax 3.2.2's TeX input jax.
// It compiles the fixed package set used by D2 (base, AMS, mathtools, AMS-CD,
// braket, cancel, cases, color, gensymb, mhchem, and physics) into the shared
// internal MathML tree.
package tex

import (
	"errors"

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
)

const maxMacros = 1000

// Compiler implements pipeline.Compiler.  Compiler has no mutable parse state;
// every Compile call receives fresh macro and environment tables, as does the
// MathJax TeX input jax after ParseOptions.clear().
type Compiler struct {
	// augmentedPackages is used only by the pinned source-handler oracle,
	// whose provenance explicitly loads the official empheq and newcommand
	// components in addition to D2's package list.  Public compilers keep the
	// exact fixed D2 registrations.
	augmentedPackages bool
}

// NewCompiler constructs the fixed MathJax 3.2.2 compiler used by D2.
func NewCompiler() *Compiler { return &Compiler{} }

var _ pipeline.Compiler = (*Compiler)(nil)

// Compile parses one TeX expression and wraps the result in MathJax's math
// node. Display mode is represented by the explicit display="block"
// attribute, matching TeX.compile in ts/input/tex.ts.
func (c *Compiler) Compile(source string, display bool) (*mml.Node, error) {
	state := newParseState()
	state.augmentedPackages = c.augmentedPackages
	p := &parser{source: source, state: state, display: display}
	children, stop, err := p.parseRow(0, false)
	if err != nil {
		var parseError *Error
		if !errors.As(err, &parseError) {
			return nil, err
		}
		return mathError(parseError.Message, display), nil
	}
	if stop != "" {
		return nil, texError("ExtraCloseMissingOpen", "Extra close brace or missing open brace")
	}
	children, err = p.amsTagFinalize(children)
	if err != nil {
		var parseError *Error
		if !errors.As(err, &parseError) {
			return nil, err
		}
		return mathError(parseError.Message, display), nil
	}
	// MmlMath has an inferred-mrow child even when the TeX stack reduces to a
	// single node. Keep that wrapper explicit in the shared Go tree so layout
	// sees the same structure as the JavaScript MML factory.
	root := node("math", children...)
	if display {
		root.Attributes.Set("display", "block")
	}
	root.Walk(func(n *mml.Node) bool {
		n.RemoveProperty(resolvedFontScope)
		return true
	})
	setMathMLInheritance(root, display)
	cleanMathMLAttributes(root)
	return root, nil
}

func mathError(message string, display bool) *mml.Node {
	merror := setAttributes(node("merror", token("mtext", message)), map[string]any{"data-mjx-error": message})
	root := node("math", merror)
	if display {
		root.Attributes.Set("display", "block")
	}
	setMathMLInheritance(root, display)
	cleanMathMLAttributes(root)
	return root
}
