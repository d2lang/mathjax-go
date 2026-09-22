package mathjax_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
	"github.com/d2lang/mathjax-go/internal/tex"
)

func TestAccentScriptBasePinnedReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/accent_script_base_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVGSHA256 string
			Display              bool
			Width, Height        int
			PropertiesTree       *limitsTree
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile("testdata/accent_script_base_boundaries.json")
	if err != nil {
		t.Fatal(err)
	}
	var bounds struct {
		Baseline string
		Cases    map[string]struct {
			Category, PrimarySHA256, ExpectedSHA256, BaselineTreeSHA256 string
			ExpectedTree                                                *limitsTree
			Differences                                                 []struct {
				Path, BaselinePath []int
				Field              string
				Primary, Expected  map[string]any
			}
		}
	}
	if err = json.Unmarshal(data, &bounds); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 52 || bounds.Baseline != "f1ae550cda1fb97e41bebf86799ffeae48cd3042" || len(bounds.Cases) != 44 {
		t.Fatal("unbound accent matrix")
	}
	rawSVG, rawOwn, rawExplicit := 0, 0, 0
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			wantTree, wantSVG := c.PropertiesTree, c.SVGSHA256
			rawTree, rawAttributes := true, true
			if q, ok := bounds.Cases[c.Name]; ok {
				if q.PrimarySHA256 != c.SVGSHA256 || len(q.BaselineTreeSHA256) != 64 {
					t.Fatal("changed exact historical binding")
				}
				rawTree = false
				switch q.Category {
				case "exact-inherited-own-fields":
					if q.ExpectedTree != nil || q.ExpectedSHA256 != c.SVGSHA256 || len(q.Differences) == 0 {
						t.Fatal("invalid exact metadata boundary")
					}
					for _, d := range q.Differences {
						at := limitsNodeAt(wantTree, d.Path)
						if at == nil {
							t.Fatal("missing bound node")
						}
						field := limitsField(at, d.Field)
						if field == nil || !reflect.DeepEqual(field, d.Primary) {
							t.Fatal("primary fields changed at exact boundary")
						}
						if d.Field == "attributes" {
							at.Attributes = d.Expected
							rawAttributes = false
						} else if d.Field == "properties" {
							at.Properties = d.Expected
						} else {
							t.Fatal("unexpected field")
						}
					}
				default:
					t.Fatal("unknown qualification")
				}
			}
			if wantSVG == c.SVGSHA256 {
				rawSVG++
			}
			if rawTree {
				rawOwn++
			}
			if rawAttributes {
				rawExplicit++
			}
			root, err := tex.NewCompiler().Compile(c.TeX, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			// JSON round-trip converts the Go numeric Property representation to the
			// primary fixture's JSON number representation; no fields are omitted.
			raw, err := json.Marshal(limitsProjection(root))
			if err != nil {
				t.Fatal(err)
			}
			var actualTree *limitsTree
			if err = json.Unmarshal(raw, &actualTree); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(actualTree, wantTree) {
				t.Error("complete kind/text/explicit attributes/own properties/children differ")
			}
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			actual, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256([]byte(actual))); got != wantSVG {
				t.Errorf("whole SVG %s want %s", got, wantSVG)
			}
			repeat, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil || repeat != actual {
				t.Fatal("repeat rendering changed")
			}
			if c.Display && wantSVG == c.SVGSHA256 {
				width, height, err := mathjax.Measure(c.TeX)
				if err != nil || width != c.Width || height != c.Height {
					t.Fatalf("measure %dx%d %v want %dx%d", width, height, err, c.Width, c.Height)
				}
			}
		})
	}
	if rawSVG != 52 || rawExplicit != 48 || rawOwn != 8 {
		t.Fatalf("raw SVG/explicit/own counts %d/%d/%d want52/48/8", rawSVG, rawExplicit, rawOwn)
	}
}

// Constructor-time base measurement warms the transparent wrapper around an
// x-arrow. Its later multi-part stretch must invalidate that measured wrapper.
func TestAccentScriptXArrowPinnedReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/accent_script_xarrow_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVGSHA256 string
			Display              bool
			Width, Height        int
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 2 {
		t.Fatal("unbound x-arrow reference")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			actual, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if fmt.Sprintf("%x", sha256.Sum256([]byte(actual))) != c.SVGSHA256 {
				t.Fatal("complete x-arrow SVG differs")
			}
			if c.Display {
				w, h, err := mathjax.Measure(c.TeX)
				if err != nil || w != c.Width || h != c.Height {
					t.Fatalf("measure %dx%d %v want %dx%d", w, h, err, c.Width, c.Height)
				}
			}
		})
	}
}
