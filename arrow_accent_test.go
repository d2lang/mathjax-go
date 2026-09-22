package mathjax_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

type arrowQualification struct {
	Name, Kind, Issue, BaselineSVG, PrimarySVG, BaselineSHA256, PrimarySHA256 string
	Width, Height                                                             int
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
	if f.Base != "7101dc7e9bc1cfb830f91db9e23314c597c0d81d" || len(f.Cases) != 2 {
		t.Fatal("missing exact qualified controls")
	}
	expected := map[string]string{"ordinary-arrows-inline": "unchanged", "ordinary-arrows-display": "unchanged"}
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
				width, height = q.Width, q.Height
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
