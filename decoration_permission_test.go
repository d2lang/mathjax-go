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

func TestDecorationPermissionPinnedReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/decoration_permission_mathjax_3_2_2.json")
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
	data, err = os.ReadFile("testdata/decoration_permission_boundaries.json")
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
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 60 || bounds.Baseline != "efd6d63071862218af992a6074e3c8680b634ab4" || len(bounds.Cases) != 50 {
		t.Fatal("unbound decoration-permission matrix")
	}
	rawSVG, rawOwn, rawExplicit := 0, 0, 0
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			wantSVG := c.SVGSHA256
			// Qualification works on a fresh tree; the frozen primary remains intact.
			primaryBytes, err := json.Marshal(c.PropertiesTree)
			if err != nil {
				t.Fatal(err)
			}
			var wantTree *limitsTree
			if err = json.Unmarshal(primaryBytes, &wantTree); err != nil {
				t.Fatal(err)
			}
			rawTree, rawAttributes := true, true
			if q, ok := bounds.Cases[c.Name]; ok {
				if q.PrimarySHA256 != c.SVGSHA256 || len(q.BaselineTreeSHA256) != 64 {
					t.Fatal("changed exact historical binding")
				}
				rawTree = false
				switch q.Category {
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
	if rawSVG != 60 || rawExplicit != 12 || rawOwn != 10 {
		t.Fatalf("raw SVG/explicit/own counts %d/%d/%d want60/12/10", rawSVG, rawExplicit, rawOwn)
	}
}
