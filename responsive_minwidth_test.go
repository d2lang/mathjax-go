package mathjax

import (
	"encoding/json"
	"os"
	"testing"
)

func TestResponsiveMinWidthPublic(t *testing.T) {
	var fixture struct {
		Cases []struct {
			Name, TeX, SVG, PrimarySVG, Acceptance string
			Display                                bool
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
	promoted := 0
	for _, c := range fixture.Cases {
		want, acceptance := c.SVG, c.Acceptance
		if c.Name == "public-two-tables-inline" || c.Name == "public-two-tables-reversed-inline" {
			if c.PrimarySVG == "" {
				t.Fatal("missing original equation-nesting error", c.Name)
			}
			want, acceptance = c.PrimarySVG, "original equation-nesting error"
			promoted++
		}
		t.Run(c.Name, func(t *testing.T) {
			o := DefaultOptions()
			o.Display = c.Display
			got, err := RenderWithOptions(c.TeX, o)
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Fatalf("%s complete SVG mismatch\ngot: %s\nwant: %s", acceptance, got, want)
			}
		})
	}
	if promoted != 2 {
		t.Fatal("changed equation-nesting promotion inventory")
	}
}
