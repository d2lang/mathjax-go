// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package svg

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/tex"
)

func TestMmlTokenFrozenSVG(t *testing.T) {
	data, err := os.ReadFile("../tex/testdata/mml_token_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name, TeX string
			SVGSHA256 *string
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, test := range fixture.Cases {
		if test.SVGSHA256 == nil {
			continue
		}
		count++
		t.Run(test.Name, func(t *testing.T) {
			root, err := tex.NewCompiler().Compile(test.TeX, true)
			if err != nil {
				t.Fatal(err)
			}
			options := pipeline.DefaultOptions()
			options.Display = true
			got, err := NewTypesetter().Typeset(root, options)
			if err != nil {
				t.Fatal(err)
			}
			if hash := fmt.Sprintf("%x", sha256.Sum256([]byte(got))); hash != *test.SVGSHA256 {
				t.Fatalf("complete SVG = %s, want pinned %s", hash, *test.SVGSHA256)
			}
		})
	}
	if count != 24 {
		t.Fatal("incomplete unspaced complete-SVG controls")
	}
}
