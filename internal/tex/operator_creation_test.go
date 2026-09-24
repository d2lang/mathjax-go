// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package tex

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// These are the existing 68 authored requests. The test observes the actual
// per-parse registry; walking the result only checks coverage and never adds
// entries or substitutes tree order for constructor order.
func TestOperatorCreationCoverage(t *testing.T) {
	data, err := os.ReadFile("testdata/operator_creation_requests.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name, TeX string
			Display   bool
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 68 {
		t.Fatalf("requests = %d, want 68", len(fixture.Cases))
	}
	// The accepted current Compiler renders these two unresolved errors.
	// They are partial-construction coverage, not successful primary renders.
	boundaries := map[string]string{
		"construction-dots-inline":  `Undefined control sequence \dots`,
		"construction-dots-display": `Undefined control sequence \dots`,
	}
	seen := make(map[string]bool)
	foreign := make(map[*mml.Node]string)
	boundaryCount := 0
	for _, c := range fixture.Cases {
		if seen[c.Name] {
			t.Fatalf("duplicate request %q", c.Name)
		}
		seen[c.Name] = true
		if _, boundary := boundaries[c.Name]; boundary {
			boundaryCount++
		}
		t.Run(c.Name, func(t *testing.T) {
			root, state, parseErr := operatorCreationParse(c.TeX, c.Display)
			operatorCreationCheck(t, root, state.operators, foreign)
			if message, boundary := boundaries[c.Name]; boundary {
				var texErr *Error
				if !errors.As(parseErr, &texErr) || texErr.ID != "UndefinedControlSequence" || texErr.Message != message {
					t.Fatalf("accepted-baseline parse boundary = %#v, want UndefinedControlSequence %q", parseErr, message)
				}
				merrors := root.Find("merror")
				if len(merrors) != 1 {
					t.Fatalf("merror count = %d, want 1", len(merrors))
				}
				if value, _ := merrors[0].Attributes.GetExplicit("data-mjx-error"); value != message {
					t.Fatalf("accepted-baseline merror = %#v, want %q", value, message)
				}
				return // Actual Compile returns here, before the relation filter.
			}
			if parseErr != nil {
				t.Fatal(parseErr)
			}
			if len(root.Find("merror")) != 0 {
				t.Fatal("unexpected rendered error")
			}
			before := append([]*mml.Node(nil), state.operators...)
			state.operators = combineRelations(root, state.operators)
			operatorCreationCheckAfter(t, root, before, state.operators)
		})
	}
	if boundaryCount != 2 {
		t.Fatalf("accepted-baseline boundaries = %d, want 2", boundaryCount)
	}
}

// These exact strings come from TestCasesNumCasesFrozenShape and
// TestEmpheqEnvironmentFrozenMML. They are direct source-method tests, not a
// claim that Empheq is in D2's public package configuration. numcases uses the
// same explicit augmented state as its existing test. Empheq is called directly
// as in its existing tests. No newly authored formula or primary golden is used.
func TestOperatorCreationCopyRoutes(t *testing.T) {
	tests := []struct {
		name, source, method string
		augmented            bool
	}{
		{"numcases", `{f(x)=}x&if x>0\\0&otherwise\end{numcases}`, "numcases", true},
		{"empheq-left-right", `[left={(},right={)}]{align}a&=b\end{empheq}`, "empheq", false},
		{"empheq-gather-left", `[left={(}]{gather}a\\b\end{empheq}`, "empheq", false},
	}
	foreign := make(map[*mml.Node]string)
	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			p := &parser{source: c.source, state: newParseState(), display: true}
			p.state.augmentedPackages = c.augmented
			var children []*mml.Node
			var handled bool
			var err error
			if c.method == "numcases" {
				children, handled, err = p.casesEnvironment(c.method)
			} else {
				children, handled, err = p.empheqEnvironment(c.method)
			}
			if err != nil {
				t.Fatal(err)
			}
			if !handled || p.pos != len(p.source) {
				t.Fatalf("handled=%v, cursor=%d/%d", handled, p.pos, len(p.source))
			}
			root := node("math", children...)
			root.Attributes.Set("display", "block")
			root = operatorCreationFinish(root, true)
			operatorCreationCheck(t, root, p.state.operators, foreign)
			if c.name == "empheq-left-right" {
				// The original table snapshot is detached; its live copies
				// have distinct '=' pointers. The other two exact inputs do
				// execute copy routes, but their copied tables contain no mo.
				detached, equalSigns := 0, 0
				for _, n := range p.state.operators {
					if !operatorCreationAttached(root, n) {
						detached++
					}
					if textContent(n) == "=" {
						equalSigns++
					}
				}
				if detached == 0 || equalSigns < 2 {
					t.Fatalf("missing real copy coverage: detached=%d, distinct equals=%d", detached, equalSigns)
				}
			}
			before := append([]*mml.Node(nil), p.state.operators...)
			p.state.operators = combineRelations(root, p.state.operators)
			operatorCreationCheckAfter(t, root, before, p.state.operators)
		})
	}
}

