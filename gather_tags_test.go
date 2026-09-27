// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"encoding/json"
	"os"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

// The frozen original outputs cover gather and gather* in both display modes:
// row-local tags/labels, suppression and replacement, duplicate errors, and the
// shared equation-nesting guard, including the Mathtools special-row route.
// Compare complete SVGs, including tag anchors and their row placement.
func TestGatherTagsPublicReferences(t *testing.T) {
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVG, OriginalRecordSHA256 string
			Display                              bool
		}
	}
	data, err := os.ReadFile("testdata/gather_tags_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 67 {
		t.Fatal("unbound gather-tag references")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			if len(c.OriginalRecordSHA256) != 64 || c.SVG == "" {
				t.Fatal("missing original reference")
			}
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			svg, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			if svg != c.SVG {
				t.Errorf("whole original SVG differs\ngot: %s\nwant: %s", svg, c.SVG)
			}
		})
	}
}
