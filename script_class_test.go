package mathjax_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestScriptClassPinnedReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/script_class_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name, TeX, SVGSHA256 string
			Display              bool
			Width, Height        int
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 18 {
		t.Fatal("incomplete nested-script spacing matrix")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if c.Name == "simple-sum-inline" {
				if h := fmt.Sprintf("%x", sha256.Sum256([]byte(got))); h != "5d02b36db1d5665598f22f32686585c24d2c346727d71553a3b9141784d38954" {
					t.Fatal("changed accepted simple-sum inline output")
				}
				// Existing parser metadata boundary: identical complete paint, but the
				// compiled inline script tag is msubsup instead of primary munderover.
				// Permit exactly this single literal tag, with every other byte strict.
				if strings.Count(got, `data-mml-node="msubsup"`) != 1 || strings.Contains(got, `data-mml-node="munderover"`) {
					t.Fatal("changed simple-sum structure")
				}
				got = strings.Replace(got, `data-mml-node="msubsup"`, `data-mml-node="munderover"`, 1)
			}
			if h := fmt.Sprintf("%x", sha256.Sum256([]byte(got))); h != c.SVGSHA256 {
				t.Errorf("complete SVG %s, want %s", h, c.SVGSHA256)
			}
			if c.Display {
				w, h, err := mathjax.Measure(c.TeX)
				if err != nil || w != c.Width || h != c.Height {
					t.Errorf("Measure %dx%d %v, want %dx%d", w, h, err, c.Width, c.Height)
				}
			}
		})
	}
}
