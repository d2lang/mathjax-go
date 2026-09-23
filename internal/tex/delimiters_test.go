package tex

import (
	"encoding/json"
	"os"
	"testing"
)

func TestDelimiterReadersMatchPrimary(t *testing.T) {
	var fixture struct {
		Readers []struct {
			Name, Source, Command, CursorBoundary   string
			Cursor, PrimaryCursor, RetainedGoCursor int
			BraceOK                                 bool
			Value                                   *string
			Error                                   *Error
		}
		Converters []struct {
			Name, Raw, Trimmed string
			Registered         bool
			Character          *string
		}
	}
	data, err := os.ReadFile("testdata/delimiter_readers_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Readers) != 65 || len(fixture.Converters) != 16 {
		t.Fatal("incomplete primary fixture")
	}
	for _, c := range fixture.Readers {
		t.Run("reader/"+c.Name, func(t *testing.T) {
			p := &parser{source: c.Source, pos: c.Cursor, state: newParseState()}
			value, err := p.readDelimiter(c.Command, c.BraceOK)
			if c.Error != nil {
				got, ok := err.(*Error)
				if !ok || got.ID != c.Error.ID || got.Message != c.Error.Message {
					t.Fatalf("error = %#v, want %#v", err, c.Error)
				}
				if c.Value != nil {
					t.Fatal("error fixture unexpectedly has a value")
				}
			} else if err != nil || c.Value == nil || value != *c.Value {
				t.Fatalf("delimiter = %q, %v; primary = %v", value, err, c.Value)
			}
			wantCursor := c.PrimaryCursor
			// The shared Go scanners retain two pre-existing private cursor
			// boundaries after errors: dangling control-sequence EOF and an
			// unterminated argument. This reader change does not rewrite them.
			if c.CursorBoundary != "" {
				if c.Name != "dangling-backslash" && c.Name != "unclosed-brace" {
					t.Fatal("unknown cursor boundary")
				}
				if c.PrimaryCursor == c.RetainedGoCursor {
					t.Fatal("missing retained boundary")
				}
				wantCursor = c.RetainedGoCursor
			}
			if p.pos != wantCursor || p.source != c.Source || p.state.macroCount != 0 {
				t.Fatalf("reader state changed: cursor %d (want %d), source %q, count %d", p.pos, wantCursor, p.source, p.state.macroCount)
			}
		})
	}
	for _, c := range fixture.Converters {
		t.Run("argument/"+c.Name, func(t *testing.T) {
			p := &parser{state: newParseState()}
			value, err := p.convertDelimiterArgument(c.Raw)
			if c.Registered {
				if err != nil || c.Character == nil || value != *c.Character {
					t.Fatalf("delimiter = %q, %v; primary = %v", value, err, c.Character)
				}
			} else if c.Trimmed == "" {
				// GetDelimiterArg accepts an empty argument; the separate primary
				// map lookup returns null. These are different method contracts.
				if err != nil || value != "" {
					t.Fatalf("empty argument = %q, %v", value, err)
				}
			} else {
				got, ok := err.(*Error)
				if !ok || got.ID != "MissingOrUnrecognizedDelim" {
					t.Fatalf("unregistered argument accepted: %q, %v", value, err)
				}
			}
		})
	}
}
