package tex

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

type dimensionObservation struct {
	Source, Name, Remaining              string
	Value                                *string
	Error                                *struct{ ID, Message string }
	CursorBytesWithinSource, BeyondUTF16 int
}

func assertDimensionObservation(t *testing.T, c dimensionObservation, name string) {
	t.Helper()
	p := &parser{source: c.Source, state: newParseState()}
	value, err := p.readDimension(name)
	if c.Error == nil {
		if err != nil || c.Value == nil || value != *c.Value {
			t.Fatalf("value/error=%q/%v want=%v", value, err, c.Value)
		}
	} else {
		var e *Error
		if !errors.As(err, &e) || e.ID != c.Error.ID || e.Message != c.Error.Message || value != "" {
			t.Fatalf("error=%v value=%q", err, value)
		}
	}
	physical := min(p.pos, len(p.source))
	if physical != c.CursorBytesWithinSource || p.pos-physical != c.BeyondUTF16 || p.source[physical:] != c.Remaining || p.source != c.Source {
		t.Fatalf("physical=%d overrun=%d suffix=%q", physical, p.pos-physical, p.source[physical:])
	}
}
func TestDimensionGrammarRegisteredGetDimen(t *testing.T) {
	var f struct {
		Dimensions []dimensionObservation
		LabelCases []dimensionObservation
	}
	b, e := os.ReadFile("testdata/dimension_grammar_mathjax_3_2_2.json")
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(b, &f); e != nil {
		t.Fatal(e)
	}
	if len(f.Dimensions) != 106 || len(f.LabelCases) != 4 {
		t.Fatal("incomplete primary contract")
	}
	for i, c := range f.Dimensions {
		t.Run(fmt.Sprintf("%03d-%s", i, c.Name), func(t *testing.T) { assertDimensionObservation(t, c, strings.TrimPrefix(c.Name, "\\")) })
	}
	for i, c := range f.LabelCases {
		t.Run(fmt.Sprintf("label-%d", i), func(t *testing.T) { assertDimensionObservation(t, c, "begin") })
	}
}
func TestDimensionGrammarErrorReturnsBeforeContinuation(t *testing.T) {
	for _, prefix := range []string{`\kern`, `\mkern`, `\hskip`, `\mskip`, `\hspace`, `\raise`, `\lower`, `\minCDarrowwidth`, `\minCDarrowheight`, `x\above`, `\begin{spreadlines}`} {
		t.Run(prefix, func(t *testing.T) {
			source := prefix + `{1pc\`
			p := &parser{source: source, state: newParseState()}
			_, _, err := p.parseRow(0, false)
			var e *Error
			if !errors.As(err, &e) || e.ID != "MissingCloseBrace" || p.pos != len(source)+1 || p.source != source {
				t.Fatalf("error=%v cursor=%d/%d", err, p.pos, len(source))
			}
			compiled, err := NewCompiler().Compile(source, false)
			if err != nil || len(compiled.Find("merror")) != 1 {
				t.Fatalf("compiler must preserve the error without touching overrun cursor: %v", err)
			}
		})
	}
}
