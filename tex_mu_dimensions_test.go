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
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
	"github.com/d2lang/mathjax-go/internal/tex"
)

func TestTeXMuDimensionsPinnedReferences(t *testing.T) {
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVGSHA256, DirectMathMLWidth string
			Display                                 bool
			Width, Height                           int
			Tree                                    *stackTree
			PropertiesTree                          *limitsTree
		}
	}
	read := func(path string, out any) {
		t.Helper()
		b, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		if e = json.Unmarshal(b, out); e != nil {
			t.Fatal(e)
		}
	}
	read("testdata/tex_mu_dimensions_mathjax_3_2_2.json", &fixture)
	var boundaries struct {
		Baseline string
		Cases    map[string]struct {
			TeX, Mode, BaselineSVG, ExpectedSVG, PrimarySVG string
			Display                                         bool
			BaselineTree                                    *limitsTree
			AttributesAtRootFirstChild                      map[string]any
		}
	}
	read("testdata/tex_mu_dimensions_boundaries.json", &boundaries)
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 86 || boundaries.Baseline != "ab6c3c973d305288864a96a8bbab5d5471762c6b" || len(boundaries.Cases) != 18 {
		t.Fatal("unbound mu corpus")
	}
	expectedModes := map[string]string{"raise": "mu-consumer", "lower": "mu-consumer", "cd-height": "svg-only", "braced-junk": "unchanged", "exponent": "unchanged", "double-sign": "unchanged", "uppercase": "unchanged", "unbraced-comma": "unchanged", "pc": "unchanged"}
	for name, b := range boundaries.Cases {
		stem := strings.TrimSuffix(strings.TrimSuffix(name, "-inline"), "-display")
		if expectedModes[stem] != b.Mode {
			t.Fatal("unexpected boundary", name, b.Mode)
		}
		if b.Mode != "mu-consumer" && (b.ExpectedSVG != b.BaselineSVG || len(b.AttributesAtRootFirstChild) != 0) {
			t.Fatal("inherited SVG boundary changed", name)
		}
	}
	// D079: these exact 12 historical grammar qualifications now compare raw
	// primary SVG and trees. Their unchanged original receipts remain on disk.
	for _, stem := range []string{"braced-junk", "exponent", "double-sign", "uppercase", "unbraced-comma", "pc"} {
		for _, mode := range []string{"inline", "display"} {
			name := stem + "-" + mode
			if boundaries.Cases[name].Mode != "unchanged" {
				t.Fatal("missing historical grammar receipt", name)
			}
			delete(boundaries.Cases, name)
		}
	}
	// PositionItem now matches the original raise/lower shape and signs.
	// Retain the historical receipts, but compare these cases to the raw oracle.
	for _, stem := range []string{"raise", "lower"} {
		for _, mode := range []string{"inline", "display"} {
			delete(boundaries.Cases, stem+"-"+mode)
		}
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			root, err := tex.NewCompiler().Compile(c.TeX, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			if c.DirectMathMLWidth != "" {
				spaces := root.Find("mspace")
				if len(spaces) != 1 {
					t.Fatal("MathML space inventory")
				}
				spaces[0].Attributes.Set("width", c.DirectMathMLWidth)
			}
			wantTree, wantSVG := c.PropertiesTree, c.SVGSHA256
			b, qualified := boundaries.Cases[c.Name]
			if qualified {
				if b.TeX != c.TeX || b.Display != c.Display || b.PrimarySVG != c.SVGSHA256 || b.ExpectedSVG == c.SVGSHA256 {
					t.Fatal("boundary/source mismatch")
				}
				wantSVG = b.ExpectedSVG
				if b.Mode == "unchanged" {
					wantTree = b.BaselineTree
				}

			}
			ownJSON, err := json.Marshal(limitsProjection(root))
			if err != nil {
				t.Fatal(err)
			}
			var gotOwn *limitsTree
			if err = json.Unmarshal(ownJSON, &gotOwn); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(gotOwn, wantTree) {
				t.Fatal("complete explicit/own-property tree differs")
			}
			if !qualified || b.Mode == "svg-only" {
				explicitJSON, err := json.Marshal(projectStackTree(root))
				if err != nil {
					t.Fatal(err)
				}
				var gotExplicit *stackTree
				if err = json.Unmarshal(explicitJSON, &gotExplicit); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(gotExplicit, c.Tree) {
					t.Fatal("complete explicit tree differs from primary")
				}
			}
			var result string
			if c.DirectMathMLWidth != "" {
				opts := pipeline.DefaultOptions()
				opts.Display = c.Display
				result, err = svg.NewTypesetter().Typeset(root, opts)
				v, _ := root.Find("mspace")[0].Attributes.GetExplicit("width")
				if v != c.DirectMathMLWidth {
					t.Fatal("direct MathML width changed")
				}
			} else {
				opts := mathjax.DefaultOptions()
				opts.Display = c.Display
				result, err = mathjax.RenderWithOptions(c.TeX, opts)
			}
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256([]byte(result))); got != wantSVG {
				t.Fatalf("complete SVG=%s want=%s", got, wantSVG)
			}
			if c.Display && !qualified && c.DirectMathMLWidth == "" {
				w, h, e := mathjax.Measure(c.TeX)
				if e != nil || w != c.Width || h != c.Height {
					t.Fatalf("dimensions=%dx%d %v want=%dx%d", w, h, e, c.Width, c.Height)
				}
			}
		})
	}
}
