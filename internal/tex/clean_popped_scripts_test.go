package tex

import (
	"testing"

	"github.com/d2lang/mathjax-go/internal/mml"
)

func TestPoppedScriptCleanupSourcePhase(t *testing.T) {
	for _, tc := range []struct{ family, want string; upper bool }{
		{"msubsup", "msub", false}, {"msubsup", "msup", true},
		{"munderover", "munder", false}, {"munderover", "mover", true},
	} {
		t.Run(tc.want, func(t *testing.T) {
			base, script := token("mi", "x"), token("mn", "1")
			children := []*mml.Node{base, script}
			if tc.upper { children = []*mml.Node{base, nil, script} }
			current := node(tc.family, children...)
			current.Attributes.Set("mathcolor", "red")
			current.SetProperty("movesupsub", true)
			root := node("math", current)
			result, err := cleanPoppedScripts(root, []*mml.Node{current})
			if err != nil { t.Fatal(err) }
			got := result.Children[0].Children[0]
			if got.Kind != tc.want || got.Children[0] != base || got.Children[1] != script {
				t.Fatalf("incorrect compact script: %#v", got)
			}
			if got.Attributes != current.Attributes || got.Properties == current.Properties {
				t.Fatal("copyAttributes must share Attributes and copy Properties")
			}
			if base.Parent != got || script.Parent != got || current.Parent != nil {
				t.Fatal("replacement did not preserve source child/parent ownership")
			}
		})
	}
}

func TestPoppedScriptCopyAndDiscard(t *testing.T) {
	p := &parser{state: newParseState()}
	a, err := prepareScriptAttachment(token("mi", "x"), '^', false)
	if err != nil { t.Fatal(err) }
	view := p.publishPendingScript(a)
	if _, ok := a.base.Property(poppedScriptOrigin); ok { t.Fatal("origin leaked to old base") }
	// An ignored AutoOpen never attaches its unfinished view to the root.
	root := node("math", token("mi", "z"))
	if _, err := cleanPoppedScripts(root, p.state.poppedScripts); err != nil { t.Fatal(err) }
	// A later attachment repairs the same published view before cleanup.
	filled, err := attachScriptBase(view, token("mn", "1"), '_', false)
	if err != nil { t.Fatal(err) }
	copy := p.copyNode(filled)
	root = node("math", copy)
	if len(p.state.poppedScripts) != 2 { t.Fatalf("copy registration count %d", len(p.state.poppedScripts)) }
	if _, err := cleanPoppedScripts(root, p.state.poppedScripts); err != nil { t.Fatal(err) }
	// A live, unfilled view fails only at the explicit cleanup boundary.
	empty := p.publishPendingScript(a)
	root = node("math", empty)
	if _, err := cleanPoppedScripts(root, p.state.poppedScripts); err == nil { t.Fatal("unfinished live script accepted") }
}

func TestPoppedScriptLiveClonedGeneric(t *testing.T) {
	p := &parser{state: newParseState()}
	base, err := attachScriptBase(token("mi", "x"), token("mn", "1"), '_', false)
	if err != nil { t.Fatal(err) }
	base.SetProperty(limitsScriptOrigin, true)
	a, err := prepareScriptAttachment(base, '^', false)
	if err != nil { t.Fatal(err) }
	view := p.publishPendingScript(a)
	copy := p.copyNode(view)
	root := node("math", copy)
	result, err := cleanPoppedScripts(root, p.state.poppedScripts)
	if err != nil { t.Fatal(err) }
	if result.Children[0].Children[0].Kind != "msub" { t.Fatal("live generic clone was not canonicalized") }
	if view.Kind != "msubsup" { t.Fatal("detached original view was changed") }
}
