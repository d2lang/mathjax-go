// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"encoding/json"
	"os"
	"testing"
)

// The original primary and historical boundary fields remain untouched in the
// fixture. D103 now requires all fourteen complete primary outputs, with no
// property-array or accepted-output qualification.
func TestMacroBoundaryOperatorNameReferences(t *testing.T) {
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			macroReferenceCase
			// Preserve the historical shared-counter expectation in the fixture.
			MacroCount int
		}
	}
	data, err := os.ReadFile("testdata/macro_operatorname_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 14 {
		t.Fatal("unbound operator-name records")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			state := newParseState()
			for _, r := range c.Registrations {
				state.macros[r.Name] = macroReferenceDefinition(r)
			}
			p := &parser{source: c.TeX, state: state, display: c.Display}
			children, stop, err := p.parseRow(0, false)
			if err != nil || stop != "" {
				t.Fatalf("parse error %v, stop %q", err, stop)
			}
			// Original OperatorName uses a genuine child with its own macro count.
			if p.source != c.TeX || p.pos != len(c.TeX) || p.state != state || state.macroCount != 0 {
				t.Fatal("child expansion changed caller source/cursor/state or revisited old input")
			}
			operatorNameReferenceOutput(t, operatorNameFinalize(children, c.Display), c.Display, c.Primary)
		})
	}
}

func TestMacroBoundaryOperatorNameErrorOwnership(t *testing.T) {
	state := newParseState()
	state.macros["pick"] = macroDefinition{body: "#1", arguments: 1}
	source := `{a\pick} remaining`
	p := &parser{source: source, state: state}
	nodes, err := p.amsOperatorName("operatorname")
	macroReferenceError(t, err, &Error{ID: "MissingArgFor", Message: "Missing argument for \\pick"})
	if nodes != nil || p.source != source || p.pos != len(`{a\pick}`) || p.state != state || state.macroCount != 0 {
		t.Fatal("error path transferred child program or changed caller state")
	}
}
