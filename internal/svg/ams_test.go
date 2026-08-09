// Copyright 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package svg

import (
	"testing"

	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/tex"
)

func TestCompiledAMSFrozenSVG(t *testing.T) {
	tests := []struct {
		source string
		hash   string
	}{
		{`\operatorname*{arg\,max}_{x}`, "2aba4909e747b321911a61427ba6973579188fcba52dabeafd67e65be4ff8774"},
		{`\begin{equation}a=b\label{eq:a}\end{equation}\eqref{eq:a}`, "ba1e526dbf1754a9dd1ef91673f24668065ddf23a57d74625268a69194f301c6"},
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
