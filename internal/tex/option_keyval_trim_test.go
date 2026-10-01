// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestEmpheqOptionKeyvalOriginalWhitespace(t *testing.T) {
	data, err := os.ReadFile("testdata/option_keyval_trim_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit, KeyvalMethod string
		Cases                          []struct {
			Name, Raw string
			Options   map[string]any
			Error     *struct{ ID, Message string }
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || fixture.KeyvalMethod == "" || len(fixture.Cases) != 16 {
		t.Fatal("unbound original keyval whitespace observations")
	}
	seen := make(map[string]bool)
	for _, c := range fixture.Cases {
		if c.Name == "" || seen[c.Name] {
			t.Fatal("invalid keyval observation", c.Name)
		}
		seen[c.Name] = true
		t.Run(c.Name, func(t *testing.T) {
			options, err := empheqSplitOptions(c.Raw)
			if c.Error != nil {
				failure, ok := err.(*Error)
				if !ok || failure.ID != c.Error.ID || failure.Message != c.Error.Message {
					t.Fatalf("diagnostic=%v want %#v", err, c.Error)
				}
				return
			}
			if err != nil || !reflect.DeepEqual(options, c.Options) {
				t.Fatalf("options=%#v error=%v want %#v", options, err, c.Options)
			}
		})
	}
}
