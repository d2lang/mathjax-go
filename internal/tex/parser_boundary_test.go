// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
//
// Frozen source shapes are from D2's embedded MathJax 3.2.2 component.

package tex

import "testing"

func TestNamedFunctionGroupBoundaryFrozenMML(t *testing.T) {
	tests := []struct {
		name, source, want string
	}{
		{
			"function at group boundary",
			`\textcolor{red}{y} = \textcolor{green}{\sin} x`,
			`math([mstyle([mi(y)]),mo(=),mstyle([mi(sin)]),mi(x)])`,
		},
		{
			"function followed by argument",
			`\sin x`,
			`math([mi(sin),mo(⁡),mi(x)])`,
		},
		{
			"function followed by binary operator",
			`\sin + x`,
			`math([mi(sin),mo(+),mi(x)])`,
		},
		{
			"function followed by explicit spacer",
			`\sin\,x`,
			`math([mi(sin),mstyle([mspace()]),mi(x)])`,
		},
		{
			"scripted function followed by argument",
			`\sin^2 x`,
			`math([msup(mi(sin),mn(2)),mo(⁡),mi(x)])`,
		},
		{
			"function finalized by ordinary group",
			`{\sin} x`,
			`math([TeXAtom([mi(sin)]),mi(x)])`,
		},
	}
	compiler := NewCompiler()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root, err := compiler.Compile(test.source, true)
			if err != nil {
				t.Fatal(err)
			}
			if got := mmlString(root); got != test.want {
				t.Fatalf("MML mismatch:\n got %s\nwant %s", got, test.want)
			}
		})
	}
}

func TestMhchemMathFontMultiLetterIdentifiersFrozenMML(t *testing.T) {
	const source = `\ce{SO4^2- + Ba^2+ -> BaSO4 v}`
	const want = `math([TeXAtom([TeXAtom([mi(SO)]),msub(TeXAtom([TeXAtom([mpadded([mphantom([mi(A)])])])]),TeXAtom([TeXAtom([mpadded([mn(4)])])])),msup(TeXAtom([TeXAtom([mpadded([mphantom([mi(A)])])])]),TeXAtom([mn(2),mo(−)])),TeXAtom([]),mo(+),TeXAtom([]),TeXAtom([mi(Ba)]),msup(TeXAtom([TeXAtom([mpadded([mphantom([mi(A)])])])]),TeXAtom([mn(2),mo(+)])),TeXAtom([]),TeXAtom([mo(⟶)]),TeXAtom([]),TeXAtom([mi(BaSO)]),msub(TeXAtom([TeXAtom([mpadded([mphantom([mi(A)])])])]),TeXAtom([TeXAtom([mpadded([mn(4)])])])),mo(↓),TeXAtom([])])])`
	root, err := NewCompiler().Compile(source, true)
	if err != nil {
		t.Fatal(err)
	}
	if got := mmlString(root); got != want {
		t.Fatalf("MML mismatch:\n got %s\nwant %s", got, want)
	}
}
