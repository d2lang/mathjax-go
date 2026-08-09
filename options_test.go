// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package mathjax

import (
	"math"
	"strings"
	"testing"
)

func TestDefaultOptions(t *testing.T) {
	options := DefaultOptions()
	if options.Em != 16 || options.Ex != 8 || !options.Display || options.FontCache != FontCacheNone {
		t.Fatalf("DefaultOptions() = %+v", options)
	}
}

func TestRenderPipelineIsConnected(t *testing.T) {
	svg, err := Render(`a+b=c`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(svg, `<svg `) || !strings.Contains(svg, `data-mml-node="math"`) {
		t.Fatalf("Render returned unexpected SVG: %s", svg)
	}
}

func TestMeasurePreservesD2PixelContract(t *testing.T) {
	width, height, err := Measure(`a + b = c`)
	if err != nil {
		t.Fatal(err)
	}
	if width != 72 || height != 15 {
		t.Fatalf("Measure() = %dx%d, want 72x15", width, height)
	}
}

func TestRenderWithOptionsRejectsInvalidGeometryAndFontCache(t *testing.T) {
	tests := []struct {
		name    string
		options Options
	}{
		{"zero em", Options{Em: 0, Ex: 8, Display: true, FontCache: FontCacheNone}},
		{"negative em", Options{Em: -1, Ex: 8, Display: true, FontCache: FontCacheNone}},
		{"nan em", Options{Em: math.NaN(), Ex: 8, Display: true, FontCache: FontCacheNone}},
		{"infinite em", Options{Em: math.Inf(1), Ex: 8, Display: true, FontCache: FontCacheNone}},
		{"zero ex", Options{Em: 16, Ex: 0, Display: true, FontCache: FontCacheNone}},
		{"negative ex", Options{Em: 16, Ex: -1, Display: true, FontCache: FontCacheNone}},
		{"nan ex", Options{Em: 16, Ex: math.NaN(), Display: true, FontCache: FontCacheNone}},
		{"infinite ex", Options{Em: 16, Ex: math.Inf(1), Display: true, FontCache: FontCacheNone}},
		{"unsupported font cache", Options{Em: 16, Ex: 8, Display: true, FontCache: FontCacheMode(1)}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := RenderWithOptions("x", test.options); err == nil {
				t.Fatalf("RenderWithOptions accepted %+v", test.options)
			}
		})
	}
}
