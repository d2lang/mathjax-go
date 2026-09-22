package tex

import (
	"encoding/json"
	"github.com/d2lang/mathjax-go/internal/mml"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestPendingPrimePrimaryOwnership(t *testing.T) {
	data, err := os.ReadFile("testdata/pending_prime_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Rows []struct {
			Type, Action string
			Moves        *bool
			Allowed      bool
			Error        *string
			Reused       bool
			Tree         map[string]any
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Rows) != 162 {
		t.Fatal("missing actual PrimeItem/SubsupItem controls")
	}
	for _, c := range fixture.Rows {
		name := c.Type + "/" + c.Action
		if c.Moves != nil {
			name += map[bool]string{false: "/false", true: "/true"}[*c.Moves]
		} else {
			name += "/unset"
		}
		if c.Allowed {
			name += "/allowed"
		}
		t.Run(name, func(t *testing.T) {
			core, under, over, script := limitToken("mi", "x"), limitToken("mi", "i"), limitToken("mi", "n"), limitToken("mi", "a")
			kind := strings.Split(c.Type, "-")[0]
			kids := []*mml.Node{core, under, over}
			if kind == "msup" || kind == "mover" {
				kids = []*mml.Node{core, over}
			}
			if kind == "msub" || kind == "munder" || strings.HasSuffix(c.Type, "no-over") {
				kids = []*mml.Node{core, under}
			}
			base := core
			if kind != "mi" {
				base = texMMLFactory.Create(kind)
				base.SetChildren(kids)
				refreshDynamicFlags(base)
				base.Attributes.Set("id", "authored")
			}
			base.SetProperty("kept", 17)
			if c.Moves != nil {
				base.SetProperty("movesupsub", *c.Moves)
			}
			if c.Allowed {
				base.SetProperty("subsupOK", true)
			}
			before := base.Clone()
			p := &parser{state: newParseState()}
			pending, err := p.startPrime(base)
			var out *mml.Node
			if err == nil {
				if c.Action == "finalize" {
					out = pending.finish()
				} else {
					out, err = pending.attach(script, c.Action[0])
				}
			}
			if c.Error != nil {
				if err == nil || err.Error() != *c.Error {
					t.Fatalf("error=%v want=%s", err, *c.Error)
				}
				if !reflect.DeepEqual(base, before) {
					t.Fatal("rejection changed original base")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if (out == base) != c.Reused {
				t.Fatal("original wrapper identity differs")
			}
			ids := map[*mml.Node]string{core: "core", under: "under", over: "over", script: "script", pending.prime: "prime"}
			ids[base] = "base"
			var project func(*mml.Node) any
			project = func(n *mml.Node) any {
				if n == nil {
					return nil
				}
				id := ids[n]
				if id == "" {
					id = "new"
				}
				props := ownMap(n)
				delete(props, limitsScriptOrigin)
				delete(props, vectorFactoryToken)
				k := n.Kind
				kids := n.Children
				// Compact eager output represents the generic source family before its
				// cleanSubSup filter. Expand only parser-owned output, never authored nodes.
				if origin, _ := n.Property(limitsScriptOrigin); origin == true && len(kids) == 2 {
					if k == "msup" {
						k = "msubsup"
						kids = []*mml.Node{kids[0], nil, kids[1]}
					}
					if k == "mover" {
						k = "munderover"
						kids = []*mml.Node{kids[0], nil, kids[1]}
					}
				}
				var text any
				if n.Kind == "text" {
					text = n.Text
				}
				children := []any{}
				for _, ch := range kids {
					if ch != nil && ch.Parent != n {
						t.Fatal("child reparenting lost")
					}
					children = append(children, project(ch))
				}
				return map[string]any{"identity": id, "kind": k, "text": text, "attributes": explicitMap(n), "properties": props, "children": children}
			}
			b, _ := json.Marshal(project(out))
			var got map[string]any
			_ = json.Unmarshal(b, &got)
			// As in the D053 method fixture, compare the shared helper's
			// compact family representation while retaining the source tree.
			// Only a consumed third slot changes the authored subtype name.
			if c.Action != "finalize" && len(c.Tree["children"].([]any)) == 3 {
				if c.Tree["kind"] == "msub" {
					c.Tree["kind"] = "msubsup"
				}
				if c.Tree["kind"] == "munder" {
					c.Tree["kind"] = "munderover"
				}
			}
			if !reflect.DeepEqual(got, c.Tree) {
				t.Fatalf("complete identity/attribute/property/slot tree differs\ngot=%s\nwant=%v", b, c.Tree)
			}
		})
	}
}
