package mathjax

import (
	"encoding/json"
	"os"
	"testing"
)

func TestResponsiveMinWidthPublic(t *testing.T) {
	var fixture struct {
		Cases []struct {
			Name, TeX, SVG, Acceptance string
			Display                    bool
		}
	}
	b, err := os.ReadFile("testdata/responsive_minwidth_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 30 {
		t.Fatal("fixture count", len(fixture.Cases))
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			o := DefaultOptions()
			o.Display = c.Display
			got, err := RenderWithOptions(c.TeX, o)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.SVG {
				t.Fatalf("%s complete SVG mismatch\ngot: %s\nwant: %s", c.Acceptance, got, c.SVG)
			}
		})
	}
}
