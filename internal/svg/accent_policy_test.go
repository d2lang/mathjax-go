// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package svg

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
)

// Reuse the complete registered-MathML decoder and input snapshot from the ms
// oracle tests. No parser or attribute rewriting is applied to these trees.
// These two named same-MathML fixtures differ only in this one literal tag.
// Exact full-output hashing after the substitution rejects every other change.
func accentPolicySVGMatches(name, actual, primarySHA string, baseline map[string]string) bool {
	hash := func(s string) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(s))) }
	if hash(actual) == primarySHA {
		return true
	}
	switch name {
	case "ordinary-unmarked-inline", "ordinary-unmarked-display", "ordinary-null-inline", "ordinary-null-display":
		return hash(actual) == baseline[name]
	case "both-mixed-inline", "both-mixed-display":
		const from = `<g data-mml-node="mo" transform="translate(313.8,31) translate(-250 0)">`
		const to = `<g data-mml-node="mo" transform="translate(63.8,31)">`
		return strings.Count(actual, from) == 1 && hash(strings.Replace(actual, from, to, 1)) == primarySHA
	}
	return false
}

func TestAccentPolicySameMathML(t *testing.T) {
	b, err := os.ReadFile("testdata/accent_policy_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Cases []struct {
			Name, SHA256 string
			Display      bool
			Tree         *msTree
			Observations []struct {
				Path     []int
				Kind     string
				IsAccent *bool
			}
		}
	}
	if err = json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Cases) != 52 {
		t.Fatal("missing registered accent policies")
	}
	q, err := os.ReadFile("testdata/accent_policy_unchanged_internal.json")
	if err != nil {
		t.Fatal(err)
	}
	var boundary struct {
		Base   string
		SHA256 map[string]string
	}
	if err = json.Unmarshal(q, &boundary); err != nil {
		t.Fatal(err)
	}
	if boundary.Base != "357d77d6145edb49af124eb68966d39c9adb35c5" || len(boundary.SHA256) != 4 {
		t.Fatal("changed internal boundary")
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			root := c.Tree.node()
			var wrap func(*mml.Node, *wrapper) *wrapper
			wrap = func(n *mml.Node, parent *wrapper) *wrapper {
				w := &wrapper{node: n, parent: parent}
				for _, child := range n.Children {
					w.children = append(w.children, wrap(child, w))
				}
				return w
			}
			wrapped := wrap(root, nil)
			for _, o := range c.Observations {
				if o.Kind != "mo" {
					continue
				}
				w := wrapped
				for _, i := range o.Path {
					w = w.children[i]
				}
				if o.IsAccent == nil || w.isAccentMO() != *o.IsAccent {
					t.Errorf("path %v isAccent=%v, want %v", o.Path, w.isAccentMO(), o.IsAccent)
				}
			}
			before, _, _ := msInputSnapshot(root)
			options := pipeline.DefaultOptions()
			options.Display = c.Display
			actual, err := NewTypesetter().Typeset(root, options)
			if err != nil {
				t.Fatal(err)
			}
			if dir := os.Getenv("MATHJAX_GO_ACCENT_EVIDENCE"); dir != "" {
				if err = os.MkdirAll(dir, 0755); err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(filepath.Join(dir, c.Name+".svg"), []byte(actual), 0644); err != nil {
					t.Fatal(err)
				}
			}
			if !accentPolicySVGMatches(c.Name, actual, c.SHA256, boundary.SHA256) {
				t.Errorf("complete SVG=%x, want primary %s or the named exact boundary", sha256.Sum256([]byte(actual)), c.SHA256)
			}
			if c.Name == "both-mixed-inline" || c.Name == "both-mixed-display" {
				const tag = `<g data-mml-node="mo" transform="translate(313.8,31) translate(-250 0)">`
				if strings.Count(actual, tag) != 1 {
					t.Fatal("missing equivalent spelling control")
				}
				for name, bad := range map[string]string{
					"offset":    strings.Replace(actual, "translate(-250 0)", "translate(-249 0)", 1),
					"duplicate": strings.Replace(actual, tag, tag+tag, 1),
					"path":      strings.Replace(actual, `data-c="`, `data-c="0`, 1),
					"root":      strings.Replace(actual, `viewBox="`, `viewBox="1 `, 1),
				} {
					if accentPolicySVGMatches(c.Name, bad, c.SHA256, boundary.SHA256) {
						t.Errorf("accepted %s mutation", name)
					}
				}
			}
			after, _, _ := msInputSnapshot(root)
			if before != after {
				t.Fatal("renderer changed original MathML")
			}
			repeat, err := NewTypesetter().Typeset(root.Clone(), options)
			if err != nil || repeat != actual {
				t.Fatalf("clone changed output: %v", err)
			}
		})
	}
}
