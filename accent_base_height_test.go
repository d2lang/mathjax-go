package mathjax_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestAccentBaseHeightPinnedReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/accent_base_height_mathjax_3_2_2.json")
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
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 40 {
		t.Fatal("missing accent height or unchanged under/accent controls")
	}
	residualData, err := os.ReadFile("testdata/accent_base_height_unchanged_underline.json")
	if err != nil {
		t.Fatal(err)
	}
	var residuals struct {
		Base  string
		Cases []struct {
			Name, Issue, SVG, BaselineSVGSHA256, PrimarySVGSHA256 string
			Width, Height                                         int
		}
	}
	if err := json.Unmarshal(residualData, &residuals); err != nil {
		t.Fatal(err)
	}
	if residuals.Base != "11dd952e8211c03dc5eb989fc3577df00cab858b" || len(residuals.Cases) != 2 {
		t.Fatal("missing unchanged underline baseline")
	}
	known := map[string]bool{"underline-control-inline": true, "underline-control-display": true}
	for _, r := range residuals.Cases {
		if !known[r.Name] || r.Issue != "D048" {
			t.Fatal("unknown or duplicate residual", r.Name)
		}
		delete(known, r.Name)
		if fmt.Sprintf("%x", sha256.Sum256([]byte(r.SVG))) != r.BaselineSVGSHA256 || r.BaselineSVGSHA256 == r.PrimarySVGSHA256 {
			t.Fatal("invalid unchanged baseline", r.Name)
		}
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			svg, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			knownResidual := false
			width, height := c.Width, c.Height
			for _, r := range residuals.Cases {
				if r.Name == c.Name {
					knownResidual = true
					if r.PrimarySVGSHA256 != c.SVGSHA256 {
						t.Fatal("primary reference changed")
					}
					if svg != r.SVG {
						t.Error("underline differs from immutable baseline SVG")
					}
					width, height = r.Width, r.Height
				}
			}
			if !knownResidual {
				if got := fmt.Sprintf("%x", sha256.Sum256([]byte(svg))); got != c.SVGSHA256 {
					t.Errorf("complete SVG %s, want %s", got, c.SVGSHA256)
				}
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
