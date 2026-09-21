package tex

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

type accentMarker struct {
	Text    string `json:"text"`
	Defined bool   `json:"defined"`
	Value   any    `json:"value"`
}

func TestMathAccentPinnedMarkers(t *testing.T) {
	data, err := os.ReadFile("../../testdata/arrow_accent_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name, TeX string
			Display   bool
			Markers   []accentMarker
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 48 {
		t.Fatal("missing marker fixtures")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			root, err := NewCompiler().Compile(c.TeX, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			got := []accentMarker{}
			root.Walk(func(n *mml.Node) bool {
				if n.Kind == "mo" {
					v, ok := n.Property("mathaccent")
					got = append(got, accentMarker{textContent(n), ok, v})
				}
				return true
			})
			want := c.Markers
			if c.Name == "ordinary-arrows-inline" || c.Name == "ordinary-arrows-display" {
				// The pinned post-filter combines adjacent up/down mo tokens.
				// Preserve that recorded tree boundary explicitly: neither the
				// combined primary text nor either Go token is a math accent.
				primary := []accentMarker{{"→", false, nil}, {"+", false, nil}, {"↑↓", false, nil}}
				if !reflect.DeepEqual(want, primary) {
					t.Fatal("ordinary-arrow primary token grouping changed")
				}
				want = []accentMarker{{"→", false, nil}, {"+", false, nil}, {"↑", false, nil}, {"↓", false, nil}}
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("operator marker policy %v, want %v", got, want)
			}
		})
	}
}

func TestMathAccentInheritanceBoundaries(t *testing.T) {
	for _, c := range []struct {
		name, mark, context string
		defined             bool
		value               any
		wantDefined         bool
		want                any
	}{
		{"over", "→", "mover", false, nil, true, true},
		{"under", "→", "munder", false, nil, true, true},
		{"both", "→", "munderover", false, nil, true, true},
		{"left excluded", "←", "mover", false, nil, false, nil},
		{"multiple excluded", "→→", "mover", false, nil, false, nil},
		{"defined false", "→", "mover", true, false, true, false},
		{"defined null", "→", "mover", true, nil, true, nil},
		{"ordinary parent", "→", "mrow", false, nil, false, nil},
		{"embellished base", "→", "base", false, nil, false, nil},
		{"semantic inferred parent", "→", "inferred", false, nil, true, true},
		{"explicit row parent", "→", "explicit", false, nil, false, nil},
	} {
		t.Run(c.name, func(t *testing.T) {
			mark := token("mo", c.mark)
			base := token("mi", "x")
			if c.defined {
				mark.SetProperty("mathaccent", c.value)
			}
			switch c.context {
			case "base":
				node("mover", texAtom(mark, mml.TeXClassOrd), token("mi", "x"))
			case "inferred", "explicit":
				row := node("mrow", mark)
				row.Flags.Inferred = c.context == "inferred"
				row.Flags.NotParent = row.Flags.Inferred
				stack := node("mover", base, token("mi", "x"))
				stack.Children[1] = row
				row.Parent = stack
			case "munderover":
				node("munderover", base, mark, token("mi", "y"))
			default:
				node(c.context, base, mark)
			}
			setMathMLInheritance(mark, true)
			got, ok := mark.Property("mathaccent")
			if ok != c.wantDefined || !reflect.DeepEqual(got, c.want) {
				t.Fatalf("mathaccent %v/%v, want %v/%v", got, ok, c.want, c.wantDefined)
			}
		})
	}
}
