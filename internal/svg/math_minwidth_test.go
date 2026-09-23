package svg

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/layout"
	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/ordered"
	"github.com/d2lang/mathjax-go/internal/pipeline"
)

type minWidthPair struct {
	Name  string
	Value any
}
type minWidthNode struct {
	Kind       string
	Text       *string
	TeXClass   mml.TeXClass
	Attributes struct{ Explicit, Inherited, Defaults, Global []minWidthPair }
	Properties []minWidthPair
	Flags      struct{ NotParent, Inferred, Token, Embellished bool }
	Children   []minWidthNode
}

func minWidthMap(pairs []minWidthPair) *ordered.Map[mml.Property] {
	out := ordered.New[mml.Property]()
	for _, p := range pairs {
		v := p.Value
		if v == "_inherit_" {
			v = mml.Inherit
		}
		out.Set(p.Name, v)
	}
	return out
}
func (s minWidthNode) build() *mml.Node {
	k := s.Kind
	if k == "inferredMrow" {
		k = "mrow"
	}
	n := mml.NewNode(k, minWidthMap(s.Attributes.Defaults), minWidthMap(s.Attributes.Global))
	n.Attributes.SetList(minWidthMap(s.Attributes.Explicit))
	for _, p := range s.Attributes.Inherited {
		v := p.Value
		if v == "_inherit_" {
			v = mml.Inherit
		}
		n.Attributes.SetInherited(p.Name, v)
	}
	n.Properties = minWidthMap(s.Properties)
	n.TeXClass = s.TeXClass
	n.Flags.NotParent = s.Flags.NotParent
	n.Flags.Inferred = s.Flags.Inferred
	n.Flags.Token = s.Flags.Token
	n.Flags.Embellished = s.Flags.Embellished
	if s.Text != nil {
		n.Text = *s.Text
	}
	for _, c := range s.Children {
		n.AppendChild(c.build())
	}
	return n
}

type minWidthFixture struct {
	Constructors []struct {
		Name         string
		Tree         minWidthNode
		SelectedPath *[]int
		SVG          *string
		Display      bool
	}
	Methods []struct {
		Name                                         string
		Attributes                                   map[string]any
		Scale                                        float64
		PWidth                                       string
		SelectedBox                                  *layout.BBox
		Before, After, Shift, Em, Ex, ContainerWidth float64
		Align                                        string
	}
}

