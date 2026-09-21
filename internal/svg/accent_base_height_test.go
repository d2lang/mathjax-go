package svg

import (
	"github.com/d2lang/mathjax-go/internal/layout"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/tex"
	"testing"
)

func TestAccentBaseHeightPersistsThroughRender(t *testing.T) {
	for _, c := range []struct {
		name, source string
		want         float64
	}{
		{"low accent", `\vec{a}`, .442},
		{"high accent", `\vec{A}`, .716},
		{"unaccented stack", `\overset{b}{a}`, .441},
	} {
		t.Run(c.name, func(t *testing.T) {
			root, err := tex.NewCompiler().Compile(c.source, true)
			if err != nil {
				t.Fatal(err)
			}
			options := pipeline.DefaultOptions()
			r := &renderer{options: options, params: layout.TeXParameters, pxPerEm: options.Ex / layout.TeXParameters.XHeight}
			prepareTeXClasses(root)
			outer := r.wrap(root, nil, 0, true)
			var target *wrapper
			var walk func(*wrapper)
			walk = func(w *wrapper) {
				if w.node.Kind == "mover" {
					if target != nil {
						t.Fatal("multiple stacks")
					}
					target = w
				}
				for _, child := range w.children {
					walk(child)
				}
			}
			walk(outer)
			if target == nil {
				t.Fatal("missing stack")
			}
			base := target.scriptBase().outerBBox()
			before := base.H
			outer.outerBBox()
			if target.scriptBase().outerBBox() != base {
				t.Fatal("base cache identity changed")
			}
			if base.H != c.want {
				t.Fatalf("cached base height %g, want %g (before %g)", base.H, c.want, before)
			}
			for i := 0; i < 2; i++ {
				g := NewElement("g")
				outer.toSVG(g)
				if base.H != c.want || target.scriptBase().outerBBox() != base {
					t.Fatal("cached height/identity lost during rendering")
				}
			}
		})
	}
}
