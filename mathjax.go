// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

// Package mathjax converts TeX math to SVG without a JavaScript runtime.
//
// Its compatibility target is the custom MathJax 3.2.2 component embedded by
// D2. Render uses that component's display-mode, 16-pixel em, 8-pixel ex, and
// uncached-font-path defaults.
package mathjax

import (
	"fmt"
	"math"
	"regexp"
	"strconv"

	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
	"github.com/d2lang/mathjax-go/internal/tex"
)

// CompatibilityVersion is the upstream MathJax version whose behavior this
// module preserves.
const CompatibilityVersion = "3.2.2"

var defaultPipeline = pipeline.Pipeline{
	Compiler:   tex.NewCompiler(),
	Typesetter: svg.NewTypesetter(),
}

var svgDimensions = regexp.MustCompile(`<svg[^>]+width="([0-9.]+)ex" height="([0-9.]+)ex"[^>]+>`)

// Render converts TeX to a bare SVG element using DefaultOptions. TeX syntax
// errors are represented by MathJax-compatible merror SVG; non-nil errors
// report an invalid option or an internal conversion failure.
func Render(tex string) (string, error) {
	return RenderWithOptions(tex, DefaultOptions())
}

// RenderWithOptions converts TeX to a bare SVG element. The returned string is
// the equivalent of the inner HTML of MathJax's mjx-container. TeX syntax
// errors are represented by MathJax-compatible merror SVG.
func RenderWithOptions(tex string, options Options) (string, error) {
	return defaultPipeline.Render(tex, options.pipelineOptions())
}

// Measure renders TeX with DefaultOptions and returns its rounded-up width and
// height in pixels. It preserves D2's MathJax 3.2.2 measurement contract: the
// serialized SVG dimensions are expressed in ex units and one ex is 8 pixels.
func Measure(tex string) (width, height int, err error) {
	svg, err := Render(tex)
	if err != nil {
		return 0, 0, err
	}
	matches := svgDimensions.FindStringSubmatch(svg)
	if len(matches) != 3 {
		return 0, 0, fmt.Errorf("mathjax-go: SVG dimensions not found")
	}
	widthEx, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("mathjax-go: parse SVG width: %w", err)
	}
	heightEx, err := strconv.ParseFloat(matches[2], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("mathjax-go: parse SVG height: %w", err)
	}
	return int(math.Ceil(widthEx * 8)), int(math.Ceil(heightEx * 8)), nil
}
