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

func TestDecorationMovablePinnedReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/decoration_movable_mathjax_3_2_2.json")
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
	data, err = os.ReadFile("testdata/decoration_movable_boundaries.json")
	if err != nil {
		t.Fatal(err)
	}
	var bounds struct {
		Baseline string
		Cases    map[string]struct {
			Category, PrimarySHA256, ExpectedSHA256, BaselineTreeSHA256 string
			ExpectedTree                                                *limitsTree
			Differences                                                 []struct {
				Path                       []int
				Field                      string
				BaselineKind, ExpectedKind string
				Primary, Expected          map[string]any
			}
		}
	}
	if err = json.Unmarshal(data, &bounds); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 60 || bounds.Baseline != "69ae099022c270469016a6a83f9f9ae13d98e320" || len(bounds.Cases) != 54 {
		t.Fatal("unbound ordinary-decoration matrix")
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
				case "unchanged-D071-error", "unchanged-scripted-decoration":
					if q.Category == "unchanged-scripted-decoration" && c.Name != "overline-scripted-sum-inline" && c.Name != "overline-scripted-sum-display" {
						t.Fatal("unexpected scripted-decoration boundary")
					}
					if q.ExpectedTree == nil || q.ExpectedSHA256 == c.SVGSHA256 || len(q.Differences) != 0 {
						t.Fatal("invalid exact error boundary")
					}
					wantTree, wantSVG = q.ExpectedTree, q.ExpectedSHA256
					rawAttributes = false
				case "exact-inherited-fields":
					if q.ExpectedTree != nil || q.ExpectedSHA256 != c.SVGSHA256 || len(q.Differences) == 0 {
						t.Fatal("invalid exact metadata boundary")
					}
					for _, d := range q.Differences {
						at := limitsNodeAt(wantTree, d.Path)
						if at == nil || at.Kind != d.ExpectedKind {
							t.Fatal("missing exact bound node kind")
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
	if rawSVG != 36 || rawExplicit != 8 || rawOwn != 6 {
		t.Fatalf("raw SVG/explicit/own counts %d/%d/%d want36/8/6", rawSVG, rawExplicit, rawOwn)
	}
}
