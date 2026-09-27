// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package tex

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
)

func TestScriptScaleRetainedMultiscriptSVG(t *testing.T) {
	var fixture struct {
		MathjaxGitCommit, ExpectedSVG, SVGSHA256 string
		Spec                                     struct {
			Display bool
			Tree    json.RawMessage
		}
	}
	data, err := os.ReadFile("../../testdata/script_scaling_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || fmt.Sprintf("%x", sha256.Sum256([]byte(fixture.ExpectedSVG))) != fixture.SVGSHA256 {
		t.Fatal("unbound original multiscript SVG")
	}
	root := idAnchorBuild(fixture.Spec.Tree, fixture.Spec.Display)
	options := pipeline.DefaultOptions()
	options.Display = fixture.Spec.Display
	got, err := svg.NewTypesetter().Typeset(root, options)
	if err != nil {
		t.Fatal(err)
	}
	if got != fixture.ExpectedSVG {
		t.Fatalf("retained original multiscript SVG differs\ngot: %s\nwant: %s", got, fixture.ExpectedSVG)
	}
}
