// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestOptionKeyvalOriginalObjectProperties(t *testing.T) {
	data, err := os.ReadFile("testdata/option_keyval_object_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit, KeyvalMethod string
		Cases                          []struct {
			Name, Raw string
			Validated bool
			Options   map[string]any
			Keys      []string
			Error     *struct{ ID, Message string }
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || fixture.KeyvalMethod == "" || len(fixture.Cases) != 90 {
		t.Fatal("unbound original keyval object observations")
	}
	seen := make(map[string]bool)
	successes, failures := 0, 0
	for _, c := range fixture.Cases {
		if c.Name == "" || seen[c.Name] {
			t.Fatal("invalid keyval observation", c.Name)
		}
		seen[c.Name] = true
		if c.Error == nil {
			successes++
		} else {
			failures++
		}
		t.Run(c.Name, func(t *testing.T) {
			var allowed func(string) bool
			if c.Validated {
				allowed = func(key string) bool { return key == "left" || key == "right" }
			}
			options, err := parseUtilKeyvalOptions(c.Raw, allowed)
			if c.Error != nil {
				failure, ok := err.(*Error)
				if !ok || failure.ID != c.Error.ID || failure.Message != c.Error.Message {
					t.Fatalf("diagnostic=%v want %#v", err, c.Error)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				got := make(map[string]any, options.Len())
				options.Range(func(key string, value any) bool { got[key] = value; return true })
				if !reflect.DeepEqual(got, c.Options) || !reflect.DeepEqual(options.JavaScriptKeys(), c.Keys) {
					t.Fatalf("options=%#v keys=%#v want options=%#v keys=%#v", got, options.JavaScriptKeys(), c.Options, c.Keys)
				}
			}
			if c.Validated {
				got, err := empheqSplitOptions(c.Raw)
				if c.Error != nil {
					failure, ok := err.(*Error)
					if !ok || failure.ID != c.Error.ID || failure.Message != c.Error.Message {
						t.Fatalf("Empheq diagnostic=%v want %#v", err, c.Error)
					}
				} else if err != nil || !reflect.DeepEqual(got, c.Options) {
					t.Fatalf("Empheq options=%#v error=%v want %#v", got, err, c.Options)
				}
			}
		})
	}
	if successes != 58 || failures != 32 {
		t.Fatalf("original helper inventory = %d successes, %d errors; want 58 successes, 32 errors", successes, failures)
	}
}
