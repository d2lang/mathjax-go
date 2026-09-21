// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"github.com/d2lang/mathjax-go/internal/mml"
	"testing"
)

func TestResolvedFontScopeSurvivesParserClone(t *testing.T) {
	inner := ambientFontToken(token("mi", "x"))
	applyScopedMathVariant(inner, "normal")
	copied := inner.Clone()
	sibling := ambientFontToken(token("mi", "y"))
	kept := ambientFontToken(token("mi", "R"))
	kept.Attributes.Set("mathvariant", "double-struck")
	kept.Attributes.Set("mjx-keep-attrs", "id mathvariant")
	applyScopedMathVariant(node("mrow", copied, sibling, kept), "bold")
	for _, c := range []struct {
		n    *mml.Node
		want string
	}{{inner, "normal"}, {copied, "normal"}, {sibling, "bold"}, {kept, "double-struck"}} {
		if got, _ := c.n.Attributes.GetExplicit("mathvariant"); got != c.want {
			t.Fatalf("font choice %v, want %s", got, c.want)
		}
	}
}
