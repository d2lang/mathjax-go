// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package svg

import (
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

func TestBinaryOperatorPreviousNullVersusExplicitNone(t *testing.T) {
	tests := []struct {
		name     string
		property any
	}{
		{"unset fraction class", nil},
		{"explicit NONE class", mml.TeXClassNone},
		{"explicit NONE int", int(-1)},
		{"explicit NONE int64", int64(-1)},
		{"explicit NONE float", float64(-1)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			previous := mml.NewNode("mfrac", nil, nil)
			// The Go MML tree records an explicitly supplied NONE in the
			// texClass property; an absent property is the upstream null.
			if test.property != nil {
				previous.SetProperty("texClass", test.property)
			}
			operator := mml.NewNode("mo", nil, nil, mml.NewText("+"))
			operator.TeXClass = mml.TeXClassBin
			adjustTeXClass(operator, previous)
			wantPrevious, wantOperator := mml.TeXClassOrd, mml.TeXClassBin
			if test.property != nil {
				wantPrevious, wantOperator = mml.TeXClassNone, mml.TeXClassOrd
			}
			if operator.PrevClass != wantPrevious || operator.TeXClass != wantOperator {
				t.Fatalf("previous/operator = %d/%d, want %d/%d", operator.PrevClass, operator.TeXClass, wantPrevious, wantOperator)
			}
			if previous.TeXClass != mml.TeXClassNone {
				t.Fatal("reading a null previous class must not rewrite the source tree")
			}
		})
	}
}
