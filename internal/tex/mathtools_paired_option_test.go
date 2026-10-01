// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestMathtoolsPairedOptionOriginalStorage(t *testing.T) {
	data, err := os.ReadFile("testdata/mathtools_paired_option_observations.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit, SetOptionsSource string
		Cases                              []struct {
			Name, Raw, PairedType string
			PairedValue           any
			Error                 *struct{ ID, Message string }
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 19 || !strings.Contains(fixture.SetOptionsSource, `"pariedDelimiters"`) {
		t.Fatal("unbound original SetOptions observations")
	}
	seen := make(map[string]bool)
	for _, c := range fixture.Cases {
		if c.Name == "" || seen[c.Name] {
			t.Fatal("invalid SetOptions observation", c.Name)
		}
		seen[c.Name] = true
		t.Run(c.Name, func(t *testing.T) {
			p := &parser{source: "{" + c.Raw + "}", state: newParseState(), display: true}
			err := p.mathtoolsSetOptions("mathtoolsset")
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
			if c.PairedType != "boolean" && c.PairedType != "string" {
				t.Fatal("unexpected original stored type", c.PairedType)
			}
			stored, ok := p.state.macros[mathtoolsOptionPrefix+"pairedDelimiters"]
			if !ok || stored.body != fmt.Sprint(c.PairedValue) {
				t.Fatalf("stored pairedDelimiters=%q present=%v want %q", stored.body, ok, fmt.Sprint(c.PairedValue))
			}
			if len(p.state.pairedDelimiters) != 0 {
				t.Fatal("the setter rebuilt the delimiter registration map")
			}
		})
	}
}
