// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package tex

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
)

func TestInferredRowUnboundedConstruction(t *testing.T) {
	for _, kind := range []string{"mrow", "mfenced", "annotation"} {
		t.Run(kind, func(t *testing.T) {
			a, b, c, d := token("mi", "a"), token("mi", "b"), token("mi", "c"), token("mi", "d")
			inferred := forcedRow([]*mml.Node{b, c}, true)
			empty := forcedRow(nil, true)
			explicit := forcedRow([]*mml.Node{d}, false)
			input := []*mml.Node{a, inferred, empty, explicit}
			parent := node(kind, input...)
			if !reflect.DeepEqual(parent.Children, []*mml.Node{a, b, c, explicit}) {
				t.Fatal("unbounded constructor must flatten inferred children in order and retain explicit rows")
			}
			if !reflect.DeepEqual(input, []*mml.Node{a, inferred, empty, explicit}) || !reflect.DeepEqual(inferred.Children, []*mml.Node{b, c}) {
				t.Fatal("normalization changed the input slice or discarded the inferred row's original children")
			}
			for _, child := range parent.Children {
				if child.Parent != parent {
					t.Fatal("flattened child has the wrong parent")
				}
			}
			if d.Parent != explicit || inferred.Parent != nil || empty.Parent != nil {
				t.Fatal("flattening linked a discarded wrapper or changed the explicit subtree")
			}
		})
	}
}

func TestInferredRowFixedConstruction(t *testing.T) {
	a, b, den := token("mi", "a"), token("mi", "b"), token("mi", "c")
	inferred := forcedRow([]*mml.Node{a, b}, true)
	inferred.Attributes.Set("id", "numerator")
	inferred.SetProperty("sample", "preserved")
	parent := node("mfrac", inferred, den)
	if len(parent.Children) != 2 || parent.Children[0] == inferred || parent.Children[1] != den {
		t.Fatal("fixed arity must replace the inferred wrapper without changing child slots")
	}
	explicit := parent.Children[0]
	if explicit.Kind != "mrow" || explicit.Flags.Inferred || explicit.Flags.NotParent ||
		!reflect.DeepEqual(explicit.Children, []*mml.Node{a, b}) {
		t.Fatal("fixed-arity child is not an explicit row with original children")
	}
	if id, _ := explicit.Attributes.GetExplicit("id"); id != "numerator" {
		t.Fatal("explicit replacement lost attributes")
	}
	if value, _ := explicit.Property("sample"); value != "preserved" {
		t.Fatal("explicit replacement lost properties")
	}
	if explicit.Parent != parent || a.Parent != explicit || b.Parent != explicit || den.Parent != parent {
		t.Fatal("fixed-arity replacement lost ownership")
	}
}

func TestInferredRowNegativeArityConstruction(t *testing.T) {
	for _, kind := range []string{"math", "TeXAtom", "mstyle", "merror", "mpadded", "mphantom", "menclose", "mtd", "msqrt"} {
		t.Run(kind, func(t *testing.T) {
			a, b := token("mi", "a"), token("mi", "b")
			incoming := forcedRow([]*mml.Node{a, b}, true)
			incoming.Attributes.Set("id", "incoming-row")
			parent := node(kind, incoming)
			if len(parent.Children) != 1 {
				t.Fatal("negative arity must own exactly one inferred content row")
			}
			content := parent.Children[0]
			if content == incoming || content.Kind != "mrow" || !content.Flags.Inferred || !content.Flags.NotParent ||
				!reflect.DeepEqual(content.Children, []*mml.Node{a, b}) {
				t.Fatal("negative arity reused the incoming inferred row instead of constructing its own content row")
			}
			if content.Attributes.IsSet("id") || content.Parent != parent || a.Parent != content || b.Parent != content {
				t.Fatal("fresh content row retained incoming attributes or lost child ownership")
			}
			empty := node(kind, forcedRow(nil, true))
			if len(empty.Children) != 1 || !empty.Children[0].Flags.Inferred || len(empty.Children[0].Children) != 0 {
				t.Fatal("empty input must retain one empty inferred content row")
			}
		})
	}
}

// The original observation is constructed MathML, not an authored Studio input.
func TestInferredRowOriginalSVG(t *testing.T) {
	var fixture struct {
		MathjaxGitCommit, OriginalRecordSHA256, PassiveRecordSHA256 string
		SVG, SVGSHA256                                              string
		Spec                                                        struct {
			Display bool
			Tree    json.RawMessage
		}
	}
	data, err := os.ReadFile("testdata/inferred_row_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" ||
		fixture.OriginalRecordSHA256 != "1102dc6b6edc63d6f93d8cd43777758988e0c9fc210a253a1f945f7ec952fce1" ||
		fixture.PassiveRecordSHA256 != "06f5aef129fcd0a42e2cab78fabeb1d7574bdb6880083efbc19be89279035b07" ||
		fmt.Sprintf("%x", sha256.Sum256([]byte(fixture.SVG))) != fixture.SVGSHA256 {
		t.Fatal("unbound original inferred-row reference")
	}
	root := idAnchorBuild(fixture.Spec.Tree, fixture.Spec.Display)
	row := root.Children[0].Children[0]
	if len(row.Children) != 3 || row.Children[1].Kind != "mi" || row.Children[1].Parent != row {
		t.Error("original alias target must be directly at model path [0,0,1]")
	}
	options := pipeline.DefaultOptions()
	options.Display = fixture.Spec.Display
	got, err := svg.NewTypesetter().Typeset(root, options)
	if err != nil {
		t.Fatal(err)
	}
	if got != fixture.SVG {
		t.Errorf("whole original inferred-row SVG differs\ngot: %s\nwant: %s", got, fixture.SVG)
	}
}
