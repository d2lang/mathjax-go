// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package tex

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/ordered"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
)

func TestAnnotationChildrenDoNotInherit(t *testing.T) {
	for _, kind := range []string{"annotation", "annotation-xml"} {
		t.Run(kind, func(t *testing.T) {
			child := node("mi", mml.NewText("x"))
			child.Attributes.Set("id", "child")
			row := node("mrow", child)
			annotation := node(kind, row)
			attrs := ordered.New[inheritedAttribute]()
			attrs.Set("mathvariant", inheritedAttribute{source: "mstyle", value: "bold"})
			setInheritedAttributes(annotation, attrs, true, 2, true)
			for _, n := range []*mml.Node{row, child} {
				if names := n.Attributes.InheritedNames(); len(names) != 0 {
					t.Errorf("%s child inherited attributes: %v", n.Kind, names)
				}
				if _, ok := n.Property("texprimestyle"); ok {
					t.Errorf("%s child inherited prime style", n.Kind)
				}
			}
			if id, _ := child.Attributes.GetExplicit("id"); id != "child" || child.Parent != row || row.Parent != annotation {
				t.Fatal("annotation child identity or ownership changed")
			}
			if level, _ := annotation.Attributes.GetInherited("scriptlevel"); level != 2 {
				t.Errorf("annotation itself lost its script level: %v", level)
			}
			if display, _ := annotation.Attributes.GetInherited("displaystyle"); display != true {
				t.Errorf("annotation itself lost its display style: %v", display)
			}
			control := node("mi", mml.NewText("x"))
			setInheritedAttributes(control, attrs, true, 2, true)
			if variant, _ := control.Attributes.GetInherited("mathvariant"); variant != "bold" {
				t.Errorf("ordinary mi lost inherited variant: %v", variant)
			}
			if level, _ := control.Attributes.GetInherited("scriptlevel"); level != 2 {
				t.Errorf("ordinary mi lost inherited script level: %v", level)
			}
		})
	}
}

// These are constructed MathML inputs, not claims of an ordinary TeX route.
// Keep the original SVGs untouched; adjacent normal mi nodes are italic controls.
func TestAnnotationInheritanceOriginalSVG(t *testing.T) {
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, SVG, SVGSHA256, OriginalRecordSHA256 string
			Spec                                       struct {
				Display bool
				Tree    json.RawMessage
			}
		}
	}
	data, err := os.ReadFile("testdata/annotation_inheritance_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 2 {
		t.Fatal("unbound annotation references")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			if len(c.OriginalRecordSHA256) != 64 || fmt.Sprintf("%x", sha256.Sum256([]byte(c.SVG))) != c.SVGSHA256 {
				t.Fatal("invalid original SVG reference")
			}
			root := idAnchorBuild(c.Spec.Tree, c.Spec.Display)
			options := pipeline.DefaultOptions()
			options.Display = c.Spec.Display
			got, err := svg.NewTypesetter().Typeset(root, options)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.SVG {
				t.Errorf("annotation SVG differs from original: got %x, want %s", sha256.Sum256([]byte(got)), c.SVGSHA256)
			}
		})
	}
}
