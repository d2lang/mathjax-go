// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package svg

import "testing"

func TestElementSerializationOrderAndEscaping(t *testing.T) {
	path := NewElement("path").
		SetAttr("data-c", `A&B`).
		SetAttr("d", `M0 0"`)
	g := NewElement("g", path).
		SetStyle("vertical-align", "-0.2ex").
		SetStyle("color", `a&b`).
		SetAttr("stroke", "currentColor").
		SetAttr("data-x", "old").
		SetAttr("data-x", "new")
	got := g.String()
	want := `<g style="vertical-align: -0.2ex; color: a&amp;b;" stroke="currentColor" data-x="new"><path data-c="A&amp;B" d="M0 0&quot;"></path></g>`
	if got != want {
		t.Fatalf("serialization:\n got %s\nwant %s", got, want)
	}
}

func TestElementMutation(t *testing.T) {
	e := NewElement("g", Text("b<c"))
	e.Prepend(Text("a&"))
	e.SetAttr("one", "1").SetAttr("two", "2")
	if !e.RemoveAttr("one") || e.RemoveAttr("missing") {
		t.Fatal("RemoveAttr returned the wrong status")
	}
	if got, want := e.String(), `<g two="2">a&amp;b&lt;c</g>`; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestElementCanSerializeStyleAfterAttributes(t *testing.T) {
	e := NewElement("g").
		SetAttr("data-mml-node", "mtext").
		SetStyle("font-family", "serif").
		StylesAfterAttributes()
	if got, want := e.String(), `<g data-mml-node="mtext" style="font-family: serif;"></g>`; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
