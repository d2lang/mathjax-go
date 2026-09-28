// SPDX-License-Identifier: Apache-2.0
package svg

import (
	"encoding/json"
	"math"
	"os"
	"strconv"
	"testing"
)

func TestEmptyTableEqualRowOriginalMethod(t *testing.T) {
	data, err := os.ReadFile("testdata/empty_table_equalrows_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit, Method string
		Cases                    []struct {
			H, D     []string
			Original string
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || fixture.Method != "CommonMtable.getEqualRowHeight" || len(fixture.Cases) != 20 {
		t.Fatal("unbound original equal-row observations")
	}
	parse := func(value string) float64 {
		t.Helper()
		result, err := strconv.ParseFloat(value, 64)
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	for i, c := range fixture.Cases {
		if len(c.H) != len(c.D) {
			t.Fatalf("case %d has mismatched source row dimensions", i)
		}
		dimensions := tableDimensions{W: []float64{}}
		for j := range c.H {
			dimensions.H = append(dimensions.H, parse(c.H[j]))
			dimensions.D = append(dimensions.D, parse(c.D[j]))
		}
		layout := tableLayout{data: dimensions}
		got, want := layout.equalRowHeight(), parse(c.Original)
		if math.IsNaN(want) {
			if !math.IsNaN(got) {
				t.Errorf("case %d: got %v, want NaN", i, got)
			}
		} else if got != want || (want == 0 && math.Signbit(got) != math.Signbit(want)) {
			t.Errorf("case %d: got %v, want %s", i, got, c.Original)
		}
	}
}
