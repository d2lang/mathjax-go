package tex

import (
	"errors"
	"testing"
)

func TestEnvironmentCounterBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count, want  int
		errorID      string
	}{
		{"opening-at-limit", `\begin{array}{c}x\end{array}`, maxMacros, maxMacros + 1, "MaxMacroSub2"},
		{"opening-last-slot", `\begin{smallmatrix}x\end{smallmatrix}`, maxMacros - 1, maxMacros, ""},
		{"unknown-opening-at-limit", `\begin{unknown}`, maxMacros, maxMacros + 1, "MaxMacroSub2"},
		{"invalid-before-charge", `\begin{a\b}`, maxMacros, maxMacros, "InvalidEnv"},
		{"missing-before-charge", `\begin`, maxMacros, maxMacros, "MissingArgFor"},
		{"built-in-extra-end-free", `\end{array}`, maxMacros, maxMacros, "ExtraEnd"},
		{"unknown-extra-end-charged", `\end{unknown}`, maxMacros, maxMacros + 1, "MaxMacroSub2"},
		{"truthy-extra-end-charged", `\end{spreadlines}`, maxMacros, maxMacros + 1, "MaxMacroSub2"},
		{"cases-extra-end-charged", `\end{numcases}`, maxMacros, maxMacros + 1, "MaxMacroSub2"},
		{"invalid-extra-end-free", `\end{a\b}`, maxMacros, maxMacros, "InvalidEnv"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := newParseState()
			state.macroCount = tc.count
			p := &parser{source: tc.source, state: state}
			_, _, err := p.parseRow(0, false)
			var failure *Error
			if tc.errorID == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if !errors.As(err, &failure) || failure.ID != tc.errorID {
				t.Fatalf("error = %v; want %s", err, tc.errorID)
			}
			if state.macroCount != tc.want {
				t.Fatalf("counter = %d; want %d", state.macroCount, tc.want)
			}
		})
	}
}
