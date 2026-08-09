// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

// Package pipeline defines the narrow contracts between the TeX parser, MathML
// tree, and SVG output stages.
package pipeline

import (
	"fmt"
	"math"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// FontCacheMode controls SVG glyph path caching.
type FontCacheMode uint8

const (
	// FontCacheNone emits a path for every glyph occurrence. This is the only
	// mode in D2's frozen MathJax component and is currently the only supported
	// mathjax-go mode.
	FontCacheNone FontCacheMode = iota
)

// Options controls a single conversion.
type Options struct {
	Em        float64
	Ex        float64
	Display   bool
	FontCache FontCacheMode
}

// DefaultOptions reproduces html.convert(tex, {em: 16, ex: 8}) with the SVG
// output jax configured as {fontCache: "none"}.
func DefaultOptions() Options {
	return Options{
		Em:        16,
		Ex:        8,
		Display:   true,
		FontCache: FontCacheNone,
	}
}

// Validate rejects values for which the MathJax 3.2.2 compatibility contract
// is undefined.
func (o Options) Validate() error {
	if math.IsNaN(o.Em) || math.IsInf(o.Em, 0) || o.Em <= 0 {
		return fmt.Errorf("mathjax-go: em must be finite and positive")
	}
	if math.IsNaN(o.Ex) || math.IsInf(o.Ex, 0) || o.Ex <= 0 {
		return fmt.Errorf("mathjax-go: ex must be finite and positive")
	}
	if o.FontCache != FontCacheNone {
		return fmt.Errorf("mathjax-go: unsupported font cache mode %d", o.FontCache)
	}
	return nil
}

// Compiler parses TeX into MathJax's internal MathML tree. Implementations
// must start each call with fresh macro, tag, and parser state.
type Compiler interface {
	Compile(tex string, display bool) (*mml.Node, error)
}

// Typesetter lays out an inherited-attribute-complete MathML tree and emits the
// bare SVG element used by D2.
type Typesetter interface {
	Typeset(root *mml.Node, options Options) (string, error)
}

// Pipeline joins independently implemented compiler and typesetter stages.
// Implementations should be immutable or allocate per-call state so Pipeline
// is safe for concurrent use.
type Pipeline struct {
	Compiler   Compiler
	Typesetter Typesetter
}

// Available reports whether both stages are connected.
func (p Pipeline) Available() bool {
	return p.Compiler != nil && p.Typesetter != nil
}

// Render runs a single stateless conversion.
func (p Pipeline) Render(tex string, options Options) (string, error) {
	if err := options.Validate(); err != nil {
		return "", err
	}
	if !p.Available() {
		return "", fmt.Errorf("mathjax-go: incomplete rendering pipeline")
	}
	root, err := p.Compiler.Compile(tex, options.Display)
	if err != nil {
		return "", err
	}
	if root == nil {
		return "", fmt.Errorf("mathjax-go: compiler returned a nil MathML root")
	}
	return p.Typesetter.Typeset(root, options)
}
