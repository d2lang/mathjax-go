package svg

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/d2lang/mathjax-go/internal/font"
	"github.com/d2lang/mathjax-go/internal/layout"
	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
)

func TestMathAccentScriptConstructorReference(t *testing.T) {
	b, err := os.ReadFile("testdata/accent_script_constructor_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Source       string
		AssetsSHA256 map[string]string
		Cases        []struct {
			Name         string
			Display      bool
			RenderError  *string
			FullTree     *msTree
			Observations []struct {
				Path, BaseCorePath                    []int
				Kind                                  string
				IsMathAccent, AccentOver, AccentUnder bool
			}
		}
	}
	if err = json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Cases) != 88 || f.Source != "Unmodified pinned CommonScriptbase constructor, registered MML inputs, observer only" || len(f.AssetsSHA256) != 3 ||
		f.AssetsSHA256["mathjax.js"] != "cbbc1051a1f8abb1a181b6aa0fe927c020e3631ca630d19f52d9abb65b5ee869" ||
		f.AssetsSHA256["polyfills.js"] != "7fe1d048c78b0e09854c1259f7413868a51cc8f0c822489eb1ac36ae6c85ce01" ||
		f.AssetsSHA256["setup.js"] != "a52cb0bbabfbd7796b9fedd793b9386c474e3123e788ee151cbe47ca1fa6e881" {
		t.Fatalf("cases %d", len(f.Cases))
	}
	expectedErrors := map[string]bool{"script-selected-inline": true, "script-selected-display": true}
	for _, label := range []string{"true", "false", "null", "empty", "whitespace", "hex", "padded", "number"} {
		for _, child := range []string{"first", "second"} {
			for _, mode := range []string{"inline", "display"} {
				expectedErrors["selection-"+label+"-"+child+"-"+mode] = true
			}
		}
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			selected := expectedErrors[c.Name]
			if selected {
				if c.RenderError == nil || *c.RenderError != "TypeError: this.element.addEventListener is not a function" {
					t.Fatal("changed primary postconstructor API error")
				}
			} else if c.RenderError != nil {
				t.Fatal("unexpected primary postconstructor error")
			}
			root := c.FullTree.node()
			before, nodes, parents := msInputSnapshot(root)
			o := pipeline.DefaultOptions()
			o.Display = c.Display
			r := &renderer{options: o, params: layout.TeXParameters, pxPerEm: o.Ex / layout.TeXParameters.XHeight}
			w := r.wrap(root, nil, 0, c.Display)
			at := func(path []int) *wrapper {
				q := w
				for _, i := range path {
					if i < 0 || i >= len(q.children) {
						t.Fatalf("invalid path %v", path)
					}
					q = q.children[i]
				}
				return q
			}
			verify := func() {
				for _, v := range c.Observations {
					q := at(v.Path)
					if q.scriptBaseCore() != at(v.BaseCorePath) {
						t.Errorf("%v core identity = %s want %s", v.Path, q.scriptBaseCore().node.Kind, at(v.BaseCorePath).node.Kind)
					}
					if q.node.Kind == "munder" || q.node.Kind == "mover" || q.node.Kind == "munderover" {
						if q.isMathAccent != v.IsMathAccent {
							t.Errorf("%v mathaccent=%v want %v", v.Path, q.isMathAccent, v.IsMathAccent)
						}
					}
					if q.baseHasAccent("accent") != v.AccentOver || q.baseHasAccent("accentunder") != v.AccentUnder {
						t.Errorf("%v first accents=(%v,%v) want(%v,%v)", v.Path, q.baseHasAccent("accent"), q.baseHasAccent("accentunder"), v.AccentOver, v.AccentUnder)
					}
				}
			}
			// All 34 maction fixtures have complete constructor observations but their
			// primary renderer later rejects a missing lite-adaptor event API. These
			// assertions cover constructor identity/state, not same-MathML paint parity.
			verify()
			w.outerBBox()
			verify()
			w.toSVG(NewElement("g"))
			verify()
			w.toSVG(NewElement("g"))
			verify()
			after, newNodes, newParents := msInputSnapshot(root)
			if before != after || len(nodes) != len(newNodes) {
				t.Fatal("source tree mutation")
			}
			for i, n := range nodes {
				if n != newNodes[i] || parents[i] != newParents[i] {
					t.Fatal("source identity mutation")
				}
			}
		})
	}
}

func TestAccentStretchInvalidatesMeasuredAncestors(t *testing.T) {
	for _, width := range []float64{0, 4, 6} {
		t.Run(fmt.Sprintf("width-%g", width), func(t *testing.T) {
			build := func() (*wrapper, *wrapper, *wrapper, *mml.Node) {
				arrow := mml.NewNode("mo", nil, nil, mml.NewText("→"))
				arrow.Attributes.Set("stretchy", true)
				sibling := mml.NewNode("mi", nil, nil, mml.NewText("x"))
				row := mml.NewNode("mrow", nil, nil, arrow, sibling)
				root := mml.NewNode("math", nil, nil, mml.NewNode("mstyle", nil, nil, row))
				w := tableTestWrapper(root)
				return w, w.children[0].children[0].children[0], w.children[0].children[0].children[1], root
			}
			warm, core, sibling, input := build()
			before, nodes, parents := msInputSnapshot(input)
			warm.outerBBox()
			oldSibling := *sibling.outerBBox()
			for p := core.parent; p != nil; p = p.parent {
				if !p.bboxComputed {
					t.Fatal("ancestor was not premeasured")
				}
			}
			if !core.canStretch(font.DirectionHorizontal) {
				t.Fatal("missing horizontal arrow stretch")
			}
			core.getStretchedVariant([]float64{width}, true)
			for p := core.parent; p != nil; p = p.parent {
				if p.bboxComputed {
					t.Fatal("stretched core left an ancestor cached")
				}
			}
			if width > 0 && (!core.bboxComputed || core.size != -1 || core.bbox.W != width) {
				t.Fatal("multi-part core's final measured box lost")
			}
			if width == 0 && core.size < 0 {
				t.Fatal("fixed-size control became multi-part")
			}
			cold, coldCore, _, _ := build()
			if !coldCore.canStretch(font.DirectionHorizontal) {
				t.Fatal("cold arrow stretch missing")
			}
			coldCore.getStretchedVariant([]float64{width}, true)
			if *warm.outerBBox() != *cold.outerBBox() {
				t.Fatal("premeasurement changed final parent geometry")
			}
			if *sibling.outerBBox() != oldSibling {
				t.Fatal("unaffected sibling geometry changed")
			}
			after, newNodes, newParents := msInputSnapshot(input)
			if before != after || len(nodes) != len(newNodes) {
				t.Fatal("stretch mutated input tree")
			}
			for i, n := range nodes {
				if n != newNodes[i] || parents[i] != newParents[i] {
					t.Fatal("stretch changed node/parent identity")
				}
			}
		})
	}
}
