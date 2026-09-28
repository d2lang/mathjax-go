package tex

import (
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// Bound to frozen MathtoolsUtil.spreadLines method observations: absence of
// the dimension fails only after selecting an mtable with truthy rowspacing.
func TestSpreadPopSelectionAndMissingDimension(t *testing.T) {
	for _, tc := range []struct{ name string; wrap func(*mml.Node) *mml.Node; selected bool }{
		{"direct", func(n *mml.Node) *mml.Node { return n }, true},
		{"inferred", func(n *mml.Node) *mml.Node { return forcedRow([]*mml.Node{n, token("mi", "x")}, true) }, true},
		{"explicit", func(n *mml.Node) *mml.Node { return forcedRow([]*mml.Node{n}, false) }, false},
		{"script", func(n *mml.Node) *mml.Node { return node("msup", n, token("mn", "1")) }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			table := node("mtable")
			if err := mathtoolsSpreadPop(tc.wrap(table), nil); (err != nil) != tc.selected {
				t.Fatalf("selected %v, error %v", tc.selected, err)
			}
		})
	}
	for _, spacing := range []any{nil, "", false, 0, float64(0)} {
		table := node("mtable")
		table.Attributes.Set("rowspacing", spacing)
		if err := mathtoolsSpreadPop(table, nil); err != nil { t.Fatal(err) }
		if value, explicit := table.Attributes.GetExplicit("rowspacing"); !explicit || value != nil {
			t.Fatalf("falsy assignment lost undefined value: %v %v", value, explicit)
		}
	}
	for _, spacing := range []any{"0", " ", "1ex"} {
		table := node("mtable")
		table.Attributes.Set("rowspacing", spacing)
		if err := mathtoolsSpreadPop(table, nil); err == nil { t.Fatalf("truthy %q accepted", spacing) }
	}
}

func TestSpreadPopDefinedFalsySpacing(t *testing.T) {
	table := node("mtable")
	table.Attributes.Set("rowspacing", "")
	spread := "1pt"
	if err := mathtoolsSpreadPop(table, &spread); err != nil { t.Fatal(err) }
	if value, _ := table.Attributes.Get("rowspacing"); value != spread { t.Fatalf("got %v", value) }
}
