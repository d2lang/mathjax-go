// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
package tex

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

func explicitMap(n *mml.Node) map[string]any {
	result := map[string]any{}
	for _, k := range n.Attributes.ExplicitNames() {
		result[k], _ = n.Attributes.GetExplicit(k)
	}
	return result
}
func ownMap(n *mml.Node) map[string]any {
	result := map[string]any{}
	for _, k := range n.Properties.Keys() {
		result[k], _ = n.Property(k)
	}
	return result
}
func limitToken(kind, text string) *mml.Node {
	n := texMMLFactory.Create(kind, mml.NewText(text))
	refreshDynamicFlags(n)
	return n
}
func TestLimitsPrimaryPrivateConversions(t *testing.T) {
	b, err := os.ReadFile("testdata/limits_handler_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Rows []struct {
			Kind                                              string
			Enabled                                           bool
			ResultKind                                        string
			Replaced                                          bool
			ChildCount                                        int
			Attributes, Properties, CoreAttrs, CoreProperties map[string]any
		}
	}
	if err = json.Unmarshal(b, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Rows) != 12 {
		t.Fatal("missing primary direct-method rows")
	}
	for _, c := range fixture.Rows {
		t.Run(c.Kind+map[bool]string{true: "/limits", false: "/nolimits"}[c.Enabled], func(t *testing.T) {
			base := limitToken("mo", "∑")
			base.TeXClass = mml.TeXClassOp
			base.Attributes.Set("movablelimits", true)
			children := []*mml.Node{base, limitToken("mi", "i")}
			if c.Kind == "msubsup" || c.Kind == "munderover" {
				children = append(children, limitToken("mi", "n"))
			}
			old := texMMLFactory.Create(c.Kind, children...)
			refreshDynamicFlags(old)
			old.Attributes.Set("id", "old-wrapper")
			old.Attributes.Set("movablelimits", true)
			old.SetProperty("movesupsub", true)
			old.SetProperty("oldProperty", "keep?")
			p := &parser{}
			result, err := p.setLimits([]*mml.Node{old}, map[bool]string{true: "limits", false: "nolimits"}[c.Enabled])
			if err != nil {
				t.Fatal(err)
			}
			out := result[0]
			if out.Kind != c.ResultKind || (out != old) != c.Replaced || len(out.Children) != c.ChildCount {
				t.Fatalf("kind/replacement/arity = %s/%t/%d", out.Kind, out != old, len(out.Children))
			}
			for i, child := range children {
				if out.Children[i] != child || child.Parent != out || old.Children[i] != child {
					t.Fatal("source copyChildren identity/order/old references changed")
				}
			}
			if !reflect.DeepEqual(explicitMap(out), c.Attributes) || !reflect.DeepEqual(ownMap(out), c.Properties) {
				t.Fatalf("wrapper attrs/properties=%v/%v want%v/%v", explicitMap(out), ownMap(out), c.Attributes, c.Properties)
			}
			if !reflect.DeepEqual(explicitMap(base), c.CoreAttrs) || !reflect.DeepEqual(ownMap(base), c.CoreProperties) {
				t.Fatal("core property/attribute clearing differs from primary")
			}
		})
	}
}
func TestLimitsCoreNumericSelection(t *testing.T) {
	b, err := os.ReadFile("testdata/limits_core_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Selected []struct {
			Selection any
			CorePath  *int
			CoreKind  string
		}
	}
	if err = json.Unmarshal(b, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Selected) != 80 {
		t.Fatal("missing primary selections")
	}
	for _, c := range fixture.Selected {
		first, last := limitToken("mi", "x"), limitToken("mo", "∑")
		a := texMMLFactory.Create("maction", first, last)
		a.Attributes.Set("selection", c.Selection)
		got := limitsCore(a)
		if got == nil || got.Kind != c.CoreKind {
			t.Fatalf("selection %v core=%v", c.Selection, got)
		}
		if c.CorePath != nil {
			if got != a.Children[*c.CorePath] {
				t.Fatal("selection identity mismatch")
			}
		} else if len(got.Children) != 0 || got == a || got == first || got == last || limitsCore(a) == got {
			t.Fatal("missing selection must return new empty mrow each time")
		}
		if first.Parent != a || last.Parent != a || len(a.Children) != 2 {
			t.Fatal("core lookup changed original action")
		}
	}
	// Scalar numeric forms share the pinned selected getter's clamp, and NaN
	// never becomes an array index. Object/array attribute coercion is not a
	// parser input contract.
	for _, c := range []struct {
		value any
		index int
	}{{int64(99), 1}, {math.Inf(1), 1}, {math.Inf(-1), 0}, {" 2 ", 1}, {true, 0}, {false, 0}, {nil, 0}, {"", 0}, {math.NaN(), -1}} {
		a := texMMLFactory.Create("maction", limitToken("mi", "x"), limitToken("mo", "∑"))
		a.Attributes.Set("selection", c.value)
		got := limitsCore(a)
		if c.index >= 0 {
			if got != a.Children[c.index] {
				t.Fatalf("scalar selection %v", c.value)
			}
		} else if got.Kind != "mrow" || len(got.Children) != 0 {
			t.Fatal("NaN did not select fresh empty row")
		}
	}
}
func TestLimitsCoreDelegationAndEligibility(t *testing.T) {
	// Core delegation is node-specific: a nonembellished row returns itself,
	// while a base/layout node descends even when it is not embellished.
	x := limitToken("mi", "x")
	x.TeXClass = mml.TeXClassOrd
	sum := limitToken("mo", "∑")
	sum.TeXClass = mml.TeXClassOp
	row := node("mrow", x, sum)
	if row.Flags.Embellished || limitsCore(row) != row {
		t.Fatal("nonembellished row descended")
	}
	row = node("mrow", node("mspace"), sum)
	if limitsCore(row) != sum || row.Flags.CoreIndex != 1 {
		t.Fatal("nonzero embellished core ignored")
	}
	for _, kind := range []string{"msub", "msup", "msubsup", "munder", "mover", "munderover", "mmultiscripts", "TeXAtom", "mstyle", "mpadded", "mphantom", "semantics", "mfrac", "math", "mtd"} {
		base := texMMLFactory.Create(kind, x)
		base.Flags.Embellished = false
		if limitsCore(base) != x {
			t.Fatalf("%s failed direct child-zero delegation", kind)
		}
	}
	for _, kind := range []string{"mi", "mo", "merror", "mfenced", "mtr", "mlabeledtr", "annotation", "annotation-xml"} {
		n := texMMLFactory.Create(kind)
		if limitsCore(n) != n {
			t.Fatalf("%s must return itself", kind)
		}
	}
	for _, value := range []any{false, true, 0, "", 1} {
		n := limitToken("mi", "x")
		n.TeXClass = mml.TeXClassOrd
		n.SetProperty("movesupsub", value)
		p := &parser{}
		out, err := p.setLimits([]*mml.Node{n}, "limits")
		if err != nil || out[0] != n {
			t.Fatalf("non-null movesupsub %v rejected: %v", value, err)
		}
		if v, _ := n.Property("movesupsub"); v != true {
			t.Fatal("enabled value not set")
		}
	}
	for _, null := range []bool{false, true} {
		n := limitToken("mi", "x")
		n.TeXClass = mml.TeXClassOrd
		if null {
			n.SetProperty("movesupsub", nil)
		}
		before := n.Clone()
		p := &parser{}
		if _, err := p.setLimits([]*mml.Node{n}, "limits"); err == nil {
			t.Fatal("ordinary/null operator accepted")
		}
		if !reflect.DeepEqual(n, before) {
			t.Fatal("rejected operator mutated")
		}
	}
	for _, kind := range []string{"mi", "mo", "mstyle", "mrow", "TeXAtom"} {
		n := texMMLFactory.Create(kind)
		n.TeXClass = mml.TeXClassOp
		n.SetProperty("movesupsub", true)
		n.Attributes.Set("movablelimits", true)
		p := &parser{}
		_, err := p.setLimits([]*mml.Node{n}, "nolimits")
		if err != nil {
			t.Fatal(err)
		}
		if v, _ := n.Property("movablelimits"); v != false {
			t.Fatal("outer property not cleared")
		}
		v, _ := n.Attributes.GetExplicit("movablelimits")
		if v != (kind != "mo" && kind != "mstyle") {
			t.Fatal("attribute should be written only on mo/mstyle")
		}
	}
}
func TestLimitsEagerScriptOriginAndCleanup(t *testing.T) {
	p := &parser{source: "n", state: newParseState()}
	base := limitToken("mo", "∑")
	base.TeXClass = mml.TeXClassOp
	base.SetProperty("movesupsub", true)
	nodes, err := p.attachScript([]*mml.Node{base}, '^')
	if err != nil {
		t.Fatal(err)
	}
	original := nodes[0]
	mark := original.Children[1]
	if origin, _ := original.Property(limitsScriptOrigin); origin != true {
		t.Fatal("parser script origin missing")
	}
	clone := original.Clone()
	if origin, _ := clone.Property(limitsScriptOrigin); origin != true {
		t.Fatal("clone lost transient origin")
	}
	nodes, err = p.parseLimits(nodes, "nolimits")
	if err != nil {
		t.Fatal(err)
	}
	if nodes[0].Kind != "msup" || nodes[0].Children[0] != base || nodes[0].Children[1] != mark {
		t.Fatal("generated upper script lost slot/identity")
	}
	nodes, err = p.parseLimits(nodes, "limits")
	if err != nil {
		t.Fatal(err)
	}
	if nodes[0].Kind != "mover" || nodes[0].Children[1] != mark {
		t.Fatal("repeated Limits lost upper slot")
	}
	authoredBase := limitToken("mo", "∑")
	authoredBase.TeXClass = mml.TeXClassOp
	authoredMark := limitToken("mi", "a")
	authored := node("mover", authoredBase, authoredMark)
	if _, ok := authored.Property(limitsScriptOrigin); ok {
		t.Fatal("authored annotation unexpectedly marked")
	}
	nodes, err = p.parseLimits([]*mml.Node{authored}, "nolimits")
	if err != nil {
		t.Fatal(err)
	}
	if nodes[0].Kind != "msub" || nodes[0].Children[1] != authoredMark {
		t.Fatal("authored raw child-one order changed")
	}
	for _, source := range []string{`\sum^n\nolimits\limits_i`, `\overset{a}{\sum_i^n}`, `{\sum_i}\limits^n`, `\underbrace{x}_n\limits`, `\sum_i^n`, `x_i^n`, `\sum'`, `\sum'_i\limits`, `{\sum'}\limits`, `\overset{a}{\sum'}`} {
		for _, display := range []bool{false, true} {
			root, err := NewCompiler().Compile(source, display)
			if err != nil {
				t.Fatal(err)
			}
			root.Walk(func(n *mml.Node) bool {
				if _, ok := n.Property(limitsScriptOrigin); ok {
					t.Fatalf("transient marker leaked: %s", source)
				}
				return true
			})
		}
	}
}

