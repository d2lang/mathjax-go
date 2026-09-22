package tex

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// The complete constructor snapshots come from the accepted predecessor.
// ParseUtil.underOver assigns subsupOK on the node it returns, after either
// constructing the ordinary annotation or wrapping a brace in an OP atom.
// The public pinned suite separately checks the resulting primary output.
func TestDecorationPermissionConstructorBoundary(t *testing.T) {
	data, err := os.ReadFile("../../testdata/decoration_permission_constructor_baseline.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Baseline string
		Cases    []struct {
			Name, Command, Source string
			Display               bool
			Tree                  map[string]any
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Baseline != "efd6d63071862218af992a6074e3c8680b634ab4" || len(fixture.Cases) != 48 {
		t.Fatal("unbound constructor baseline")
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			p := &parser{source: c.Source, display: c.Display, state: newParseState()}
			out, err := p.underOver(c.Command)
			if err != nil || len(out) != 1 {
				t.Fatalf("constructor: %v", err)
			}
			returned := out[0]
			if v, ok := returned.Property("subsupOK"); !ok || v != true {
				t.Fatal("returned node lacks own primary script permission")
			}
			if _, ok := returned.Attributes.GetExplicit("subsupOK"); ok {
				t.Fatal("permission is a property, not a MathML attribute")
			}
			brace := strings.HasSuffix(c.Command, "brace")
			props := c.Tree["properties"].(map[string]any)
			if brace {
				if props["subsupOK"] != true || props["movesupsub"] != true || returned.Kind != "TeXAtom" {
					t.Fatal("accepted brace permission/OP boundary changed")
				}
			} else {
				if _, exists := props["subsupOK"]; exists {
					t.Fatal("baseline unexpectedly already permits scripts")
				}
				props["subsupOK"] = true
			}
			var project func(*mml.Node) any
			project = func(n *mml.Node) any {
				attrs, properties := map[string]any{}, map[string]any{}
				n.Attributes.Explicit().Range(func(k string, v any) bool { attrs[k] = v; return true })
				n.Properties.Range(func(k string, v any) bool { properties[k] = v; return true })
				children := []any{}
				for _, child := range n.Children {
					if child.Parent != n {
						t.Fatal("child parent identity changed")
					}
					children = append(children, project(child))
				}
				var text any
				if n.Kind == "text" {
					text = n.Text
				}
				kind := n.Kind
				if n.Flags.Inferred {
					kind = "inferredMrow"
				}
				return map[string]any{"kind": kind, "text": text, "attributes": attrs, "properties": properties, "children": children}
			}
			bytes, err := json.Marshal(project(returned))
			if err != nil {
				t.Fatal(err)
			}
			var actual map[string]any
			if err = json.Unmarshal(bytes, &actual); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(actual, c.Tree) {
				t.Fatal("constructor changed more than the ordinary returned node's own permission")
			}
		})
	}
}
