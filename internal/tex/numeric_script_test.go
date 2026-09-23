// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"unicode/utf16"

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
)

func TestNumericScriptPinnedReferences(t *testing.T) {
	var f struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVGSHA256 string
			Display              bool
			PropertiesTree       *argumentTree
			Registration         *struct {
				Name, Body string
				Arguments  int
			}
		}
	}
	readArgumentJSON(t, "../../testdata/numeric_script_mathjax_3_2_2.json", &f)
	var b struct {
		AcceptedBase string
		Cases        map[string]struct {
			Kind, TeX, PrimarySVGSHA256, AcceptedSVGSHA256, CandidateSVGSHA256 string
			Display                                                            bool
			AcceptedTree, CandidateTree                                        *argumentTree
			Fields                                                             []struct {
				Path                         []int
				Primary, Accepted, Candidate map[string]any
			}
		}
	}
	readArgumentJSON(t, "../../testdata/numeric_script_boundaries.json", &b)
	if f.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(f.Cases) != 120 || len(b.Cases) != 24 || b.AcceptedBase != "6a3573f7f15fc67b5abde6fdf8b548e89d6c7e48" {
		t.Fatal("unbound numeric references")
	}
	counts := map[string]int{}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			state := newParseState()
			if c.Registration != nil {
				r := c.Registration
				state.macros[r.Name] = macroDefinition{body: r.Body, arguments: r.Arguments}
			}
			p := &parser{source: c.TeX, state: state, display: c.Display}
			children, stop, err := p.parseRow(0, false)
			var root *mml.Node
			if err != nil {
				var e *Error
				if !errors.As(err, &e) {
					t.Fatal(err)
				}
				root = mathError(e.Message, c.Display)
			} else {
				if stop != "" {
					t.Fatal("unexpected row stop", stop)
				}
				children, err = p.amsTagFinalize(children)
				if err != nil {
					t.Fatal(err)
				}
				root = node("math", children...)
				if c.Display {
					root.Attributes.Set("display", "block")
				}
				root.Walk(func(n *mml.Node) bool {
					n.RemoveProperty(resolvedFontScope)
					n.RemoveProperty(ambientFontSource)
					n.RemoveProperty(vectorFactoryToken)
					n.RemoveProperty(vectorFactoryDone)
					n.RemoveProperty(limitsScriptOrigin)
					return true
				})
				setMathMLInheritance(root, c.Display)
				root = moveMathLimits(root)
				cleanMathMLAttributes(root)
			}
			if c.Registration == nil {
				actual, e := NewCompiler().Compile(c.TeX, c.Display)
				if e != nil {
					t.Fatal(e)
				}
				if !reflect.DeepEqual(argumentProjection(root), argumentProjection(actual)) {
					t.Fatal("registered harness differs from actual compiler")
				}
				root = actual
			}
			root.Walk(func(n *mml.Node) bool {
				for _, ch := range n.Children {
					if ch.Parent != n {
						t.Fatal("child ownership changed")
					}
				}
				return true
			})
			want, hash := c.PropertiesTree, c.SVGSHA256
			if q, ok := b.Cases[c.Name]; ok {
				if q.TeX != c.TeX || q.Display != c.Display || q.PrimarySVGSHA256 != c.SVGSHA256 {
					t.Fatal("qualification input/primary changed")
				}
				counts[q.Kind]++
				switch q.Kind {
				case "unchanged-scanner":
					if q.CandidateSVGSHA256 != q.AcceptedSVGSHA256 || !reflect.DeepEqual(q.CandidateTree, q.AcceptedTree) {
						t.Fatal("unchanged scanner boundary changed")
					}
				case "initial-split-with-retained-comma-scanner":
					if q.CandidateSVGSHA256 == q.AcceptedSVGSHA256 || q.CandidateSVGSHA256 == c.SVGSHA256 || reflect.DeepEqual(q.CandidateTree, q.AcceptedTree) {
						t.Fatal("comma observation lost its partial-change boundary")
					}
				case "inherited-prime-pseudoscript":
					if q.CandidateSVGSHA256 != c.SVGSHA256 || len(q.Fields) != 1 {
						t.Fatal("prime SVG/field scope changed")
					}
					for _, d := range q.Fields {
						at, old, next := want, q.AcceptedTree, q.CandidateTree
						for _, i := range d.Path {
							at = at.Children[i]
							old = old.Children[i]
							next = next.Children[i]
						}
						if !reflect.DeepEqual(at.Properties, d.Primary) || !reflect.DeepEqual(old.Properties, d.Accepted) || !reflect.DeepEqual(next.Properties, d.Candidate) || !reflect.DeepEqual(d.Accepted, d.Candidate) {
							t.Fatal("prime field binding changed")
						}
						if _, ok := d.Primary["pseudoscript"].(bool); !ok {
							t.Fatal("missing exact boolean boundary")
						}
						copy := map[string]any{}
						for k, v := range d.Primary {
							if k != "pseudoscript" {
								copy[k] = v
							}
						}
						if !reflect.DeepEqual(copy, d.Candidate) {
							t.Fatal("another prime field changed")
						}
						at.Properties = copy
					}
					if !reflect.DeepEqual(want, q.CandidateTree) {
						t.Fatal("prime differs outside exact inherited field")
					}
				default:
					t.Fatal("unknown boundary", q.Kind)
				}
				want, hash = q.CandidateTree, q.CandidateSVGSHA256
			} else {
				counts["raw-primary"]++
			}
			got := argumentProjection(root)
			encoded, _ := json.Marshal(got)
			var decoded *argumentTree
			if err = json.Unmarshal(encoded, &decoded); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(decoded, want) {
				t.Fatalf("complete explicit/own tree differs: %s", encoded)
			}
			opts := pipeline.DefaultOptions()
			opts.Display = c.Display
			s, e := svg.NewTypesetter().Typeset(root, opts)
			if e != nil {
				t.Fatal(e)
			}
			if actual := fmt.Sprintf("%x", sha256.Sum256([]byte(s))); actual != hash {
				t.Fatalf("whole SVG %s want %s", actual, hash)
			}
		})
	}
	if !reflect.DeepEqual(counts, map[string]int{"raw-primary": 96, "inherited-prime-pseudoscript": 4, "initial-split-with-retained-comma-scanner": 4, "unchanged-scanner": 16}) {
		t.Fatal("reference classification changed", counts)
	}
}