func TestLimitsPendingPrimeLifetime(t *testing.T) {
	p := &parser{state: newParseState()}
	base := limitToken("mo", "∑")
	base.TeXClass = mml.TeXClassOp
	base.SetProperty("movesupsub", true)
	pending, err := p.startPrime(base)
	if err != nil {
		t.Fatal(err)
	}
	if pending.base != base || pending.prime == nil {
		t.Fatal("pending pair lost original identities")
	}
	for _, command := range []string{"limits", "nolimits"} {
		for _, gap := range []string{"", " ", "\\notag"} {
			q := &parser{state: newParseState(), source: "\\sum'" + gap + "\\" + command}
			_, _, got := q.parseRow(0, false)
			_, want := p.setLimits([]*mml.Node{limitToken("mo", "′")}, command)
			if got == nil || want == nil || got.Error() != want.Error() {
				t.Fatalf("pending prime eligibility differs: %v / %v", got, want)
			}
		}
	}
	result, err := pending.attach(limitToken("mi", "i"), '_')
	if err != nil {
		t.Fatal(err)
	}
	consumed := []*mml.Node{result}
	if origin, _ := result.Property(limitsScriptOrigin); origin != true {
		t.Fatal("consumed script origin missing")
	}
	for _, command := range []string{"limits", "nolimits", "limits"} {
		consumed, err = p.parseLimits(consumed, command)
		if err != nil {
			t.Fatal(err)
		}
		if len(consumed[0].Children) != 3 || consumed[0].Children[0] != base || consumed[0].Children[2] != pending.prime {
			t.Fatal("consumed prime slot/identity changed")
		}
	}
}
