// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestGetBracketsWhitespacePrimaryMethod(t *testing.T) {
	data, err := os.ReadFile("testdata/getbrackets_whitespace_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Methods          map[string]string
		Cases            []struct {
			Name, Source, Remaining string
			DefaultValue, Value     *string
			Present                 bool
			CursorBytes             int
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 198 ||
		!strings.Contains(fixture.Methods["nextIsSpace"], `/\s/`) || fixture.Methods["GetBrackets"] == "" {
		t.Fatal("unbound original GetBrackets method observations")
	}
	for _, c := range fixture.Cases {
		name := c.Name + "-none"
		if c.DefaultValue != nil {
			name = c.Name + "-default"
		}
		t.Run(name, func(t *testing.T) {
			p := &parser{source: c.Source}
			got, present, err := p.readBrackets(c.DefaultValue)
			want := ""
			if c.Value != nil {
				want = *c.Value
			}
			if err != nil || got != want || present != c.Present || p.pos != c.CursorBytes || p.source[p.pos:] != c.Remaining {
				t.Fatalf("original method result differs: value=%q, present=%v, cursor=%d, remaining=%q, err=%v; want value=%q, present=%v, cursor=%d, remaining=%q", got, present, p.pos, p.source[p.pos:], err, want, c.Present, c.CursorBytes, c.Remaining)
			}
		})
	}
}
