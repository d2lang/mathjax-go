// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package svg

import (
	"testing"

	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/tex"
)

func TestCompiledAccentsFrozenSVG(t *testing.T) {
	tests := []struct {
		source string
		hash   string
	}{
		{`\mathring{x}`, "52a0b23de9139863d0dd4565315a3e672d00e179bf43f12d682a1ad9ce64560f"},
		{`\va*{x}`, "543713247ad5316ada2fe81f6665d44af924bd7488b8d08648fbef03d73b2b44"},
		{`\overbracket[.2em][.3em]{x+y}`, "e20afb953e3b510bfba60ab9409dfdcaaab26c57a5c38e053040fdf70214a860"},
	}
	for _, test := range tests {
		root, err := tex.NewCompiler().Compile(test.source, true)
		if err != nil {
			t.Fatal(err)
		}
		got, err := NewTypesetter().Typeset(root, pipeline.DefaultOptions())
		if err != nil {
			t.Fatal(err)
		}
		requireFrozenWrapperHash(t, got, test.hash)
	}
}
