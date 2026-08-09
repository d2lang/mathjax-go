// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package tex

import "testing"

// MathJax's BaseMethods.Accent stores mathaccent as an internal node property,
// not as a MathML attribute.  The distinction is visible in exact SVG output.
func TestAccentMathaccentProperty(t *testing.T) {
	root, err := NewCompiler().Compile(`\mathring{x}\va*{x}`, true)
	if err != nil {
		t.Fatal(err)
	}
	movers := root.Find("mover")
	if len(movers) != 2 {
		t.Fatalf("mover count = %d, want 2", len(movers))
	}
	for i, mover := range movers {
		if len(mover.Children) != 2 || mover.Children[1].Kind != "mo" {
			t.Fatalf("mover %d = %#v", i, mover)
		}
		if value, ok := mover.Children[1].Property("mathaccent"); !ok || value != true {
			t.Fatalf("mover %d mathaccent = %#v, present %v", i, value, ok)
		}
	}
	if value, _ := movers[1].Children[0].Attributes.Get("mathvariant"); value != "bold-italic" {
		t.Fatalf("starred vector base mathvariant = %#v, want bold-italic", value)
	}
}
