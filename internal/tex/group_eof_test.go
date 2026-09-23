// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
)

func TestParsedGroupEOFPinnedReferences(t *testing.T) {
	type output struct {
		SVGSHA256 string
		Tree      any
		Error     *Error
	}
	var fixture struct {
		Cases []struct {
			Name, TeX string
			Display   bool
			Cursor    infixScopeCursor
			output
		}
	}
	var baseline struct {
		Baseline string
		Controls map[string]struct {
			TeX      string
			Display  bool
			FullTree any
			Parser   struct {
				Source          string
				ByteCursor      int
				Remaining, Stop string
				MacroCount      int
			}
			output
		}
	}
	var primaryMissingRight, primaryHash map[string]any
	for name, target := range map[string]any{"group_eof_mathjax_3_2_2.json": &fixture, "group_eof_boundaries.json": &baseline, "missing_right_full_mathjax_3_2_2.json": &primaryMissingRight, "hash_diagnostic_full_mathjax_3_2_2.json": &primaryHash} {
		b, err := os.ReadFile(filepath.Join("../../testdata", name))
		if err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(b, target); err != nil {
			t.Fatal(err)
		}
	}
	// Six former diagnostic controls now use the untouched primary outputs.
	// Keep the historical accepted records for source/cursor checks; the
	// argument-reader and hash cursor boundaries remain separate.
	promoted := map[string]bool{
		"missing-right-inline": true, "missing-right-display": true,
		"right-before-group-inline": true, "right-before-group-display": true,
		"hash-before-eof-inline": true, "hash-before-eof-display": true,
	}
	cursorBoundary := map[string]bool{
		"argument-first-inline": true, "argument-first-display": true,
		"argument-second-inline": true, "argument-second-display": true,
		"font-argument-inline": true, "font-argument-display": true,
		"root-argument-inline": true, "root-argument-display": true,
		"hash-before-eof-inline": true, "hash-before-eof-display": true,
	}
	if len(primaryHash) != 2 || len(primaryMissingRight) != 4 || len(fixture.Cases) != 64 || len(baseline.Controls) != 40 || baseline.Baseline != "01eef36d33412dddb14af76d14720c1ff21bd6f2" {
		t.Fatal("unbound parsed-group corpus")
	}
	primary, fixed, controls, baselineCursors := 0, 0, 0, 0
	seen := map[string]bool{}
	for _, c := range fixture.Cases {
		if seen[c.Name] {
			t.Fatal("duplicate case", c.Name)
		}
		seen[c.Name] = true
		t.Run(c.Name, func(t *testing.T) {
			want := c.output
			b, control := baseline.Controls[c.Name]
			fullPrimary := primaryMissingRight[c.Name]
			wantID, wantMessage := "ExtraLeftMissingRight", "Extra \\left or missing \\right"
			if hash, ok := primaryHash[c.Name]; ok {
				fullPrimary = hash
				wantID, wantMessage = "CantUseHash1", "You can't use 'macro parameter character #' in math mode"
			}
			if promoted[c.Name] && (!control || fullPrimary == nil || c.Error == nil || c.Error.ID != wantID || c.Error.Message != wantMessage) {
				t.Fatal("unbound diagnostic promotion")
			}
			if control {
				controls++
				if b.TeX != c.TeX || b.Display != c.Display {
					t.Fatal("unbound control")
				}
			} else {
				fixed++
				if c.Error == nil || c.Error.ID != "ExtraOpenMissingClose" || c.Error.Message != "Extra open brace or missing close brace" {
					t.Fatal("unexpected targeted error")
				}
			}
			primary++
			p := &parser{source: c.TeX, state: newParseState(), display: c.Display}
			_, stop, err := p.parseRow(0, false)
			var typed *Error
			if err != nil && !errors.As(err, &typed) {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(typed, want.Error) {
				t.Errorf("error %#v; want %#v", typed, want.Error)
			}
			cursor := infixScopeCursor{p.source, p.pos, p.source[p.pos:]}
			if cursorBoundary[c.Name] {
				if !control {
					t.Fatal("cursor qualification outside accepted control")
				}
				baselineCursors++
			} else if cursor != c.Cursor {
				t.Errorf("source/cursor/tail %#v; want %#v", cursor, c.Cursor)
			}
			if stop != "" {
				t.Fatal("unexpected root stop", stop)
			}
			if control && (p.source != b.Parser.Source || p.pos != b.Parser.ByteCursor || p.source[p.pos:] != b.Parser.Remaining || stop != b.Parser.Stop || p.state.macroCount != b.Parser.MacroCount) {
				t.Fatal("accepted parser state changed")
			}
			root, err := NewCompiler().Compile(c.TeX, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			full := infixMacroTree(t, root)
			if promoted[c.Name] && !reflect.DeepEqual(infixScopeValue(t, full), fullPrimary) {
				t.Error("complete primary diagnostic tree differs")
			}
			if !reflect.DeepEqual(infixScopeValue(t, infixScopeProjection(full)), want.Tree) {
				t.Error("complete ordered explicit/own tree differs")
			}
			if control && !promoted[c.Name] && !reflect.DeepEqual(infixScopeValue(t, full), b.FullTree) {
				t.Error("complete accepted control tree changed")
			}
			opts := pipeline.DefaultOptions()
			opts.Display = c.Display
			s, err := svg.NewTypesetter().Typeset(root, opts)
			if err != nil {
				t.Fatal(err)
			}
			hash := fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
			if hash != want.SVGSHA256 {
				t.Errorf("whole SVG %s; want %s", hash, want.SVGSHA256)
			}
			if control && !promoted[c.Name] && hash != b.SVGSHA256 {
				t.Error("whole accepted control SVG changed")
			}
		})
	}
	if primary != 64 || fixed != 24 || controls != 40 || baselineCursors != 10 {
		t.Fatal("changed qualification counts", primary, fixed, controls, baselineCursors)
	}
}
