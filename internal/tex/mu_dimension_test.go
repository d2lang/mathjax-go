package tex

import (
	"encoding/json"
	"os"
	"testing"
)

func TestTeXMuRegisteredValuesAndExtraction(t *testing.T) {
	var f struct {
		Normalization []struct{ Source, Expected string }
		Extremes      []struct{ Source, Expected string }
		Extraction    []struct {
			Source, Value, ExpectedValue, Remaining string
			Error                                   *string
			CursorBytesUTF8                         int
		}
	}
	b, err := os.ReadFile("testdata/mu_dimension_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Normalization) != 44 || len(f.Extraction) != 38 || len(f.Extremes) != 3 {
		t.Fatal("incomplete registered corpus")
	}
	for _, c := range append(f.Normalization, f.Extremes...) {
		t.Run("value/"+c.Source, func(t *testing.T) {
			if got := normalizeTeXMu(c.Source); got != c.Expected {
				t.Fatalf("value=%q, primary conversion=%q", got, c.Expected)
			}
		})
	}
	// The scanner is deliberately NOT claimed primary-equivalent. Each cursor,
	// error and tail is frozen from accepted ab6c's byte-identical e93 scanner;
	// ExpectedValue uses the actual registered primary on its returned value.
	for _, c := range f.Extraction {
		t.Run("extraction/"+c.Source, func(t *testing.T) {
			p := &parser{source: c.Source, state: newParseState()}
			value, err := p.readDimension("kern")
			if c.Error == nil {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || err.Error() != *c.Error {
				t.Fatalf("error=%v want=%s", err, *c.Error)
			}
			if value != c.ExpectedValue || p.pos != c.CursorBytesUTF8 || p.source[p.pos:] != c.Remaining || p.source != c.Source {
				t.Fatalf("got value=%q cursor=%d tail=%q; want=%q/%d/%q", value, p.pos, p.source[p.pos:], c.ExpectedValue, c.CursorBytesUTF8, c.Remaining)
			}
		})
	}
}
