package mathjax_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/tex"
)

func TestAnnotationAttachmentPinnedReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/annotation_attachment_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name, TeX, SVGSHA256 string
			Display              bool
			Width, Height        int
			Tree                 *stackTree
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile("testdata/annotation_attachment_boundaries.json")
	if err != nil {
		t.Fatal(err)
	}
	var boundaries struct {
		Baseline string
		Cases    map[string]struct {
			TeX                                                   string
			Display                                               bool
			PrimarySVG, ExpectedSVG, BaselineTreeSHA256, Category string
			Tree                                                  *stackTree
		}
	}
	if err = json.Unmarshal(data, &boundaries); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 48 || len(boundaries.Cases) != 4 || boundaries.Baseline != "97eebae061dc96a046462472574ef2d6d4772078" {
		t.Fatal("changed bound annotation/prime matrix")
	}
	rawBoundaries := 0
	primaryErrors := 0
	for _, c := range fixture.Cases {
		if c.Tree.Children[0].Children[0].Kind == "merror" {
			primaryErrors++
		}
		t.Run(c.Name, func(t *testing.T) {
			expectedSVG, expectedTree := c.SVGSHA256, c.Tree
			if b, ok := boundaries.Cases[c.Name]; ok {
				if b.TeX != c.TeX || b.Display != c.Display || b.PrimarySVG != c.SVGSHA256 || b.Category == "" || len(b.BaselineTreeSHA256) != 64 {
					t.Fatal("changed precise accepted-parent boundary")
				}
				if b.ExpectedSVG != b.PrimarySVG {
					rawBoundaries++
				}
				expectedSVG, expectedTree = b.ExpectedSVG, b.Tree
			}
			root, err := tex.NewCompiler().Compile(c.TeX, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			root.Walk(func(n *mml.Node) bool {
				if _, ok := n.Property("_texLimitsScriptOrigin"); ok {
					t.Error("transient script origin leaked into public tree")
				}
				return true
			})
			data, err := json.Marshal(projectStackTree(root))
			if err != nil {
				t.Fatal(err)
			}
			var actualTree *stackTree
			if err = json.Unmarshal(data, &actualTree); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(actualTree, expectedTree) {
				t.Error("complete explicit kind/text/attribute/child tree differs")
			}
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			actual, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256([]byte(actual))); got != expectedSVG {
				t.Errorf("complete SVG=%s want=%s", got, expectedSVG)
			}
			repeated, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil || repeated != actual {
				t.Fatalf("repeat render changed: %v", err)
			}
			// Complete SVG bytes bind inline and display dimensions. Display mode also
			// exercises the public measurement API for every raw primary case.
			if c.Display && expectedSVG == c.SVGSHA256 {
				width, height, err := mathjax.Measure(c.TeX)
				if err != nil || width != c.Width || height != c.Height {
					t.Fatalf("measure=%dx%d,%v want=%dx%d", width, height, err, c.Width, c.Height)
				}
			}
		})
	}
	if rawBoundaries != 0 || primaryErrors != 14 {
		t.Fatalf("prime boundaries/required primary errors=%d/%d want0/14", rawBoundaries, primaryErrors)
	}
}
