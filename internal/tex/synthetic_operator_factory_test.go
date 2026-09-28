// Copyright 2018-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package tex

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/ordered"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
)

type syntheticOriginal struct {
	Name string
	Display bool
	Spec json.RawMessage
	Original struct { SVG string }
	OriginalSHA256 string
	InheritedCSSControl bool
	Baseline167 struct { SVG string }
	BaselineSHA256 string
	FakeObservations []struct {
		ParentExplicit, ParentInherited map[string]any
		Fake []struct {
			Kind, Text string
			Explicit, Inherited, Properties map[string]any
		}
	}
}

func syntheticOperatorOriginals(t *testing.T) []syntheticOriginal {
	t.Helper()
	var fixture struct {
		MathJaxGitCommit string
		Cases []syntheticOriginal
	}
	data, err := os.ReadFile("testdata/synthetic_operator_factory_mathjax_3_2_2.json")
	if err != nil { t.Fatal(err) }
	if err := json.Unmarshal(data, &fixture); err != nil { t.Fatal(err) }
	if fixture.MathJaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 96 {
		t.Fatal("unbound original constructed-MathML inventory")
	}
	seen := map[string]bool{}
	css, observed := 0, 0
	for _, c := range fixture.Cases {
		if seen[c.Name] { t.Fatal("duplicate original name", c.Name) }
		seen[c.Name] = true
		if fmt.Sprintf("%x", sha256.Sum256([]byte(c.Original.SVG))) != c.OriginalSHA256 { t.Fatal("original SVG binding", c.Name) }
		if c.InheritedCSSControl {
			css++
			if fmt.Sprintf("%x", sha256.Sum256([]byte(c.Baseline167.SVG))) != c.BaselineSHA256 { t.Fatal("baseline SVG binding", c.Name) }
		}
		if len(c.FakeObservations) != 0 { observed++ }
	}
	if css != 8 || observed != 36 { t.Fatal("constructed-MathML partition changed", css, observed) }
	return fixture.Cases
}

// All trees use the real registered Go constructors and inheritance pass.
// None is rendered from a projection of the original's prepared tree.
func TestSyntheticOperatorFactoryOriginalSVG(t *testing.T) {
	for _, c := range syntheticOperatorOriginals(t) {
		t.Run(c.Name, func(t *testing.T) {
			root := idAnchorBuild(c.Spec, c.Display)
			options := pipeline.DefaultOptions()
			options.Display = c.Display
			want := c.Original.SVG
			if c.InheritedCSSControl { want = c.Baseline167.SVG }
			for run := 0; run < 2; run++ {
				got, err := svg.NewTypesetter().Typeset(root, options)
				if err != nil { t.Fatal(err) }
				if got != want {
					t.Fatalf("complete SVG mismatch (run %d; inherited CSS control %v)\ngot: %s\nwant: %s", run, c.InheritedCSSControl, got, want)
				}
			}
		})
	}
}

func syntheticMap(values *ordered.Map[mml.Property]) map[string]any {
	out := map[string]any{}
	values.Range(func(name string, value mml.Property) bool { out[name] = value; return true })
	return out
}

func syntheticEqualJSON(t *testing.T, got, want any) {
	t.Helper()
	a, err := json.Marshal(got)
	if err != nil { t.Fatal(err) }
	b, err := json.Marshal(want)
	if err != nil { t.Fatal(err) }
	if string(a) != string(b) { t.Fatalf("original metadata differs\ngot: %s\nwant: %s", a, b) }
}

func syntheticFakes(n *mml.Node) []*mml.Node {
	if n.Fenced == nil { return nil }
	var nodes []*mml.Node
	for _, fake := range append([]*mml.Node{n.Fenced.Open, n.Fenced.Close}, n.Fenced.Separators...) {
		if fake != nil { nodes = append(nodes, fake) }
	}
	return nodes
}

func TestFencedIncomingInheritanceOriginal(t *testing.T) {
	for _, c := range syntheticOperatorOriginals(t) {
		if len(c.FakeObservations) == 0 { continue }
		t.Run(c.Name, func(t *testing.T) {
			root := idAnchorBuild(c.Spec, c.Display)
			fences := root.Find("mfenced")
			if len(fences) != len(c.FakeObservations) { t.Fatal("original observer count differs") }
			for i, fence := range fences {
				observation := c.FakeObservations[i]
				syntheticEqualJSON(t, syntheticMap(fence.Attributes.Explicit()), observation.ParentExplicit)
				syntheticEqualJSON(t, syntheticMap(fence.Attributes.Inherited()), observation.ParentInherited)
				fakes := syntheticFakes(fence)
				if len(fakes) != len(observation.Fake) { t.Fatal("original fake-node count differs") }
				for j, fake := range fakes {
					want := observation.Fake[j]
					if fake.Kind != want.Kind || textContent(fake) != want.Text || fake.Parent != fence {
						t.Fatal("fake-node identity or ownership differs")
					}
					syntheticEqualJSON(t, syntheticMap(fake.Attributes.Explicit()), want.Explicit)
					syntheticEqualJSON(t, syntheticMap(fake.Attributes.Inherited()), want.Inherited)
					syntheticEqualJSON(t, syntheticMap(fake.Properties), want.Properties)
				}
			}
			clone := root.Clone()
			clonedFences := clone.Find("mfenced")
			for i, fence := range fences {
				cloned := clonedFences[i]
				if fence == cloned || fence.Fenced == cloned.Fenced { t.Fatal("clone retained mutable fenced state") }
				originalFakes, clonedFakes := syntheticFakes(fence), syntheticFakes(cloned)
				if len(originalFakes) != len(clonedFakes) { t.Fatal("clone lost fake nodes") }
				for j, fake := range clonedFakes {
					if fake == originalFakes[j] || fake.Parent != cloned || originalFakes[j].Parent != fence { t.Fatal("clone fake ownership differs") }
				}
			}
			options := pipeline.DefaultOptions()
			options.Display = c.Display
			originalSVG, err := svg.NewTypesetter().Typeset(root, options)
			if err != nil { t.Fatal(err) }
			clonedSVG, err := svg.NewTypesetter().Typeset(clone, options)
			if err != nil || clonedSVG != originalSVG { t.Fatal("clone changes rendered output", err) }
		})
	}
}

func TestFencedUninitializedAndEmptyClone(t *testing.T) {
	fence := node("mfenced")
	if fence.Fenced != nil || fence.Clone().Fenced != nil { t.Fatal("uninitialized fake nodes were materialized") }
	fence.Attributes.Set("open", "")
	fence.Attributes.Set("close", "")
	fence.Attributes.Set("separators", "")
	setMathMLInheritance(fence, false)
	clone := fence.Clone()
	if fence.Fenced == nil || clone.Fenced == nil || fence.Fenced == clone.Fenced || !reflect.DeepEqual(fence.Fenced, clone.Fenced) {
		t.Fatal("initialized empty fake nodes were not preserved independently")
	}
}
