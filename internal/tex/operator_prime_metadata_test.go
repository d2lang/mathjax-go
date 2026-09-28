// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

type operatorPrimeMetadata struct {
	Text             string         `json:"text"`
	Properties       map[string]any `json:"properties"`
	InheritedSpacing map[string]any `json:"inheritedSpacing,omitempty"`
	RawParent        *string        `json:"rawParent"`
	LogicalParent    *string        `json:"logicalParent"`
	CoreParent       string         `json:"coreParent"`
}

func operatorPrimeNodeKind(n *mml.Node) *string {
	if n == nil {
		return nil
	}
	kind := n.Kind
	if kind == "mrow" && n.Flags.Inferred {
		kind = "inferredMrow"
	}
	return &kind
}

func TestOperatorPrimeOriginalMetadata(t *testing.T) {
	data, err := os.ReadFile("../../testdata/operator_primes_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		MathjaxGitCommit string
		Cases            []struct {
			Name, TeX string
			Display   bool
			Operators *[]operatorPrimeMetadata
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" {
		t.Fatal("unbound operator metadata")
	}
	count := 0
	for _, c := range fixture.Cases {
		if c.Operators == nil {
			continue
		}
		count++
		t.Run(c.Name, func(t *testing.T) {
			root, err := NewCompiler().Compile(c.TeX, c.Display)
			if err != nil {
				t.Fatal(err)
			}
			actual := []operatorPrimeMetadata{}
			root.Walk(func(n *mml.Node) bool {
				if n.Kind != "mo" {
					return true
				}
				o := operatorPrimeMetadata{Text: textContent(n), Properties: map[string]any{}}
				core := n
				for parent := core.Parent; parent != nil && parent.Kind != "math" && parent.Flags.Embellished && limitsCore(parent) == n; parent = core.Parent {
					core = parent
				}
				o.RawParent = operatorPrimeNodeKind(n.Parent)
				o.LogicalParent = operatorPrimeNodeKind(core.ParentNode())
				o.CoreParent = *operatorPrimeNodeKind(core)
				for _, key := range []string{"pseudoscript", "primes"} {
					if value, ok := n.Property(key); ok {
						o.Properties[key] = value
					}
				}
				if o.Properties["pseudoscript"] == true {
					o.InheritedSpacing = map[string]any{}
					for _, key := range []string{"lspace", "rspace"} {
						if value, ok := n.Attributes.GetInherited(key); ok {
							o.InheritedSpacing[key] = value
						}
					}
				}
				actual = append(actual, o)
				return true
			})
			// Use JSON's numeric representation for inherited spacing on both sides.
			encoded, err := json.Marshal(actual)
			if err != nil {
				t.Fatal(err)
			}
			var canonical []operatorPrimeMetadata
			if err := json.Unmarshal(encoded, &canonical); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(canonical, *c.Operators) {
				t.Fatalf("original raw-token/prime metadata differs\ngot: %s\nwant: %#v", encoded, *c.Operators)
			}
		})
	}
	if count != 3172 {
		t.Fatalf("metadata inventory changed: %d", count)
	}
}
