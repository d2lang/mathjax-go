package mathjax_test

import (
	"crypto/sha256"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

type arrowQualification struct {
	Name, Kind, Issue, BaselineSVG, PrimarySVG, BaselineSHA256, PrimarySHA256 string
	Width, Height                                                             int
}

// Only the two named bar fixtures permit the primary's one transparent ORD
// TeXAtom wrapper to be absent. Every other byte, including both transforms,
// root attributes, child order and path data must remain the primary's bytes.
func primaryBarWithoutAtom(svg string) (string, error) {
	type node struct {
		XMLName  xml.Name
		Attr     []xml.Attr `xml:",any,attr"`
		Text     string     `xml:",chardata"`
		Children []node     `xml:",any"`
	}
	var root node
	if err := xml.Unmarshal([]byte(svg), &root); err != nil {
		return "", err
	}
	one := func(n node) (node, error) {
		if len(n.Children) != 1 || n.Text != "" {
			return node{}, fmt.Errorf("not one exact child")
		}
		return n.Children[0], nil
	}
	outer, err := one(root)
	if err != nil {
		return "", err
	}
	math, err := one(outer)
	if err != nil {
		return "", err
	}
	atom, err := one(math)
	if err != nil {
		return "", err
	}
	mover, err := one(atom)
	if err != nil {
		return "", err
	}
	attrs := func(n node, want map[string]string) bool {
		if len(n.Attr) != len(want) {
			return false
		}
		for _, a := range n.Attr {
			if a.Name.Space != "" || want[a.Name.Local] != a.Value {
				return false
			}
		}
		return true
	}
	if root.XMLName.Local != "svg" || outer.XMLName.Local != "g" || math.XMLName.Local != "g" || atom.XMLName.Local != "g" || mover.XMLName.Local != "g" || !attrs(math, map[string]string{"data-mml-node": "math"}) || !attrs(atom, map[string]string{"data-mml-node": "TeXAtom", "data-mjx-texclass": "ORD"}) || !attrs(mover, map[string]string{"data-mml-node": "mover"}) {
		return "", fmt.Errorf("unexpected bar wrapper structure or attributes")
	}
	for _, n := range []node{root, outer, math, atom, mover} {
		if n.XMLName.Space != "http://www.w3.org/2000/svg" {
			return "", fmt.Errorf("wrong namespace")
		}
	}
	const opening = `<g data-mml-node="TeXAtom" data-mjx-texclass="ORD">`
	const parent = `<g data-mml-node="math">`
	const suffix = `</g></g></g></svg>`
	at := strings.Index(svg, opening)
	close := strings.LastIndex(svg, suffix)
	if strings.Count(svg, opening) != 1 || at != strings.Index(svg, parent)+len(parent) || !strings.HasSuffix(svg, suffix) || close <= at {
		return "", fmt.Errorf("unexpected exact wrapper spelling")
	}
	return svg[:at] + svg[at+len(opening):close] + svg[close+len("</g>"):], nil
}

func arrowQualifications(t *testing.T) map[string]arrowQualification {
	t.Helper()
	data, err := os.ReadFile("testdata/arrow_accent_qualifications.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Base  string
		Cases []arrowQualification
	}
	if err = json.Unmarshal(data, &f); err != nil {
		t.Fatal(err)
	}
	if f.Base != "7101dc7e9bc1cfb830f91db9e23314c597c0d81d" || len(f.Cases) != 4 {
		t.Fatal("missing exact qualified controls")
	}
	expected := map[string]string{}
	for _, name := range []string{"ordinary-arrows", "bar-control"} {
		for _, mode := range []string{"inline", "display"} {
			kind := "unchanged"
			if name == "bar-control" {
				kind = "bar-wrapper"
			}
			expected[name+"-"+mode] = kind
		}
	}
	result := map[string]arrowQualification{}
	for _, q := range f.Cases {
		if expected[q.Name] != q.Kind || result[q.Name].Name != "" {
			t.Fatal("unexpected/duplicate qualification", q.Name)
		}
		if fmt.Sprintf("%x", sha256.Sum256([]byte(q.BaselineSVG))) != q.BaselineSHA256 || fmt.Sprintf("%x", sha256.Sum256([]byte(q.PrimarySVG))) != q.PrimarySHA256 || q.BaselineSHA256 == q.PrimarySHA256 {
			t.Fatal("invalid frozen qualification", q.Name)
		}
		result[q.Name] = q
	}
	return result
}

func TestArrowAccentPinnedReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/arrow_accent_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Cases []struct {
			Name, TeX, SVGSHA256 string
			Display              bool
			Width, Height        int
		}
	}
	if err = json.Unmarshal(data, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Cases) != 48 {
		t.Fatal("missing arrow/accent cases")
	}
	qualifications := arrowQualifications(t)
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			actual, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			width, height := c.Width, c.Height
			if q, ok := qualifications[c.Name]; ok {
				if q.PrimarySHA256 != c.SVGSHA256 {
					t.Fatal("primary reference changed")
				}
				expected := q.BaselineSVG
				if q.Kind == "bar-wrapper" {
					expected, err = primaryBarWithoutAtom(q.PrimarySVG)
					if err != nil {
						t.Fatal(err)
					}
				} else {
					width, height = q.Width, q.Height
				}
				if actual != expected {
					t.Error("complete qualified SVG differs from exact permitted bytes")
				}
			} else if got := fmt.Sprintf("%x", sha256.Sum256([]byte(actual))); got != c.SVGSHA256 {
				t.Errorf("complete SVG %s, want %s", got, c.SVGSHA256)
			}
			if c.Display {
				w, h, err := mathjax.Measure(c.TeX)
				if err != nil || w != width || h != height {
					t.Errorf("Measure %dx%d, %v; want %dx%d", w, h, err, width, height)
				}
			}
		})
	}
}

func TestBarWrapperQualificationRejectsOtherStructure(t *testing.T) {
	q := arrowQualifications(t)["bar-control-display"]
	if _, err := primaryBarWithoutAtom(q.PrimarySVG); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{
		strings.Replace(q.PrimarySVG, `data-mjx-texclass="ORD"`, `data-mjx-texclass="ORD" opacity="0.5"`, 1),
		strings.Replace(q.PrimarySVG, `data-mjx-texclass="ORD"`, `data-mjx-texclass="OP"`, 1),
		strings.Replace(q.PrimarySVG, `data-mml-node="mover"`, `data-mml-node="munder"`, 1),
		strings.Replace(q.PrimarySVG, `<g data-mml-node="mover">`, `<g/><g data-mml-node="mover">`, 1),
	} {
		if _, err := primaryBarWithoutAtom(bad); err == nil {
			t.Fatal("unexpected wrapper structure accepted")
		}
	}
}
