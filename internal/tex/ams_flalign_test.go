// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"testing"
)

func TestFlalignOriginalOverflowDiagnostics(t *testing.T) {
	data, err := os.ReadFile("../../testdata/flalign_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name, TeX, SVG, Comparison string
			Display                    bool
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	diagnostic := regexp.MustCompile(`data-mjx-error="(Extra & in row of (?:xalignat\*?|xxalignat))"`)
	count := 0
	for _, c := range fixture.Cases {
		if c.Comparison != "safeXML" {
			continue
		}
		count++
		t.Run(c.Name, func(t *testing.T) {
			match := diagnostic.FindStringSubmatch(c.SVG)
			if len(match) != 2 {
				t.Fatal("missing raw primary diagnostic")
			}
			p := &parser{source: c.TeX, display: c.Display, state: newParseState()}
			_, _, err := p.parseRow(0, false)
			var got *Error
			if !errors.As(err, &got) || got.ID != "XalignOverflow" || got.Message != match[1] {
				t.Fatalf("FlalignItem.EndEntry: got %#v; want XalignOverflow, %q", err, match[1])
			}
		})
	}
	if count != 74 {
		t.Fatal("unexpected primary diagnostic count", count)
	}
}