func minWidthCases(t *testing.T) minWidthFixture {
	t.Helper()
	b, e := os.ReadFile("testdata/responsive_minwidth_contract.json")
	if e != nil {
		t.Fatal(e)
	}
	var f minWidthFixture
	if e = json.Unmarshal(b, &f); e != nil {
		t.Fatal(e)
	}
	return f
}
func TestResponsiveMinWidthConstructors(t *testing.T) {
	f := minWidthCases(t)
	if len(f.Constructors) != 42 {
		t.Fatal(len(f.Constructors))
	}
	for _, c := range f.Constructors {
		t.Run(c.Name, func(t *testing.T) {
			root := c.Tree.build()
			o := pipeline.DefaultOptions()
			o.Display = c.Display
			r := &renderer{options: o, params: layout.TeXParameters, pxPerEm: o.Ex / layout.TeXParameters.XHeight}
			w := r.wrap(root, nil, 0, o.Display)
			var want *mml.Node
			if c.SelectedPath != nil {
				want = root
				for _, i := range *c.SelectedPath {
					want = want.Children[i]
				}
			}
			if (r.table == nil) != (want == nil) || r.table != nil && r.table.node != want {
				t.Fatal("constructor-selected table identity mismatch")
			}
			selected := r.table
			var tables []*wrapper
			var walk func(*wrapper)
			walk = func(x *wrapper) {
				for _, child := range x.children {
					walk(child)
				}
				if x.node.Kind == "mtable" {
					tables = append(tables, x)
				}
			}
			walk(w)
			for i := len(tables) - 1; i >= 0; i-- {
				tables[i].outerBBox()
				if r.table != selected {
					t.Fatal("lazy/reverse measurement changed selected identity")
				}
			}
			other := &renderer{options: o, params: layout.TeXParameters, pxPerEm: r.pxPerEm}
			other.wrap(mml.NewNode("math", nil, nil), nil, 0, true)
			if other.table != nil {
				t.Fatal("renderer selection leaked across renders")
			}
		})
	}
}
func TestResponsiveMinWidthMML(t *testing.T) {
	f := minWidthCases(t)
	count := 0
	for _, c := range f.Constructors {
		if c.SVG == nil {
			continue
		}
		count++
		t.Run(c.Name, func(t *testing.T) {
			o := pipeline.DefaultOptions()
			o.Display = c.Display
			s, e := NewTypesetter().Typeset(c.Tree.build(), o)
			if e != nil {
				t.Fatal(e)
			}
			if s != *c.SVG {
				t.Fatalf("complete primary SVG mismatch\ngot: %s\nwant: %s", s, *c.SVG)
			}
		})
	}
	if count != 31 {
		t.Fatal("strict MML count", count)
	}
}
func TestResponsiveMinWidthActualMethods(t *testing.T) {
	f := minWidthCases(t)
	if len(f.Methods) != 81 {
		t.Fatal(len(f.Methods))
	}
	for _, c := range f.Methods {
		t.Run(c.Name, func(t *testing.T) {
			o := pipeline.DefaultOptions()
			o.Em = c.Em
			o.Ex = c.Ex
			if 80*o.Ex != c.ContainerWidth {
				t.Fatal("captured metric contract")
			}
			r := &renderer{options: o, params: layout.TeXParameters, pxPerEm: o.Ex / layout.TeXParameters.XHeight, minWidth: c.Before}
			n := mml.NewNode("math", nil, nil)
			for k, v := range c.Attributes {
				n.Attributes.Set(k, v)
			}
			w := &wrapper{renderer: r, node: n, bbox: layout.ZeroBBox(), bboxComputed: true}
			w.bbox.Scale = c.Scale
			w.bbox.PWidth = c.PWidth
			if c.SelectedBox != nil {
				r.table = &wrapper{renderer: r, node: mml.NewNode("mtable", nil, nil), bbox: c.SelectedBox.Clone(), bboxComputed: true}
			}
			before := r.table
			var box *layout.BBox
			if before != nil {
				box = before.bbox.Clone()
			}
			align, shift := w.mathAlignShift()
			if align != c.Align || shift != c.Shift {
				t.Fatalf("actual getAlignShift = %q %v, want %q %v", align, shift, c.Align, c.Shift)
			}
			w.handleMathMinWidth()
			if r.minWidth != c.After {
				t.Fatalf("actual handleDisplay minwidth = %.17g, want %.17g", r.minWidth, c.After)
			}
			if r.table != before || before != nil && !reflect.DeepEqual(before.bbox, box) {
				t.Fatal("selected identity or cached box mutated")
			}
		})
	}
}

// These source-contract controls distinguish raw pixel containerWidth from
// the em reference used by existing table layout consumers.
func TestResponsiveMinWidthAlignmentUnits(t *testing.T) {
	for _, c := range []struct {
		name, shift string
		want        float64
	}{{"percent", "5%", 32}, {"unitless", "0.5", 320}, {"absolute", "8px", .221}} {
		t.Run(c.name, func(t *testing.T) {
			o := pipeline.DefaultOptions()
			r := &renderer{options: o, params: layout.TeXParameters, pxPerEm: o.Ex / layout.TeXParameters.XHeight}
			n := mml.NewNode("math", nil, nil)
			n.Attributes.Set("indentalign", "right")
			n.Attributes.Set("indentshift", c.shift)
			w := &wrapper{renderer: r, node: n, bbox: layout.ZeroBBox()}
			w.bbox.Scale = 2
			a, s := w.mathAlignShift()
			if a != "right" || s != c.want {
				t.Fatalf("right explicit shift = %q %.17g, want %.17g", a, s, c.want)
			}
		})
	}
}
