// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"encoding/json"
	"os"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestQuantityUnsupportedStarPublicReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/quantity_star_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit, PrimaryCaptureFreezeSHA256 string
		Cases                                        []struct {
			Name, TeX, SVG string
			Display        bool
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || fixture.PrimaryCaptureFreezeSHA256 != "04e27d5925fac731c4c63045e10480c3434b01039d27d08e216e2cd66a8d0448" || len(fixture.Cases) != 4 {
		t.Fatal("unbound Quantity references")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.SVG {
				t.Fatalf("complete primary SVG differs\ngot: %s\nwant: %s", got, c.SVG)
			}
		})
	}
}
