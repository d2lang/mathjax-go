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

// This is a constructed MathML witness, not an authored TeX/Studio route.
func TestBevelledFractionOriginalSVG(t *testing.T) {
	var fixture struct {
		MathjaxGitCommit, OriginalRecordSHA256, PassiveRecordSHA256 string
		SVG, SVGSHA256                                              string
		Spec                                                        struct {
			Display bool
			Tree    json.RawMessage
		}
	}
	data, err := os.ReadFile("testdata/bevelled_fraction_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" ||
		fixture.OriginalRecordSHA256 != "ae75065a86497274618d5681b6bc1d07cde9c3a07cb5a30e851e7c5c24fe0fdb" ||
		fixture.PassiveRecordSHA256 != "b08243a227ce9833d653a4da5e2b09b6fc72ec1181c3f670a36e598c9f915fe5" ||
		fmt.Sprintf("%x", sha256.Sum256([]byte(fixture.SVG))) != fixture.SVGSHA256 {
		t.Fatal("unbound original bevelled fraction reference")
	}
	root := idAnchorBuild(fixture.Spec.Tree, fixture.Spec.Display)
	options := pipeline.DefaultOptions()
	options.Display = fixture.Spec.Display
	got, err := svg.NewTypesetter().Typeset(root, options)
	if err != nil {
		t.Fatal(err)
	}
	if got != fixture.SVG {
		t.Errorf("whole original bevelled fraction SVG differs\ngot: %s\nwant: %s", got, fixture.SVG)
	}
}
