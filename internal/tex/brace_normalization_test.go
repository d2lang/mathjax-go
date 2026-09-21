package tex

import (
	"fmt"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// ParseUtil.underOver's compiled AST in both modes from pinned MathJax 3.2.2
// oracle contains an empty, zero-child mo before an embellished sum. A token
// with one empty text child is structurally different even if its text is "".
func TestBraceEmbellishedSumNormalization(t *testing.T) {
	for _, display := range []bool{true, false} {
		t.Run(fmt.Sprintf("display=%t", display), func(t *testing.T) {
			root, err := NewCompiler().Compile(`\overbrace{\sum_{i}^{n}}^{N}`, display)
			if err != nil {
				t.Fatal(err)
			}
			if errors := root.Find("merror"); len(errors) != 0 {
				t.Fatalf("unexpected error: %s", textContent(errors[0]))
			}
			outer := root.Children[0].Children[0]
			if outer.Kind != "mover" || len(outer.Children) != 2 || textContent(outer.Children[1]) != "N" {
				t.Fatalf("brace label is not an outer stacked over-script: %#v", outer)
			}
			atom := outer.Children[0]
			if atom.Kind != "TeXAtom" || atom.TeXClass != mml.TeXClassOp {
				t.Fatalf("brace is not an OP atom: %#v", atom)
			}
			for _, property := range []string{"movesupsub", "subsupOK"} {
				if value, ok := atom.Property(property); !ok || value != true {
					t.Fatalf("brace %s = %v, %v; want true", property, value, ok)
				}
			}
			brace := atom.Children[0].Children[0]
			if brace.Kind != "mover" || len(brace.Children) != 2 || textContent(brace.Children[1]) != "⏞" {
				t.Fatalf("missing inner brace layer: %#v", brace)
			}
			base := brace.Children[0]
			if base.Kind != "mrow" || len(base.Children) != 2 || base.Children[1].Kind != "munderover" {
				t.Fatalf("missing normalization row: %#v", base)
			}
			empty := base.Children[0]
			if empty.Kind != "mo" || len(empty.Children) != 0 {
				t.Fatalf("normalization operator = %s with %d children; want zero-child mo", empty.Kind, len(empty.Children))
			}
			if value, ok := empty.Attributes.GetExplicit("rspace"); !ok || value != 0 {
				t.Fatalf("empty operator rspace = %v, %v", value, ok)
			}
			core := base.Children[1].Children[0]
			if core.Kind != "mo" || textContent(core) != "∑" {
				t.Fatalf("normalized core is not the original sum: %#v", core)
			}
			for _, name := range []string{"lspace", "rspace"} {
				if value, ok := core.Attributes.GetExplicit(name); !ok || value != 0 {
					t.Fatalf("sum %s = %v, %v; want explicit zero", name, value, ok)
				}
			}
		})
	}
}

func TestInlineBraceScriptNormalizationBoundary(t *testing.T) {
	for _, test := range []struct {
		name, source, kind string
		normalized         bool
	}{
		{"subscript", `\overbrace{\sum_i}^{N}`, "munder", true},
		{"superscript", `\overbrace{\sum^n}^{N}`, "mover", true},
		{"explicit-nolimits", `\overbrace{\sum\nolimits_i^n}^{N}`, "msubsup", false},
		{"ordinary-script", `\overbrace{x_i^n}^{N}`, "msubsup", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			root, err := NewCompiler().Compile(test.source, false)
			if err != nil {
				t.Fatal(err)
			}
			if errors := root.Find("merror"); len(errors) != 0 {
				t.Fatalf("unexpected error: %s", textContent(errors[0]))
			}
			outer := root.Children[0].Children[0]
			if outer.Kind != "mover" || len(outer.Children) != 2 || textContent(outer.Children[1]) != "N" {
				t.Fatalf("missing outer brace label: %#v", outer)
			}
			brace := outer.Children[0].Children[0].Children[0]
			if brace.Kind != "mover" || len(brace.Children) != 2 || textContent(brace.Children[1]) != "⏞" {
				t.Fatalf("missing inner brace: %#v", brace)
			}
			base := brace.Children[0]
			if test.normalized {
				if base.Kind != "mrow" || len(base.Children) != 2 {
					t.Fatalf("missing normalization row: %#v", base)
				}
				empty := base.Children[0]
				if empty.Kind != "mo" || len(empty.Children) != 0 {
					t.Fatalf("normalization operator must have zero children: %#v", empty)
				}
				base = base.Children[1]
				core := base.Children[0]
				if core.Kind != "mo" || textContent(core) != "∑" {
					t.Fatalf("missing core sum: %#v", core)
				}
				for _, name := range []string{"lspace", "rspace"} {
					if value, ok := core.Attributes.GetExplicit(name); !ok || value != 0 {
						t.Fatalf("sum %s = %v, %v; want explicit zero", name, value, ok)
					}
				}
			}
			if base.Kind != test.kind {
				t.Fatalf("script kind = %s, want %s", base.Kind, test.kind)
			}
		})
	}
}
