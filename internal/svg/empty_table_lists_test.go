// SPDX-License-Identifier: Apache-2.0
package svg

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestEmptyTableListOriginalMethods(t *testing.T) {
	data, err := os.ReadFile("testdata/empty_table_lists_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Methods          []struct {
			Attribute, Method, Fallback, Value string
			Count                              int
			Original                           struct{ Values, Result []string }
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Methods) != 80 {
		t.Fatal("unbound original table-list observations")
	}
	for _, c := range fixture.Methods {
		got := repeatTableField(c.Original.Values, c.Count-1, c.Fallback)
		if !reflect.DeepEqual(got, c.Original.Result) {
			t.Errorf("%s(%q), count %d: got %q, want %q", c.Method, c.Value, c.Count, got, c.Original.Result)
		}
	}
}
