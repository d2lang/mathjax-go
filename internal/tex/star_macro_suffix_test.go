// Copyright 2017-2022 MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
)

func TestStarMacroCallerSuffixReference(t *testing.T) {
	var reference struct {
		MathjaxGitCommit, SVG, FinalSource string
		FinalCursor, FinalMacroCount       int
	}
	data, err := os.ReadFile("../../testdata/star_macro_suffix_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &reference); err != nil {
		t.Fatal(err)
	}
	if reference.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(reference.SVG) != 937 {
		t.Fatal("unbound original StarMacro suffix reference")
	}
	state := newParseState()
	// This is the saved private registered-Macro witness, not public D2 setup.
	state.macros["vec"] = macroDefinition{body: "#2", arguments: 2}
	p := &parser{source: `\va{q}{z}`, state: state, display: true}
	root, parseError, err := starMacroFixtureCompile(p)
	if err != nil || parseError != nil {
		t.Fatalf("compile = %v, %v; want original success", err, parseError)
	}
	if p.state != state || p.source != reference.FinalSource || p.pos != reference.FinalCursor || state.macroCount != reference.FinalMacroCount {
		t.Fatalf("caller program = %q, cursor %d, count %d; want %q, %d, %d", p.source, p.pos, state.macroCount, reference.FinalSource, reference.FinalCursor, reference.FinalMacroCount)
	}
	got, err := svg.NewTypesetter().Typeset(root, pipeline.DefaultOptions())
	if err != nil || got != reference.SVG {
		t.Fatalf("original SVG differs: %v\ngot: %s\nwant: %s", err, got, reference.SVG)
	}
}

func TestStarMacroInstallsCallerProgramBeforeCharge(t *testing.T) {
	for _, name := range []string{"va", "vectorarrow", "vu", "vectorunit"} {
		for _, star := range []string{"", "*"} {
			t.Run(name+star, func(t *testing.T) {
				state := newParseState()
				state.macroCount = maxMacros
				p := &parser{source: star + `{q}+r`, state: state}
				_, err := p.vectorAccent(name)
				if err == nil || !strings.Contains(err.Error(), "maximum macro substitution") {
					t.Fatalf("error = %v", err)
				}
				accent := "vec"
				if name == "vu" || name == "vectorunit" {
					accent = "hat"
				}
				want := `\` + accent + `{\vb` + star + `{q}}+r`
				if p.source != want || p.pos != 0 || state.macroCount != maxMacros+1 {
					t.Fatalf("program %q, cursor %d, count %d; want %q, 0, %d", p.source, p.pos, state.macroCount, want, maxMacros+1)
				}
				if p.vectorAlias || p.starMacroChildren {
					t.Fatal("StarMacro leaked temporary child context to caller")
				}
			})
		}
	}
}

func TestStarMacroPublicContinuationControls(t *testing.T) {
	// These are metamorphic continuation guards; exact primary SVG coverage
	// remains in the retained fourteen-case StarMacro fixture and suffix case.
	for _, pair := range [][2]string{
		{`\va{q}+r`, `\vec{\vb{q}}+r`},
		{`\vu{q}+r`, `\hat{\vb{q}}+r`},
		{`\vectorarrow*{q}+r`, `\vec{\vb*{q}}+r`},
		{`\vectorunit*{q}+r`, `\hat{\vb*{q}}+r`},
		{`x^{\va{q}}+r`, `x^{\vec{\vb{q}}}+r`},
		{`\mathrm{\va{abc}}+r`, `\mathrm{\vec{\vb{abc}}}+r`},
	} {
		t.Run(pair[0], func(t *testing.T) {
			var rendered [2]string
			for i, source := range pair {
				p := &parser{source: source, state: newParseState(), display: true}
				root, parseError, err := starMacroFixtureCompile(p)
				if err != nil || parseError != nil {
					t.Fatalf("%q: %v, %v", source, err, parseError)
				}
				rendered[i], err = svg.NewTypesetter().Typeset(root, pipeline.DefaultOptions())
				if err != nil {
					t.Fatal(err)
				}
				if p.vectorAlias || p.starMacroChildren {
					t.Fatal("temporary child context leaked to caller")
				}
			}
			if rendered[0] != rendered[1] {
				t.Fatalf("alias differs from its explicit caller continuation\n%s\n%s", rendered[0], rendered[1])
			}
		})
	}
}

func TestStarMacroRealChildBoundaries(t *testing.T) {
	for _, name := range []string{"vec", "hat", "vb"} {
		t.Run(name, func(t *testing.T) {
			state := newParseState()
			state.macroCount = maxMacros
			state.macros["probe"] = macroDefinition{body: "x"}
			p := &parser{source: `{\probe}`, state: state}
			if _, err := p.command(name); err != nil {
				t.Fatalf("real child inherited caller's exhausted budget: %v", err)
			}
			if p.state != state || state.macroCount != maxMacros || p.vectorAlias || p.starMacroChildren {
				t.Fatal("real child did not restore caller state and context")
			}
		})
	}
	// A registered accent can forward the generated vector argument. The
	// VectorBold child remains independent after both same-caller charges.
	t.Run("registered-forwarding", func(t *testing.T) {
		state := newParseState()
		state.macroCount = maxMacros - 2
		state.macros["vec"] = macroDefinition{body: "#1", arguments: 1}
		state.macros["probe"] = macroDefinition{body: "x"}
		p := &parser{source: `\va{\probe}+r`, state: state}
		if _, _, err := p.parseRow(0, false); err != nil {
			t.Fatal(err)
		}
		if p.state != state || state.macroCount != maxMacros || p.vectorAlias || p.starMacroChildren {
			t.Fatal("registered forwarding changed caller budget or context")
		}
	})
	t.Run("missing-suffix", func(t *testing.T) {
		state := newParseState()
		state.macros["vec"] = macroDefinition{body: "#2", arguments: 2}
		p := &parser{source: `\va{q}`, state: state}
		_, _, err := p.parseRow(0, false)
		if err == nil || err.Error() != `Missing argument for \vec` || state.macroCount != 1 {
			t.Fatalf("missing suffix = %v, count %d; want original Macro argument error before its own charge", err, state.macroCount)
		}
	})
}
