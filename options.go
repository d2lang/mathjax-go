// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package mathjax

import "github.com/d2lang/mathjax-go/internal/pipeline"

// FontCacheMode controls how glyph paths are represented in rendered SVG.
// MathJax-go initially supports FontCacheNone, which is the mode used by D2's
// frozen MathJax 3.2.2 component.
type FontCacheMode uint8

const (
	// FontCacheNone emits an independent SVG path for each glyph occurrence.
	FontCacheNone FontCacheMode = iota
)

// Options controls TeX-to-SVG conversion.
type Options struct {
	// Em is the em size in pixels used to resolve relative lengths.
	Em float64
	// Ex is the ex size in pixels used to resolve relative lengths and report
	// the serialized SVG dimensions.
	Ex float64
	// Display selects display math when true and inline math when false.
	Display bool
	// FontCache controls SVG glyph reuse. Only FontCacheNone is supported.
	FontCache FontCacheMode
}

// DefaultOptions returns the options used by D2's frozen MathJax 3.2.2
// integration.
func DefaultOptions() Options {
	return Options{Em: 16, Ex: 8, Display: true, FontCache: FontCacheNone}
}

func (o Options) pipelineOptions() pipeline.Options {
	return pipeline.Options{
		Em:        o.Em,
		Ex:        o.Ex,
		Display:   o.Display,
		FontCache: pipeline.FontCacheMode(o.FontCache),
	}
}
