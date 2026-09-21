// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package tex

import "testing"

// MmlMo's numeric defaults are distinct from attributes. Range fallback and
// empty text are covered directly without relying on unsupported glyph parsing.
func TestOperatorMathMLSpacingDefaults(t *testing.T) {
	for _, test := range []struct {
		text, form  string
		left, right float64
	}{
		{"+", "prefix", 1.0 / 18, 2.0 / 18}, {"+", "infix", 5.0 / 18, 5.0 / 18},
		{"∑", "prefix", 2.0 / 18, 3.0 / 18}, {"(", "prefix", 1.0 / 18, 1.0 / 18},
		{"☃", "infix", 1.0 / 18, 1.0 / 18}, {"", "infix", 1.0 / 18, 1.0 / 18},
	} {
		t.Run(test.text+test.form, func(t *testing.T) {
			n := token("mo", test.text)
			n.Attributes.Set("form", test.form)
			applyOperatorInheritance(n)
			if n.OperatorLspace != test.left || n.OperatorRspace != test.right {
				t.Fatalf("defaults %g/%g, want %g/%g", n.OperatorLspace, n.OperatorRspace, test.left, test.right)
			}
			if n.Attributes.IsSet("lspace") || n.Attributes.IsSet("rspace") {
				t.Fatal("dictionary defaults became authored spacing")
			}
			clone := n.Clone()
			if clone.OperatorLspace != n.OperatorLspace || clone.OperatorRspace != n.OperatorRspace {
				t.Fatal("clone lost numeric defaults")
			}
			clone.OperatorLspace = 7
			if n.OperatorLspace == 7 {
				t.Fatal("clone mutated original")
			}
		})
	}
}
