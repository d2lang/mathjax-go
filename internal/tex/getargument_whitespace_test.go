// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestGetArgumentWhitespacePrimaryMethod(t *testing.T) {
	data, err := os.ReadFile("testdata/getargument_whitespace_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Methods          map[string]string
		Cases            []struct {
			Source, Remaining string
			NoneOK            bool
			Value             *string
			Error             *struct{ ID, Message string }
			CursorBytes       int
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 308 ||
		!strings.Contains(fixture.Methods["nextIsSpace"], `/\s/`) || !strings.Contains(fixture.Methods["GetArgument"], "GetNext") {
		t.Fatal("unbound original GetArgument method observations")
	}
	seen := make(map[struct {
		source string
		noneOK bool
	}]bool)
	for i, c := range fixture.Cases {
		key := struct {
			source string
			noneOK bool
		}{c.Source, c.NoneOK}
		if seen[key] {
			t.Fatal("duplicate original method observation", key)
		}
		seen[key] = true
		t.Run(fmt.Sprintf("case-%03d", i), func(t *testing.T) {
			p := &parser{source: c.Source}
			got, _, err := p.readArgument("sample", c.NoneOK)
			if c.Error != nil {
				gotError, ok := err.(*Error)
				if !ok || gotError.ID != c.Error.ID || gotError.Message != c.Error.Message {
					t.Fatalf("original error differs: got %v, want %+v", err, *c.Error)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			want := ""
			if c.Value != nil {
				want = *c.Value
			}
			if got != want || p.pos != c.CursorBytes || p.source[p.pos:] != c.Remaining {
				t.Fatalf("original argument result differs: value=%q cursor=%d remaining=%q; want value=%q cursor=%d remaining=%q", got, p.pos, p.source[p.pos:], want, c.CursorBytes, c.Remaining)
			}
		})
	}
}
