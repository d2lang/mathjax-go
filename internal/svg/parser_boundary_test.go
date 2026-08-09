// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
//
// Frozen SVG hashes were produced by D2's embedded MathJax 3.2.2 Goja
// component (polyfills.js, mathjax.js, and setup.js), not a current Node
// MathJax installation.

package svg

import (
	"testing"

	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/tex"
)

func TestCompiledParserBoundariesFrozenSVG(t *testing.T) {
	tests := []struct {
		name, source, hash string
	}{
		{
			"named function group boundary",
			`\textcolor{red}{y} = \textcolor{green}{\sin} x`,
			"2f7c0e41866ec49817a3e8a240db821ea4218b21a50f496b86b84251dc60bfa3",
		},
		{
			"ordinary named function",
			`\sin x`,
			"c228482a36b98d24cb5db669b2ea5dbc6ba571979e4da7a7c80a77311a9a7de8",
		},
		{
			"mhchem multi-letter identifiers",
			`\ce{SO4^2- + Ba^2+ -> BaSO4 v}`,
			"112c24dc4b9a2664e2bf8c7f8415772c960e23ad4eeba0a29666e391bd50e08c",
		},
	}
	compiler := tex.NewCompiler()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root, err := compiler.Compile(test.source, true)
			if err != nil {
				t.Fatal(err)
			}
			got, err := NewTypesetter().Typeset(root, pipeline.DefaultOptions())
			if err != nil {
				t.Fatal(err)
			}
			requireFrozenWrapperHash(t, got, test.hash)
		})
	}
}
