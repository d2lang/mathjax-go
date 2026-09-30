// SPDX-License-Identifier: Apache-2.0
package tex

import "testing"

// Scalar and ordering expectations come from the frozen public original's
// passive requests 38, 39, 42 and 43, independently qualified in
// original44-saved-observation-peer185.json (SHA256 ab5d09cc...).
// NodeStack.Last is the oldest node: StackItem.ts:149-150. This tests the
// physical parse entrypoint, without seeded counts or fabricated End items.
func TestNumCasesDirectPopPhysicalContinuation(t *testing.T) {
	for _, route := range []struct {
		name   string
		cursor int
	}{
		{"numcases", 109},
		{"subnumcases", 115},
	} {
		for _, display := range []bool{false, true} {
			mode := "inline"
			if display {
				mode = "display"
			}
			t.Run(route.name+"-"+mode, func(t *testing.T) {
				source := `\DeclareMathOperator{\empheqlbrace}{L}\begin{` + route.name + `}{F=}x\end{spreadlines}\matrix{1&2\\3&4}\end{` + route.name + `}+Z`
				state := newParseState()
				p := &parser{source: source, state: state, display: display}
				nodes, stop, err := p.parseRow(0, false)
				if err != nil || stop != "" {
					t.Fatalf("physical parse: stop=%q error=%v", stop, err)
				}
				// These source strings are ASCII; byte and UTF-16 cursors agree.
				remaining := "<out-of-bounds>"
				if p.pos >= 0 && p.pos <= len(p.source) {
					remaining = p.source[p.pos:]
				}
				if p.pos != route.cursor || p.pos != len(p.source) || remaining != "" || state.macroCount != 3 {
					t.Fatalf("final physical input: cursor=%d remaining=%q macroCount=%d", p.pos, remaining, state.macroCount)
				}
				if len(nodes) != 4 || nodes[0].Kind != "mi" || nodes[1].Kind != "mtable" || nodes[2].Kind != "mo" || nodes[3].Kind != "mi" {
					t.Fatalf("physical node order differs: %#v", nodes)
				}
				// The original text node receives an mtd child during left
				// decoration. Assert its text and explicit attributes, without
				// imposing a complete child-tree golden on that representation.
				for _, entry := range []struct {
					index int
					text  string
				}{{0, "x"}, {2, "+"}, {3, "Z"}} {
					if len(nodes[entry.index].Children) == 0 || nodes[entry.index].Children[0].Text != entry.text {
						t.Fatalf("authored node %d text differs", entry.index)
					}
				}
				for attribute, want := range map[string]string{"columnalign": "right ", "columnspacing": "0em "} {
					got, explicit := nodes[0].Attributes.GetExplicit(attribute)
					if !explicit || got != want {
						t.Fatalf("oldest mi %s = %#v; want %q", attribute, got, want)
					}
				}
				// The later matrix keeps its captured own table metadata.
				for attribute, want := range map[string]string{"rowspacing": "4pt", "columnspacing": "1em"} {
					got, explicit := nodes[1].Attributes.GetExplicit(attribute)
					if !explicit || got != want {
						t.Fatalf("later matrix %s = %#v; want %q", attribute, got, want)
					}
				}
				tags := state.amsTags
				if tags == nil || tags.current == nil || tags.current.environment != route.name || len(tags.stack) != 1 {
					t.Fatalf("direct Pop lost active Cases tags: %#v", tags)
				}
			})
		}
	}
}
