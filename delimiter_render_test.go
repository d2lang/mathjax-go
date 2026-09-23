package mathjax

import (
	"encoding/json"
	"os"
	"testing"
)

func TestDelimiterRenderMatchesPrimary(t *testing.T) {
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVG string
			Display        bool
		}
	}
	data, err := os.ReadFile("testdata/delimiter_svg_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 46 {
		t.Fatal("incomplete pinned primary fixture")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			options := DefaultOptions()
			options.Display = c.Display
			got, err := RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.SVG {
				index, actual, expected := firstSVGDiff(got, c.SVG)
				t.Fatalf("complete primary SVG differs at %d: got %q, want %q", index, actual, expected)
			}
		})
	}
}
