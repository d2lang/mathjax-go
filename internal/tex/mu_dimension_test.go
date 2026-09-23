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
		Dimensions    []struct {
			Source, Remaining string
			Value             *string
			Error             *struct{ ID, Message string }
			CursorBytesUTF8   int
		}
	}
	b, err := os.ReadFile("testdata/mu_dimension_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Normalization) != 44 || len(f.Dimensions) != 38 || len(f.Extremes) != 3 {
		t.Fatal("incomplete registered corpus")
	}
	for _, c := range append(f.Normalization, f.Extremes...) {
		t.Run("value/"+c.Source, func(t *testing.T) {
			if got := normalizeTeXMu(c.Source); got != c.Expected {
				t.Fatalf("value=%q, primary conversion=%q", got, c.Expected)
			}
		})
	}
	// D079 upgrades extraction to the same fixture's actual registered GetDimen
	// observations. Historical accepted extraction records remain in the fixture.
	for _, c := range f.Dimensions {
		t.Run("extraction/"+c.Source, func(t *testing.T) {
			p := &parser{source: c.Source, state: newParseState()}
			value, err := p.readDimension("kern")
			if c.Error == nil {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || err.Error() != c.Error.Message {
				t.Fatalf("error=%v want=%s", err, c.Error.Message)
			}
			want := ""
			if c.Value != nil {
				want = *c.Value
			}
			if value != want || p.pos != c.CursorBytesUTF8 || p.source[p.pos:] != c.Remaining || p.source != c.Source {
				t.Fatalf("got value=%q cursor=%d tail=%q; want=%q/%d/%q", value, p.pos, p.source[p.pos:], want, c.CursorBytesUTF8, c.Remaining)
			}
		})
	}
}
