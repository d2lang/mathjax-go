// Copyright (c) 2018-2022 MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2 tests.
//
// Source: ts/input/tex/cases/CasesConfiguration.ts.

package tex

import "testing"

func TestCasesNumCasesFrozenShape(t *testing.T) {
	p := &parser{
		source:  `{f(x)=}x&if x>0\\0&otherwise\end{numcases}`,
		state:   newParseState(),
		display: true,
	}
	p.state.augmentedPackages = true
	nodes, handled, err := p.casesEnvironment("numcases")
	if err != nil {
		t.Fatal(err)
	}
	if !handled {
		t.Fatal("numcases was not handled")
	}
	root := node("math", nodes...)
	want := `math([mtable(mtr(mtd([mpadded([mpadded([mi(f),mo((),mi(x),mo()),mo(=),mo({),mstyle([mspace()]),mphantom([mpadded([mtable(mtr(mtd([mi(x)]),mtd([mstyle([mtext(if x>0)])])),mtr(mtd([mn(0)]),mtd([mstyle([mtext(otherwise)])])))])])]),mphantom([mpadded([mtable(mtr(mtd([mi(x)]),mtd([mstyle([mtext(if x>0)])])))])])])]),mtd([mi(x)]),mtd([mstyle([mtext(if x>0)])])),mtr(mtd([]),mtd([mn(0)]),mtd([mstyle([mtext(otherwise)])])))])`
	if got := mmlString(root); got != want {
		t.Fatalf("MML mismatch:\n got %s\nwant %s", got, want)
	}
}

func TestCasesNumCasesExtraEntry(t *testing.T) {
	p := &parser{
		source:  `{f(x)=}x&text&extra\end{numcases}`,
		state:   newParseState(),
		display: true,
	}
	_, handled, err := p.casesEnvironment("numcases")
	if !handled {
		t.Fatal("numcases was not handled")
	}
	texErr, ok := err.(*Error)
	if !ok || texErr.ID != "ExtraCasesAlignTab" {
		t.Fatalf("error = %#v, want ExtraCasesAlignTab", err)
	}
}