func TestNumericScriptRegisteredInitialHandler(t *testing.T) {
	var f struct {
		MathjaxGitCommit string
		Rows             []struct {
			Marker, Tail, SourceAfter, Remaining string
			Occupied                             bool
			CursorUTF16                          int
			BeforeBase, AfterBase, Prepared      *argumentTree
			ReusedBase, PreparedOwnsOriginalBase *bool
			Error                                *struct{ ID, Message string }
		}
	}
	readArgumentJSON(t, "testdata/numeric_script_handler_mathjax_3_2_2.json", &f)
	if f.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(f.Rows) != 48 {
		t.Fatal("unbound registered handler records")
	}
	for i, c := range f.Rows {
		t.Run(fmt.Sprintf("%d/%s/%q", i, c.Marker, c.Tail), func(t *testing.T) {
			base := limitToken("mi", "x")
			if c.Occupied {
				base = texMMLFactory.Create("msubsup", base, limitToken("mi", "i"), limitToken("mi", "n"))
			}
			before := argumentProjection(base)
			originalChildren := append([]*mml.Node(nil), base.Children...)
			if !reflect.DeepEqual(before, c.BeforeBase) {
				t.Fatal("original base differs")
			}
			p := &parser{source: c.Tail, state: newParseState()}
			p.scriptInitialLookahead()
			attachment, err := prepareScriptAttachment(base, c.Marker[0], false)
			if p.source != c.SourceAfter || len(utf16.Encode([]rune(p.source[:p.pos]))) != c.CursorUTF16 || p.source[p.pos:] != c.Remaining {
				t.Fatalf("registered source/cursor: %q/%d/%q", p.source, p.pos, p.source[p.pos:])
			}
			if !reflect.DeepEqual(argumentProjection(base), c.AfterBase) || len(base.Children) != len(originalChildren) {
				t.Fatal("original base identity/content changed")
			}
			for j, child := range originalChildren {
				if base.Children[j] != child {
					t.Fatal("original child identity or order changed")
				}
			}
			if c.Error != nil {
				var e *Error
				if !errors.As(err, &e) || e.ID != c.Error.ID || e.Message != c.Error.Message {
					t.Fatal("registered error changed", err)
				}
				return
			}
			if err != nil || attachment.base != base || attachment.reuse != *c.ReusedBase || !*c.PreparedOwnsOriginalBase {
				t.Fatal("attachment ownership changed", err)
			}
			view := attachment.pendingBase()
			if !attachment.reuse && (len(view.Children) == 0 || view.Children[0] != base) {
				t.Fatal("prepared attachment does not own the original base")
			}
			if !reflect.DeepEqual(argumentProjection(view), c.Prepared) {
				t.Fatal("prepared family differs")
			}
		})
	}
}

