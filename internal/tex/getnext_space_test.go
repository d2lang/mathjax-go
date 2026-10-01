// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"testing"
	"unicode/utf8"
)

func TestGetNextWhitespaceOriginalCursor(t *testing.T) {
	file, err := os.Open("../../testdata/getnext_space_mathjax_3_2_2.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	var fixture struct {
		MathjaxGitCommit string
		SourceBindings   struct {
			WhitespaceObservations []struct {
				Point, Cursor int
				IsSpace       bool
				Next          string
			}
		}
	}
	if err := json.NewDecoder(reader).Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.SourceBindings.WhitespaceObservations) != 28 {
		t.Fatal("unbound original GetNext cursor observations")
	}
	seen, spaces := make(map[int]bool), 0
	for _, c := range fixture.SourceBindings.WhitespaceObservations {
		if seen[c.Point] || c.Cursor < 0 || c.Cursor > 1 || c.IsSpace != (c.Cursor == 1) {
			t.Fatal("invalid original cursor observation", c)
		}
		seen[c.Point] = true
		if c.IsSpace {
			spaces++
		}
		p := &parser{source: string(rune(c.Point)) + "x"}
		p.skipNextSpaces()
		want := 0
		if c.Cursor == 1 {
			want = utf8.RuneLen(rune(c.Point))
		}
		if p.pos != want || string(p.peekRune()) != c.Next {
			t.Fatalf("U+%04X: cursor=%d next=%q want cursor=%d next=%q", c.Point, p.pos, string(p.peekRune()), want, c.Next)
		}
	}
	if spaces != 25 || !seen[133] || !seen[6158] || !seen[8203] {
		t.Fatal("incomplete original whitespace membership")
	}
}
