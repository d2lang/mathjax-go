// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"encoding/json"
	"os"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestOperatorNameChildPublicReferences(t *testing.T) {
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVG, OriginalRecordSHA256 string
			Display                              bool
			Width, Height                        int
		}
	}
	data, err := os.ReadFile("testdata/operator_name_child_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 8 {
		t.Fatal("unbound operator-name child references")
	}
	displayCases := 0
	for _, c := range fixture.Cases {
		if c.Display {
			displayCases++
		}
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
			if c.Display {
				width, height, err := mathjax.Measure(c.TeX)
				if err != nil || width != c.Width || height != c.Height {
					t.Errorf("measure=%dx%d, %v; want=%dx%d", width, height, err, c.Width, c.Height)
				}
			}
		})
	}
	if displayCases != 4 {
		t.Fatal("changed display measurement inventory")
	}
}
