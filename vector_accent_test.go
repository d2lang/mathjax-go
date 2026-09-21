package mathjax_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestVectorAccentPinnedReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/vector_accent_mathjax_3_2_2.json")
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
	if len(fixture.Cases) != 44 {
		t.Fatal("missing vector or unchanged wide-accent controls")
	}
	residualData, err := os.ReadFile("testdata/vector_accent_unchanged_residuals.json")
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
	if residuals.Base != "c22e62130eaa42e523a3b0bc8cee66ba19c45c5c" || len(residuals.Cases) != 2 {
		t.Fatal("missing immutable known-residual baseline")
	}
	known := map[string]bool{"combined-brace-inline": true, "combined-brace-display": true}
	for _, r := range residuals.Cases {
		if !known[r.Name] {
			t.Fatal("unexpected/duplicate residual", r.Name)
		}
		delete(known, r.Name)
		if fmt.Sprintf("%x", sha256.Sum256([]byte(r.SVG))) != r.BaselineSVGSHA256 || r.BaselineSVGSHA256 == r.PrimarySVGSHA256 {
			t.Fatal("invalid known residual", r.Name)
		}
		if r.Issue != "D047" {
			t.Fatal("wrong residual issue", r.Name)
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
						t.Error("known residual changed from exact baseline SVG")
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
