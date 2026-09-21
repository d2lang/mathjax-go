package tex

import (
	"github.com/d2lang/mathjax-go/internal/mml"
	"testing"
)

func TestVectorAccentStretchPolicy(t *testing.T) {
	for _, c := range []struct {
		source, mark string
		stretchy     bool
	}{
		{`\vec{x}`, "→", false}, {`\vec{abcdef}`, "→", false}, {`\vb{\vec{y}}`, "→", false},
		{`\va*{x}`, "→", false}, {`\widehat{abcdef}`, "^", true}, {`\widetilde{abcdef}`, "~", true},
	} {
		t.Run(c.source, func(t *testing.T) {
			root, err := NewCompiler().Compile(c.source, true)
			if err != nil {
				t.Fatal(err)
			}
			found := 0
			root.Walk(func(n *mml.Node) bool {
				if n.Kind == "mo" && textContent(n) == c.mark {
					found++
					if value, _ := n.Attributes.Get("stretchy"); value != c.stretchy {
						t.Errorf("stretchy %v, want %v", value, c.stretchy)
					}
				}
				return true
			})
			if found != 1 {
				t.Fatalf("expected exactly one accent, got %d", found)
			}
		})
	}
}
