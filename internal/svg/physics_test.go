// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package svg

import (
	"testing"

	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/tex"
)

func TestCompiledPhysicsOperatorsFrozenSVG(t *testing.T) {
	tests := []struct {
		source string
		hash   string
	}{
		{`\sin[2](x)`, "dbe88b159de297d51e23c696663bfb828ab07598811e24fd7cb7a2bf836feac4"},
		{`\divergence(A)`, "a6b497236c00000098c11a7fc924de0b9baac4d65216e0f3510acb5c696d745f"},
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
