// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/svg"
)

type accentSpec struct {
	Kind, Text                        string
	Attributes, Properties, Inherited map[string]any
	Children                          []accentSpec
}

func buildAccentSpec(s accentSpec) *mml.Node {
	if s.Kind == "text" {
		return mml.NewText(s.Text)
	}
	children := make([]*mml.Node, len(s.Children))
	for i, c := range s.Children {
		children[i] = buildAccentSpec(c)
	}
	n := node(s.Kind, children...)
	for k, v := range s.Attributes {
		n.Attributes.Set(k, v)
	}
	for k, v := range s.Properties {
		n.SetProperty(k, v)
	}
	return n
}
func TestAccentInheritancePrimaryStates(t *testing.T) {
	b, err := os.ReadFile("../svg/testdata/accent_policy_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Cases []struct {
			Name, SHA256 string
			Display      bool
			Spec         accentSpec
			Observations []struct {
				Path   []int
				Kind   string
				Values map[string]map[string]any
				Prime  any
			}
		}
	}
	if err = json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Cases) != 52 {
		t.Fatal("missing accent inheritance cases")
	}
	absent := func(v any, ok bool) any {
		if !ok {
			return "__absent__"
		}
		return v
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			root := node("math", buildAccentSpec(c.Spec))
			if c.Display {
				root.Attributes.Set("display", "block")
			}
			setMathMLInheritance(root, c.Display)
			for _, o := range c.Observations {
				n := root
				for _, i := range o.Path {
					if i >= len(n.Children) {
						t.Fatalf("missing path %v", o.Path)
					}
					n = n.Children[i]
				}
				kind := n.Kind
				if kind == "mrow" && n.Flags.Inferred {
					kind = "inferredMrow"
				}
				if kind != o.Kind {
					t.Fatalf("path %v kind=%s want%s", o.Path, kind, o.Kind)
				}
				for key, want := range o.Values {
					got := map[string]any{"explicit": absent(n.Attributes.GetExplicit(key)), "inherited": absent(n.Attributes.GetInherited(key)), "resolved": absent(n.Attributes.Get(key))}
					data, _ := json.Marshal(got)
					json.Unmarshal(data, &got)
					if !reflect.DeepEqual(got, want) {
						t.Errorf("path %v %s=%v want%v", o.Path, key, got, want)
					}
				}
				if got := absent(n.Property("texprimestyle")); !reflect.DeepEqual(got, o.Prime) {
					t.Errorf("path %v prime=%v want%v", o.Path, got, o.Prime)
				}
			}
			if dir := os.Getenv("MATHJAX_GO_ACCENT_INHERITANCE_EVIDENCE"); dir != "" {
				if err = os.MkdirAll(dir, 0755); err != nil {
					t.Fatal(err)
				}
				data, err := json.MarshalIndent(d2RichGoNode(root), "", "  ")
				if err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(filepath.Join(dir, c.Name+".json"), data, 0644); err != nil {
					t.Fatal(err)
				}
				options := pipeline.DefaultOptions()
				options.Display = c.Display
				actual, err := svg.NewTypesetter().Typeset(root, options)
				if err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(filepath.Join(dir, c.Name+".svg"), []byte(actual), 0644); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
