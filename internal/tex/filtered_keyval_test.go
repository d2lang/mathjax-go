// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/tex/extensions/enclose"
)

func TestFilteredKeyvalOriginalReferences(t *testing.T) {
	data, err := os.ReadFile("testdata/filtered_keyval_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit, KeyvalMethod string
		AllowedKeys                    []string
		Cases                          []struct {
			Name, Raw               string
			Allowed, ErrorOnUnknown bool
			Options                 map[string]any
			Keys                    []string
			Error                   *struct{ ID, Message string }
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || fixture.KeyvalMethod == "" ||
		len(fixture.Cases) != 90 || !reflect.DeepEqual(fixture.AllowedKeys, enclose.AllowedOptions) {
		t.Fatal("unbound original filtered keyval observations")
	}
	seen := make(map[string]bool)
	successes, failures := 0, 0
	for _, c := range fixture.Cases {
		if c.Name == "" || seen[c.Name] {
			t.Fatal("invalid filtered keyval observation", c.Name)
		}
		seen[c.Name] = true
		if c.Error == nil {
			successes++
		} else {
			failures++
		}
		t.Run(c.Name, func(t *testing.T) {
			var allowed func(string) bool
			if c.Allowed {
				allowed = func(key string) bool {
					for _, name := range enclose.AllowedOptions {
						if key == name {
							return true
						}
					}
					return false
				}
			}
			options, err := parseUtilKeyvalOptions(c.Raw, allowed, c.ErrorOnUnknown)
			if c.Error != nil {
				failure, ok := err.(*Error)
				if !ok || failure.ID != c.Error.ID || failure.Message != c.Error.Message {
					t.Fatalf("diagnostic=%v want %#v", err, c.Error)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			got := make(map[string]any, options.Len())
			options.Range(func(key string, value any) bool { got[key] = value; return true })
			if !reflect.DeepEqual(got, c.Options) || !reflect.DeepEqual(options.JavaScriptKeys(), c.Keys) {
				t.Fatalf("options=%#v keys=%#v want options=%#v keys=%#v", got, options.JavaScriptKeys(), c.Options, c.Keys)
			}
			if c.Allowed && !c.ErrorOnUnknown {
				attributes, err := keyvalOptions(c.Raw, enclose.AllowedOptions)
				if err != nil || !reflect.DeepEqual(attributes, c.Options) {
					t.Fatalf("Cancel/Enclose options=%#v error=%v want %#v", attributes, err, c.Options)
				}
			}
		})
	}
	if successes != 66 || failures != 24 {
		t.Fatalf("original keyval inventory = %d successes, %d errors; want 66, 24", successes, failures)
	}
}
