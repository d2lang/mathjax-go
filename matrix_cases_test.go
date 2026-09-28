// SPDX-License-Identifier: Apache-2.0
package mathjax_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	mathjax "github.com/d2lang/mathjax-go"
)

func TestMatrixCasesPublicReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/matrix_cases_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVG string
			Display        bool
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 246 {
		t.Fatal("unbound public cases command references")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			options := mathjax.DefaultOptions()
			options.Display = c.Display
			got, err := mathjax.RenderWithOptions(c.TeX, options)
			if err != nil {
				t.Fatal(err)
			}
			// The original error serializer emits a bare ampersand in this
			// attribute. Keep the frozen output intact and require Go's safe
			// XML escaping; the diagnostic and rendered text are unchanged.
			want := strings.Replace(c.SVG, `data-mjx-error="Misplaced &"`, `data-mjx-error="Misplaced &amp;"`, 1)
			if got != want {
				t.Fatalf("complete primary SVG differs\ngot: %s\nwant: %s", got, want)
			}
		})
	}
}
