package mathjax_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/tex"
)

type dimensionGrammarTree struct {
	Kind       string                  `json:"kind"`
	Text       *string                 `json:"text"`
	Attributes map[string]any          `json:"attributes"`
	Properties map[string]any          `json:"properties"`
	Children   []*dimensionGrammarTree `json:"children"`
}

func dimensionGrammarProjection(n *mml.Node) *dimensionGrammarTree {
	k := n.Kind
	if k == "mrow" && n.Flags.Inferred {
		k = "inferredMrow"
	}
	r := &dimensionGrammarTree{Kind: k, Attributes: map[string]any{}, Properties: map[string]any{}, Children: []*dimensionGrammarTree{}}
	if n.Kind == "text" {
		s := n.Text
		r.Text = &s
	}
	n.Attributes.Explicit().Range(func(k string, v any) bool { r.Attributes[k] = v; return true })
	n.Properties.Range(func(k string, v any) bool { r.Properties[k] = v; return true })
	for _, c := range n.Children {
		r.Children = append(r.Children, dimensionGrammarProjection(c))
	}
	return r
}
func TestDimensionGrammarPinnedReferences(t *testing.T) {
	var f struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVGSHA256 string
			Display              bool
			Tree                 *dimensionGrammarTree
		}
	}
	var boundaries struct {
		Baseline string
		Cases    map[string]struct {
			Name, TeX, PrimarySVG, BaselineSVG string
			Display                            bool
			BaselineTree                       *dimensionGrammarTree
		}
	}
	read := func(p string, v any) {
		t.Helper()
		b, e := os.ReadFile(p)
		if e != nil {
			t.Fatal(e)
		}
		if e = json.Unmarshal(b, v); e != nil {
			t.Fatal(e)
		}
	}
	read("testdata/dimension_grammar_mathjax_3_2_2.json", &f)
	read("testdata/dimension_grammar_boundaries.json", &boundaries)
	if f.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(f.Cases) != 60 || boundaries.Baseline != "6a3573f7f15fc67b5abde6fdf8b548e89d6c7e48" || len(boundaries.Cases) != 2 {
		t.Fatal("unbound dimension corpus")
	}
	stems := map[string]bool{"rule-unchanged-boundary": true}
	for name := range boundaries.Cases {
		stem := strings.TrimSuffix(strings.TrimSuffix(name, "-inline"), "-display")
		if !stems[stem] {
			t.Fatal("unexpected qualification", name)
		}
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			want, hash := c.Tree, c.SVGSHA256
			if b, ok := boundaries.Cases[c.Name]; ok {
				if b.Name != c.Name || b.TeX != c.TeX || b.Display != c.Display || b.PrimarySVG != c.SVGSHA256 || b.PrimarySVG == b.BaselineSVG {
					t.Fatal("invalid retained boundary")
				}
				want, hash = b.BaselineTree, b.BaselineSVG
			}
			if strings.HasPrefix(c.Name, "raise-valid-boundary-") {
				// The reader fixes the comma. Existing RaiseLower omits only the leading
				// plus on voffset; the entire remaining primary tree and SVG stay strict.
				b, _ := json.Marshal(want)
				var copyTree *dimensionGrammarTree
				if e := json.Unmarshal(b, &copyTree); e != nil {
					t.Fatal(e)
				}
				want = copyTree
				attrs := want.Children[0].Children[0].Attributes
				if !reflect.DeepEqual(attrs, map[string]any{"voffset": "+1.5pt", "height": "+1.5pt", "depth": "-1.5pt"}) {
					t.Fatal("primary raise shape changed")
				}
				attrs["voffset"] = "1.5pt"
			}
			root, e := tex.NewCompiler().Compile(c.TeX, c.Display)
			if e != nil {
				t.Fatal(e)
			}
			b, e := json.Marshal(dimensionGrammarProjection(root))
			if e != nil {
				t.Fatal(e)
			}
			var got *dimensionGrammarTree
			if e = json.Unmarshal(b, &got); e != nil {
				t.Fatal(e)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatal("complete explicit and own-property tree differs")
			}
			opts := mathjax.DefaultOptions()
			opts.Display = c.Display
			s, e := mathjax.RenderWithOptions(c.TeX, opts)
			if e != nil {
				t.Fatal(e)
			}
			if actual := fmt.Sprintf("%x", sha256.Sum256([]byte(s))); actual != hash {
				t.Fatalf("whole SVG=%s want=%s", actual, hash)
			}
		})
	}
}

func TestDimensionGrammarUnsupportedCommandErrors(t *testing.T) {
	var f struct {
		Baseline string
		Cases    []struct {
			Name, TeX, ExpectedSVG, PrimarySVG string
			Display                            bool
			Tree, PrimaryTree                  *dimensionGrammarTree
		}
	}
	b, e := os.ReadFile("testdata/dimension_grammar_alias_boundaries.json")
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(b, &f); e != nil {
		t.Fatal(e)
	}
	if f.Baseline != "6a3573f7f15fc67b5abde6fdf8b548e89d6c7e48" || len(f.Cases) != 12 {
		t.Fatal("unbound alias controls")
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			alias := strings.HasPrefix(c.Name, "vspace-") || strings.HasPrefix(c.Name, "raisebox-")
			if alias {
				if c.ExpectedSVG == c.PrimarySVG {
					t.Fatal("historical alias witness changed")
				}
			} else if c.ExpectedSVG != c.PrimarySVG || !reflect.DeepEqual(c.Tree, c.PrimaryTree) {
				t.Fatal("supported reader control must be raw primary")
			}
			root, e := tex.NewCompiler().Compile(c.TeX, c.Display)
			if e != nil {
				t.Fatal(e)
			}
			b, _ := json.Marshal(dimensionGrammarProjection(root))
			var got *dimensionGrammarTree
			if e = json.Unmarshal(b, &got); e != nil {
				t.Fatal(e)
			}
			if !reflect.DeepEqual(got, c.PrimaryTree) {
				t.Fatal("complete alias/control tree differs")
			}
			opts := mathjax.DefaultOptions()
			opts.Display = c.Display
			s, e := mathjax.RenderWithOptions(c.TeX, opts)
			if e != nil {
				t.Fatal(e)
			}
			if fmt.Sprintf("%x", sha256.Sum256([]byte(s))) != c.PrimarySVG {
				t.Fatal("complete alias/control SVG differs")
			}
		})
	}
}
