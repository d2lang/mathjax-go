package mathjax_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
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
