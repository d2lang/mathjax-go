// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestGetCSLineEndingPrimaryMethod(t *testing.T) {
	data, err := os.ReadFile("testdata/getcs_line_endings_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Source, Value, Remaining string
			ConsumedBytes            int
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 74 {
		t.Fatal("unbound GetCS source observations")
	}
	seen := make(map[string]bool)
	for i, c := range fixture.Cases {
		if seen[c.Source] {
			t.Fatal("duplicate GetCS source observation", c.Source)
		}
		seen[c.Source] = true
		t.Run(fmt.Sprintf("case-%03d", i), func(t *testing.T) {
			p := &parser{source: c.Source}
			value := p.readControlSequence()
			// Compare consumed UTF-8 bytes and remaining source. The original
			// UTF-16 cursor is retained in the fixture, including its virtual
			// one-unit advance at EOF; Go keeps a bounded physical byte cursor.
			if value != c.Value || p.pos != c.ConsumedBytes || p.source[p.pos:] != c.Remaining {
				t.Fatalf("GetCS differs: value=%q bytes=%d remaining=%q; want value=%q bytes=%d remaining=%q", value, p.pos, p.source[p.pos:], c.Value, c.ConsumedBytes, c.Remaining)
			}
		})
	}
}
