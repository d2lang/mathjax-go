package tex

import (
	"errors"
	"testing"
)

func TestRelaxMacroPrecedence(t *testing.T) {
	for _, tc := range []struct {
		name, body, source string
		undefined          bool
	}{
		{"named-relax", "x", `\relax^n`, false},
		{"nested-named-relax", "x", `{\relax}_i`, false},
		{"expansion-relax", `\relax`, `\probe`, true},
		{"argument-expansion", `#1`, `\probe{\relax}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newParseState()
			name := "relax"
			if tc.undefined {
				name = "probe"
			}
			s.macros[name] = macroDefinition{body: tc.body, arguments: macroArguments(tc.body)}
			p := &parser{source: tc.source, state: s}
			nodes, stop, err := p.parseRow(0, false)
			if tc.undefined {
				var e *Error
				if !errors.As(err, &e) || e.ID != "UndefinedControlSequence" || e.Message != `Undefined control sequence \relax` {
					t.Fatalf("error=%v", err)
				}
				return
			}
			if err != nil || stop != "" || len(nodes) != 1 {
				t.Fatalf("defined macro failed: %v %s", err, stop)
			}
			if nodes[0].Kind != "msup" && nodes[0].Kind != "msub" {
				t.Fatal("defined macro no longer participates in scripts", nodes[0].Kind)
			}
			n := nodes[0]
			for len(n.Children) > 0 {
				n = n.Children[0]
			}
			if n.Kind != "text" || n.Text != "x" {
				t.Fatal("defined macro did not retain replacement")
			}
		})
	}
}