func operatorCreationAttached(root, n *mml.Node) bool {
	for parent := n; parent != nil; parent = parent.Parent {
		if parent == root {
			return true
		}
	}
	return false
}

func operatorCreationCheck(t *testing.T, root *mml.Node, operators []*mml.Node, foreign map[*mml.Node]string) {
	t.Helper()
	seen := make(map[*mml.Node]bool)
	for _, n := range operators {
		if n == nil || n.Kind != "mo" {
			t.Fatalf("non-mo registration: %#v", n)
		}
		if seen[n] {
			t.Fatalf("operator %q registered more than once", textContent(n))
		}
		seen[n] = true
		if owner, exists := foreign[n]; exists {
			t.Fatalf("operator pointer reused from parse %s", owner)
		}
		// Retaining the actual pointer also prevents GC address reuse from
		// being mistaken for a cross-parse registration leak.
		foreign[n] = t.Name()
	}
	root.Walk(func(n *mml.Node) bool {
		if n.Kind == "mo" && !seen[n] {
			t.Fatalf("live mo %q has no actual creation event", textContent(n))
		}
		return true
	})
}

func operatorCreationCheckAfter(t *testing.T, root *mml.Node, before, after []*mml.Node) {
	t.Helper()
	live := make(map[*mml.Node]bool)
	root.Walk(func(n *mml.Node) bool {
		if n.Kind == "mo" {
			live[n] = true
		}
		return true
	})
	expected := []*mml.Node{}
	for _, n := range before {
		if operatorCreationAttached(root, n) {
			expected = append(expected, n)
		}
	}
	if len(after) != len(expected) {
		t.Fatalf("filtered registry length=%d, live creation subsequence=%d", len(after), len(expected))
	}
	seen := make(map[*mml.Node]bool)
	for i, n := range after {
		if n != expected[i] {
			t.Fatalf("filtered registry changed creation order at %d", i)
		}
		if seen[n] || !live[n] {
			t.Fatalf("duplicate or non-live filtered mo %q", textContent(n))
		}
		seen[n] = true
	}
	for n := range live {
		if !seen[n] {
			t.Fatalf("filtered live mo %q missing from registry", textContent(n))
		}
	}
}

// This is the current Compiler prefix through attribute cleanup, using its
// real parser and constructors. The source-only preparation binds the exact
// copied body. It exposes the parse state and underlying TeX error for coverage
// assertions; public Compile instead renders that error and returns nil error.
// Filtering remains at the callsite so both pre/post lists can be checked.
func operatorCreationParse(source string, display bool) (*mml.Node, *parseState, error) {
	state := newParseState()
	p := &parser{source: source, state: state, display: display}
	children, stop, err := p.parseRow(0, false)
	if err != nil {
		var parseError *Error
		if !errors.As(err, &parseError) {
			return nil, state, err
		}
		return mathError(parseError.Message, display), state, err
	}
	if stop != "" {
		return nil, state, texError("ExtraCloseMissingOpen", "Extra close brace or missing open brace")
	}
	children, err = p.amsTagFinalize(children)
	if err != nil {
		var parseError *Error
		if !errors.As(err, &parseError) {
			return nil, state, err
		}
		return mathError(parseError.Message, display), state, err
	}
	// MmlMath has an inferred-mrow child even when the TeX stack reduces to a
	// single node. Keep that wrapper explicit in the shared Go tree so layout
	// sees the same structure as the JavaScript MML factory.
	root := node("math", children...)
	if display {
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
	setMathMLInheritance(root, display)
	root = moveMathLimits(root)
	cleanMathMLAttributes(root)
	return root, state, nil
}

// Same post-parse cleanup as Compiler, for the direct-method copy controls.
func operatorCreationFinish(root *mml.Node, display bool) *mml.Node {
	root.Walk(func(n *mml.Node) bool {
		n.RemoveProperty(resolvedFontScope)
		n.RemoveProperty(ambientFontSource)
		n.RemoveProperty(vectorFactoryToken)
		n.RemoveProperty(vectorFactoryDone)
		n.RemoveProperty(limitsScriptOrigin)
		return true
	})
	setMathMLInheritance(root, display)
	root = moveMathLimits(root)
	cleanMathMLAttributes(root)
	return root
}
