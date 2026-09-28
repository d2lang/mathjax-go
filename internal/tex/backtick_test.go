// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package tex

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
)

func TestBacktickOriginalSVGAndMetadata(t *testing.T) {
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX, SVG, SVGSHA256, OriginalRecordSHA256 string
			Display                                         bool
			LeftQuotes                                      []struct {
				Path           []int
				Text           string
				Primes         *string
				Pseudoscript   bool
				Lspace, Rspace *float64
			}
		}
	}
	data, err := os.ReadFile("testdata/backtick_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 12 || fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" {
		t.Fatal("unbound backtick references")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			if len(c.OriginalRecordSHA256) != 64 || fmt.Sprintf("%x", sha256.Sum256([]byte(c.SVG))) != c.SVGSHA256 {
				t.Fatal("unbound SVG")
			}
			root, err := NewCompiler().Compile(c.TeX, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			count := 0
			root.Walk(func(n *mml.Node) bool {
				if n.Kind == "mo" && textContent(n) == "‘" {
					count++
				}
				return true
			})
			if count != len(c.LeftQuotes) {
				t.Fatalf("raw left-quote count = %d, want %d", count, len(c.LeftQuotes))
			}
			for _, o := range c.LeftQuotes {
				n := root
				for _, i := range o.Path {
					if i >= len(n.Children) {
						t.Fatalf("missing path %v", o.Path)
					}
					n = n.Children[i]
				}
				if n.Kind != "mo" || textContent(n) != o.Text {
					t.Fatalf("raw token changed at %v", o.Path)
				}
				if got, ok := n.Property("pseudoscript"); !ok || got != o.Pseudoscript {
					t.Errorf("path %v pseudoscript = %v/%v, want %v", o.Path, got, ok, o.Pseudoscript)
				}
				got, ok := n.Property("primes")
				if o.Primes == nil {
					if ok {
						t.Errorf("path %v unexpected primes=%v", o.Path, got)
					}
				} else if !ok || got != *o.Primes {
					t.Errorf("path %v primes=%v/%v, want %s", o.Path, got, ok, *o.Primes)
				}

				for key, want := range map[string]*float64{"lspace": o.Lspace, "rspace": o.Rspace} {
					got, ok := n.Attributes.Inherited().Get(key)
					if want == nil {
						if ok {
							t.Errorf("path %v unexpected inherited %s=%v", o.Path, key, got)
						}
						continue
					}
					if !ok || (got != int(*want) && got != *want) {
						t.Errorf("path %v inherited %s=%v, want %v", o.Path, key, got, *want)
					}
				}
			}
			options := pipeline.DefaultOptions()
			options.Display = c.Display
			got, err := svg.NewTypesetter().Typeset(root, options)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.SVG {
				t.Errorf("whole original SVG differs: got %x want %s", sha256.Sum256([]byte(got)), c.SVGSHA256)
			}
		})
	}
}

func TestBacktickLogicalParentAndEligibility(t *testing.T) {
	for _, kind := range []string{"msub", "msup", "msubsup", "mmultiscripts", "mover"} {
		t.Run(kind, func(t *testing.T) {
			quote := token("mo", "‘‘")
			// Both an embellished explicit row and a not-parent mstyle intervene.
			argument := node("mstyle", node("mrow", quote))
			children := []*mml.Node{token("mi", "x"), argument}
			if kind == "msubsup" || kind == "mmultiscripts" {
				children = append(children, token("mi", "y"))
			}
			root := node("math", node(kind, children...))
			setMathMLInheritance(root, true)
			want := kind == "msub" || kind == "mover"
			if got, ok := quote.Property("pseudoscript"); !ok || got != want {
				t.Errorf("logical parent %s: %v/%v, want %v", kind, got, ok, want)
			}
			if got, _ := quote.Property("primes"); got != "‵‵" || textContent(quote) != "‘‘" {
				t.Fatal("prime glyphs must be separate from raw token text")
			}
		})
	}
	for _, text := range []string{"", "‘x", "′x", "’x", "‵x"} {
		n := token("mo", text)
		setMathMLInheritance(node("math", n), false)
		if _, ok := n.Property("primes"); ok {
			t.Errorf("out-of-scope token %q acquired primes", text)
		}
		if _, ok := n.Property("pseudoscript"); ok {
			t.Errorf("out-of-scope token %q acquired pseudoscript", text)
		}
	}
	text := token("mtext", "‘")
	setMathMLInheritance(node("math", text), false)
	if _, ok := text.Property("primes"); ok {
		t.Fatal("mtext acquired mo metadata")
	}
}
