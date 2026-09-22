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

type stackTree struct {
	Kind       string         `json:"kind"`
	Text       *string        `json:"text"`
	Attributes map[string]any `json:"attributes"`
	Children   []*stackTree   `json:"children"`
}

func projectStackTree(n *mml.Node) *stackTree {
	r := &stackTree{Kind: n.Kind, Attributes: map[string]any{}, Children: []*stackTree{}}
	if n.Kind == "mrow" && n.Flags.Inferred {
		r.Kind = "inferredMrow"
	}
	if n.Kind == "text" {
		text := n.Text
		r.Text = &text
	}
	for _, name := range n.Attributes.ExplicitNames() {
		r.Attributes[name], _ = n.Attributes.GetExplicit(name)
	}
	for _, child := range n.Children {
		r.Children = append(r.Children, projectStackTree(child))
	}
	return r
}
func TestStackPlacementPinnedReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/stack_placement_mathjax_3_2_2.json")
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
	if len(fixture.Cases) != 50 {
		t.Fatal("missing exact stack cases")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			wantTree, wantSVG, wantWidth, wantHeight := c.Tree, c.SVGSHA256, c.Width, c.Height
			root, err := tex.NewCompiler().Compile(c.TeX, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			// Round-trip only the actual tree so JSON's numeric representation matches.
			b, err := json.Marshal(projectStackTree(root))
			if err != nil {
				t.Fatal(err)
			}
			var actualTree *stackTree
			if err = json.Unmarshal(b, &actualTree); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(actualTree, wantTree) {
				t.Errorf("compiled tree differs from the complete bound tree")
			}
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			actual, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256([]byte(actual))); got != wantSVG {
				t.Errorf("complete SVG = %s; want %s", got, wantSVG)
			}
			if c.Display {
				w, h, err := mathjax.Measure(c.TeX)
				if err != nil {
					t.Fatal(err)
				}
				if w != wantWidth || h != wantHeight {
					t.Errorf("measure = %dx%d; want %dx%d", w, h, wantWidth, wantHeight)
				}
			}
		})
	}
}
