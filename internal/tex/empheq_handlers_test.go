// Copyright (c) 2021-2022 MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package tex

import (
	"strings"
	"testing"
)

func TestEmpheqMOCommands(t *testing.T) {
	for name, character := range empheqMO {
		p := &parser{state: newParseState()}
		nodes, handled, err := p.empheqCommand(name)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if !handled || len(nodes) != 1 || nodes[0].Kind != "mo" || textContent(nodes[0]) != character {
			t.Errorf("%s = handled %v, nodes %#v", name, handled, nodes)
		}
	}
	p := &parser{state: newParseState()}
	if nodes, handled, err := p.empheqCommand("not-empheq"); handled || nodes != nil || err != nil {
		t.Fatalf("unknown command = %#v, %v, %v", nodes, handled, err)
	}
}

func TestEmpheqDelimiterCommands(t *testing.T) {
	for _, name := range []string{"empheql", "empheqr", "empheqbigl", "empheqbigr"} {
		p := &parser{source: `\langle`, state: newParseState()}
		nodes, handled, err := p.empheqCommand(name)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if !handled || len(nodes) != 1 || textContent(nodes[0]) != "⟨" {
			t.Fatalf("%s = handled %v, nodes %#v", name, handled, nodes)
		}
		for _, attribute := range []string{"stretchy", "symmetric"} {
			if value, _ := nodes[0].Attributes.GetExplicit(attribute); value != true {
				t.Errorf("%s %s = %#v, want true", name, attribute, value)
			}
		}
	}

	p := &parser{source: `\notadelimiter`, state: newParseState()}
	_, handled, err := p.empheqCommand("empheql")
	if !handled || err == nil {
		t.Fatalf("invalid delimiter = handled %v, err %v", handled, err)
	}
	texErr, ok := err.(*Error)
	if !ok || texErr.ID != "MissingOrUnrecognizedDelim" || texErr.Message != `Missing or unrecognized delimiter for \empheql` {
		t.Fatalf("invalid delimiter error = %#v", err)
	}
}

func TestEmpheqSplitOptions(t *testing.T) {
	options, err := empheqSplitOptions(` left = {(} , right = {\empheqrbrace} `)
	if err != nil {
		t.Fatal(err)
	}
	if options["left"] != "(" || options["right"] != `\empheqrbrace` {
		t.Fatalf("options = %#v", options)
	}
	options, err = empheqSplitOptions(`left={a b}`)
	if err != nil || options["left"] != "a b" {
		t.Fatalf("spaced option = %#v, %v", options, err)
	}
	options, err = empheqSplitOptions(`left=false,right`)
	if err != nil || options["left"] != false || options["right"] != true {
		t.Fatalf("boolean options = %#v, %v", options, err)
	}

	_, err = empheqSplitOptions(`box=\boxed`)
	assertEmpheqError(t, err, "InvalidOption", "Invalid option: box")
	_, err = empheqSplitOptions(`left={{(}`)
	assertEmpheqError(t, err, "ExtraOpenMissingClose", "Extra open brace or missing close brace")
	_, err = empheqSplitOptions(`box=x,left={{(}`)
	assertEmpheqError(t, err, "ExtraOpenMissingClose", "Extra open brace or missing close brace")
}

