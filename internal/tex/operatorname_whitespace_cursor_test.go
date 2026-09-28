// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"encoding/json"
	"os"
	"testing"
)

type operatorNameCursorReference struct {
	Name, Source, Remaining string
	ConsumedBytes           int
}

func operatorNameOriginalCursors(t *testing.T) map[string]operatorNameCursorReference {
	t.Helper()
	data, err := os.ReadFile("testdata/operatorname_whitespace_cursors.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []operatorNameCursorReference
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 32 {
		t.Fatal("unbound original handler cursor observations")
	}
	refs := make(map[string]operatorNameCursorReference)
	for _, c := range fixture.Cases {
		if _, duplicate := refs[c.Name]; duplicate {
			t.Fatal("duplicate cursor observation", c.Name)
		}
		if c.ConsumedBytes < 0 || c.ConsumedBytes > len(c.Source) || c.Source[c.ConsumedBytes:] != c.Remaining {
			t.Fatal("invalid original cursor", c.Name)
		}
		refs[c.Name] = c
	}
	return refs
}

func TestOperatorNameWhitespaceOriginalCursor(t *testing.T) {
	for _, c := range operatorNameOriginalCursors(t) {
		t.Run(c.Name, func(t *testing.T) {
			p := &parser{source: c.Source, state: newParseState()}
			_, handled, err := p.baseAMSCommand("operatorname")
			if !handled || p.pos != c.ConsumedBytes || p.source[p.pos:] != c.Remaining {
				t.Fatalf("original cursor differs: position %d, want %d", p.pos, c.ConsumedBytes)
			}
			if c.Name == "missing-argument" {
				macroReferenceError(t, err, &Error{ID: "MissingArgFor", Message: `Missing argument for \operatorname`})
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}
