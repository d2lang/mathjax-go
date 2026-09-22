package tex

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

func TestScriptAttachmentPrimaryMethods(t *testing.T) {
	data, err := os.ReadFile("testdata/script_attachment_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Rows []struct {
			Type, Kind, Marker, CompactKind string
			Moves                           *bool
			Allowed                         bool
			Error                           *string
			Reused, BaseRetainedWhole       bool
			Children                        []string
			ChildParents                    []bool
			Attributes, Properties          map[string]any
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Rows) != 120 {
		t.Fatal("missing pinned method/slot/ownership controls")
	}
	for _, c := range fixture.Rows {
		name := c.Type + c.Marker + "/absent"
		if c.Moves != nil {
			name = c.Type + c.Marker + map[bool]string{true: "/true", false: "/false"}[*c.Moves]
		}
		if c.Allowed {
			name += "/subsupOK"
		}
		t.Run(name, func(t *testing.T) {
			core, under, over, script := limitToken("mi", "x"), limitToken("mi", "i"), limitToken("mi", "n"), limitToken("mi", "a")
			children := []*mml.Node{core, under, over}
			switch c.Kind {
			case "msub", "munder":
				children = []*mml.Node{core, under}
			case "msup", "mover":
				children = []*mml.Node{core, over}
			}
			if strings.HasSuffix(c.Type, "no-under") {
				children = []*mml.Node{core, nil, over}
			}
			if strings.HasSuffix(c.Type, "no-over") {
				children = []*mml.Node{core, under}
			}
			base := texMMLFactory.Create(c.Kind)
			base.SetChildren(children)
			refreshDynamicFlags(base)
			base.Attributes.Set("id", "authored")
			base.SetProperty("kept", 17)
			if c.Moves != nil {
				base.SetProperty("movesupsub", *c.Moves)
			}
			if c.Allowed {
				base.SetProperty("subsupOK", true)
			}
			before := base.Clone()
			moves := c.Moves != nil && *c.Moves
			out, err := attachScriptBase(base, script, c.Marker[0], moves)
			if c.Error != nil {
				if err == nil || err.Error() != *c.Error {
					t.Fatalf("error=%v want=%s", err, *c.Error)
				}
				if !reflect.DeepEqual(base, before) {
					t.Fatal("duplicate-script rejection mutated its base")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if out.Kind != c.CompactKind || (out == base) != c.Reused {
				t.Fatalf("kind/identity=%s/%v want=%s/%v", out.Kind, out == base, c.CompactKind, c.Reused)
			}
			ids := map[*mml.Node]string{base: "base", core: "core", under: "under", over: "over", script: "script"}
			got := []string{}
			for i, n := range out.Children {
				got = append(got, ids[n])
				if n.Parent != out || !c.ChildParents[i] {
					t.Fatal("child identity/reparenting changed")
				}
			}
			if !reflect.DeepEqual(got, c.Children) || (out.Children[0] == base) != c.BaseRetainedWhole {
				t.Fatalf("whole base/slots=%v want=%v", got, c.Children)
			}
			// Match JSON scalar types without changing either attribute/property map.
			b, _ := json.Marshal(map[string]any{"attributes": explicitMap(out), "properties": ownMap(out)})
			var maps map[string]map[string]any
			_ = json.Unmarshal(b, &maps)
			if !reflect.DeepEqual(maps["attributes"], c.Attributes) || !reflect.DeepEqual(maps["properties"], c.Properties) {
				t.Fatalf("attrs/properties=%v want=%v/%v", maps, c.Attributes, c.Properties)
			}
		})
	}
}

func TestScriptAttachmentEagerAndAnnotationOwnership(t *testing.T) {
	for _, kind := range []string{"mover", "munder"} {
		for _, origin := range []bool{false, true} {
			base := node(kind, limitToken("mi", "x"), limitToken("mi", "a"))
			if origin {
				base.SetProperty(limitsScriptOrigin, true)
			}
			marker := byte('_')
			if kind == "munder" {
				marker = '^'
			}
			added := limitToken("mi", "i")
			out, err := attachScriptBase(base, added, marker, true)
			if err != nil {
				t.Fatal(err)
			}
			if kind == "mover" && !origin {
				if out.Children[0] != base {
					t.Fatal("authored annotation was split")
				}
			} else if out != base {
				t.Fatal("source reusable family lost its identity")
			}
		}
	}
}
