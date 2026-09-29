// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestEvalGetStarOriginalMethod(t *testing.T) {
	data, err := os.ReadFile("testdata/eval_getstar_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Source, Remaining string
			Value             bool
			ConsumedBytes     int
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 181 {
		t.Fatal("unbound original GetStar observations")
	}
	seen := map[string]bool{}
	for i, c := range fixture.Cases {
		if seen[c.Source] {
			t.Fatal("duplicate GetStar observation", c.Source)
		}
		seen[c.Source] = true
		t.Run(fmt.Sprintf("source-%03d", i), func(t *testing.T) {
			p := &parser{source: c.Source}
			value := p.readStar()
			if value != c.Value || p.pos != c.ConsumedBytes || p.source[p.pos:] != c.Remaining {
				t.Fatalf("GetStar value=%v bytes=%d remaining=%q; want %v bytes=%d remaining=%q", value, p.pos, p.source[p.pos:], c.Value, c.ConsumedBytes, c.Remaining)
			}
		})
	}
}
