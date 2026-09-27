// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package svg

import (
	"fmt"
	"github.com/d2lang/mathjax-go/internal/mml"
	"testing"
)

func TestTeXAtomActualClassOverridesProperty(t *testing.T) {
	for _, class := range []mml.TeXClass{mml.TeXClassOrd, mml.TeXClassVCenter, mml.TeXClassNone} {
		t.Run(fmt.Sprint(class), func(t *testing.T) {
			n := mml.NewNode("TeXAtom", nil, nil)
			n.TeXClass = class
			n.SetProperty("texClass", mml.TeXClassVCenter)
			if got := effectiveTeXClass(n); got != class {
				t.Errorf("actual %v became %v through property fallback", class, got)
			}
		})
	}
	// Preserve accepted property fallback for ordinary nodes such as invisible
	// ApplyFunction; this fix is specific to TeXAtom's concrete class field.
	n := mml.NewNode("mo", nil, nil)
	n.SetProperty("texClass", mml.TeXClassOp)
	if got := effectiveTeXClass(n); got != mml.TeXClassOp {
		t.Errorf("mo property fallback = %v", got)
	}
	n.SetProperty("texClass", mml.TeXClassNone)
	if got := effectiveTeXClass(n); got != mml.TeXClassNone {
		t.Errorf("invisible mo class = %v", got)
	}
}