func TestEmpheqEnvironmentFrozenMML(t *testing.T) {
	tests := []struct {
		name, source, want string
	}{
		{
			"align",
			`{align}a&=b\end{empheq}`,
			`math([mtable(mtr(mtd([mi(a)]),mtd([mi(),mo(=),mi(b)])))])`,
		},
		{
			"equation",
			`{equation}a=b\end{empheq}`,
			`math([mi(a),mo(=),mi(b)])`,
		},
		{
			"equation-left-source-quirk",
			`[left={(}]{equation}a=b\end{empheq}`,
			`math([mi(mtd([mpadded([mpadded([mo((),mphantom([mpadded([mi(a),mo(=),mi(b)])])]),mphantom([mpadded([mi(a)])])])]),a),mo(mtd([]),=),mi(mtd([]),b)])`,
		},
		{
			"left-right",
			`[left={(},right={)}]{align}a&=b\end{empheq}`,
			`math([mtable(mtr(mtd([mpadded([mpadded([mo((),mphantom([mpadded([mtable(mtr(mtd([mi(a)]),mtd([mi(),mo(=),mi(b)])))])])]),mphantom([mpadded([mtable(mtr(mtd([mi(a)]),mtd([mi(),mo(=),mi(b)])))])])])]),mtd([mi(a)]),mtd([mi(),mo(=),mi(b)]),mtd([mpadded([mpadded([mo()),mphantom([mpadded([mtable(mtr(mtd([mi(a)]),mtd([mi(),mo(=),mi(b)])))])])]),mphantom([mpadded([mtable(mtr(mtd([mi(a)]),mtd([mi(),mo(=),mi(b)])))])])])])))])`,
		},
		{
			"gather-left-multirow",
			`[left={(}]{gather}a\\b\end{empheq}`,
			`math([mtable(mtr(mtd([mpadded([mpadded([mo((),mphantom([mpadded([mtable(mtr(mtd([mi(a)])),mtr(mtd([mi(b)])))])])]),mphantom([mpadded([mtable(mtr(mtd([mi(a)])))])])])]),mtd([mi(a)])),mtr(mtd([]),mtd([mi(b)])))])`,
		},
		{
			"alignat",
			`{alignat=2}a&=b&c&=d\end{empheq}`,
			`math([mtable(mtr(mtd([mi(a)]),mtd([mi(),mo(=),mi(b)]),mtd([mi(c)]),mtd([mi(),mo(=),mi(d)])))])`,
		},
		{
			"bare-left",
			`[left]{align}a&=b\end{empheq}`,
			`math([mtable(mtr(mtd([mpadded([mpadded([mphantom([mpadded([mtable(mtr(mtd([mi(a)]),mtd([mi(),mo(=),mi(b)])))])])]),mphantom([mpadded([mtable(mtr(mtd([mi(a)]),mtd([mi(),mo(=),mi(b)])))])])])]),mtd([mi(a)]),mtd([mi(),mo(=),mi(b)])))])`,
		},
		{
			"flalign-star",
			`{flalign*}a&=b\end{empheq}`,
			`math([mtable(mtr(mtd([mi(a)]),mtd([mi(),mo(=),mi(b)])))])`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := &parser{source: test.source, state: newParseState(), display: true}
			nodes, handled, err := p.empheqEnvironment("empheq")
			if err != nil {
				t.Fatal(err)
			}
			if !handled || p.pos != len(p.source) {
				t.Fatalf("handled = %v, position = %d/%d", handled, p.pos, len(p.source))
			}
			root := node("math", nodes...)
			if got := mmlString(root); got != test.want {
				t.Fatalf("MML mismatch:\n got %s\nwant %s", got, test.want)
			}
		})
	}
}

func TestEmpheqAllowedEnvironments(t *testing.T) {
	for _, name := range []string{
		"equation", "equation*", "align", "align*", "gather", "gather*",
		"flalign", "flalign*", "alignat", "alignat*", "multline", "multline*",
	} {
		if !empheqAllowedEnvironment(name) {
			t.Errorf("%q should be allowed", name)
		}
	}
	for _, name := range []string{"", "matrix", "aligned", "align**", " equation"} {
		if empheqAllowedEnvironment(name) {
			t.Errorf("%q should not be allowed", name)
		}
	}
}

func TestEmpheqEnvironmentErrors(t *testing.T) {
	tests := []struct {
		name, source, id, message string
	}{
		{"unknown", `{matrix}a\end{empheq}`, "UnknownEnv", `Unknown environment "matrix"`},
		{"untrimmed", `{ align}a\end{empheq}`, "UnknownEnv", `Unknown environment " align"`},
		{"option", `[box=\boxed]{align}a&=b\end{empheq}`, "InvalidOption", "Invalid option: box"},
		{"missing-end", `{align}a&=b`, "MissingEnd", `Missing \end{empheq}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := &parser{source: test.source, state: newParseState(), display: true}
			_, handled, err := p.empheqEnvironment("empheq")
			if !handled {
				t.Fatal("empheq environment was not handled")
			}
			assertEmpheqError(t, err, test.id, test.message)
		})
	}
}

func TestEmpheqTableHelpersWithLabels(t *testing.T) {
	label := node("mtd", token("mtext", "1"))
	row := node("mlabeledtr", label, node("mtd", token("mi", "a")))
	table := node("mtable", row)
	table.Attributes.Set("columnalign", "right")
	table.Attributes.Set("columnspacing", "0em")
	original := table.Clone()
	p := &parser{state: newParseState(), display: true}
	if err := p.empheqAddLeft(table, original, "("); err != nil {
		t.Fatal(err)
	}
	if row.Children[0] != label || row.Children[1].Kind != "mtd" || textContent(row.Children[2]) != "a" {
		t.Fatalf("labeled row order = %s", mmlString(row))
	}
}

func assertEmpheqError(t *testing.T, err error, id, message string) {
	t.Helper()
	texErr, ok := err.(*Error)
	if !ok || texErr.ID != id || texErr.Message != message {
		t.Fatalf("error = %#v, want %s %q", err, id, message)
	}
}

func TestEmpheqSourceHandlerInventory(t *testing.T) {
	wantMO := 32
	if len(empheqMO) != wantMO {
		t.Fatalf("EmpheqMO command count = %d, want %d", len(empheqMO), wantMO)
	}
	if len(empheqDelimiters) != 4 {
		t.Fatalf("EmpheqDelim command count = %d, want 4", len(empheqDelimiters))
	}
	for name := range empheqMO {
		if !strings.HasPrefix(name, "empheq") {
			t.Errorf("unexpected command %q", name)
		}
	}
}
