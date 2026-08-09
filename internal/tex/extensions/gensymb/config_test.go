// Copyright (c) 2021-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package gensymb

import (
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/tex/extensions/spec"
)

func TestConfiguration(t *testing.T) {
	attrs := []spec.Attribute{
		{Name: "mathvariant", Value: "normal"},
		{Name: "class", Value: "MathML-Unit"},
	}
	want := []spec.Symbol{
		{Name: "ohm", Character: "Ω", TokenKind: "mi", Attributes: attrs},
		{Name: "degree", Character: "°", TokenKind: "mi", Attributes: attrs},
		{Name: "celsius", Character: "℃", TokenKind: "mi", Attributes: attrs},
		{Name: "perthousand", Character: "‰", TokenKind: "mi", Attributes: attrs},
		{Name: "micro", Character: "µ", TokenKind: "mi", Attributes: attrs},
	}
	if Configuration.Name != "gensymb" || Configuration.MacroMap != "gensymb-symbols" {
		t.Fatalf("Configuration identity = %#v", Configuration)
	}
	if !reflect.DeepEqual(Configuration.Symbols, want) {
		t.Fatalf("Configuration.Symbols = %#v, want %#v", Configuration.Symbols, want)
	}
}