func TestNumericScriptSourceAndCursor(t *testing.T) {
	var f struct {
		MathjaxGitCommit string
		Rows             []struct {
			Name, TeX, SourceAfter, Remaining string
			Display                           bool
			CursorAfterUTF16, MacroCountAfter int
			Error                             *struct{ ID, Message string }
		}
	}
	readArgumentJSON(t, "testdata/numeric_script_cursors.json", &f)
	if f.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(f.Rows) != 108 {
		t.Fatal("unbound delegated parser traces")
	}
	for _, c := range f.Rows {
		t.Run(c.Name, func(t *testing.T) {
			p := &parser{source: c.TeX, state: newParseState(), display: c.Display}
			_, _, err := p.parseRow(0, false)
			var e *Error
			if c.Error != nil && (!errors.As(err, &e) || e.ID != c.Error.ID || e.Message != c.Error.Message) {
				t.Fatal("error order changed", err)
			}
			if p.source != c.SourceAfter || p.source[p.pos:] != c.Remaining || len(utf16.Encode([]rune(p.source[:p.pos]))) != c.CursorAfterUTF16 || p.state.macroCount != c.MacroCountAfter {
				t.Fatalf("error source/cursor/lifetime differs: %q/%d/%q", p.source, p.pos, p.source[p.pos:])
			}
		})
	}
}

func TestNumericScriptWhitespaceBoundary(t *testing.T) {
	var f struct {
		Cases []struct {
			Name, TeX, Status, SVGSHA256 string
			Display                      bool
			PropertiesTree               *argumentTree
		}
	}
	readArgumentJSON(t, "../../testdata/numeric_script_whitespace.json", &f)
	if len(f.Cases) != 8 {
		t.Fatal("whitespace inventory changed")
	}
	statuses := map[string]int{}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.Status != "raw-primary" && c.Status != "uncomparable-primary-exception-unchanged-Go" {
				t.Fatal("unknown boundary")
			}
			statuses[c.Status]++
			root, e := NewCompiler().Compile(c.TeX, c.Display)
			if e != nil {
				t.Fatal(e)
			}
			encoded, _ := json.Marshal(argumentProjection(root))
			var got *argumentTree
			if e = json.Unmarshal(encoded, &got); e != nil {
				t.Fatal(e)
			}
			if !reflect.DeepEqual(got, c.PropertiesTree) {
				t.Fatal("whole tree differs")
			}
			o := pipeline.DefaultOptions()
			o.Display = c.Display
			s, e := svg.NewTypesetter().Typeset(root, o)
			if e != nil {
				t.Fatal(e)
			}
			if fmt.Sprintf("%x", sha256.Sum256([]byte(s))) != c.SVGSHA256 {
				t.Fatal("whole whitespace SVG differs")
			}
		})
	}
	if !reflect.DeepEqual(statuses, map[string]int{"raw-primary": 4, "uncomparable-primary-exception-unchanged-Go": 4}) {
		t.Fatal("whitespace classifications changed", statuses)
	}
}
