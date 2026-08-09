// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package mml

import (
	"testing"

	"github.com/d2lang/mathjax-go/internal/ordered"
)

func TestAttributeLayers(t *testing.T) {
	global := ordered.New[Property]()
	global.Set("mathcolor", "red")
	defaults := ordered.New[Property]()
	defaults.Set("mathcolor", Inherit)
	defaults.Set("display", "inline")
	a := NewAttributes(defaults, global)

	if got, _ := a.Get("mathcolor"); got != "red" {
		t.Fatalf("global inherited mathcolor = %v", got)
	}
	a.SetInherited("display", "block")
	a.Set("display", "inline")
	if got, _ := a.Get("display"); got != "inline" {
		t.Fatalf("explicit display = %v", got)
	}
	if !a.IsSet("display") || !a.HasDefault("display") {
		t.Fatal("display should be set and have a default")
	}
}

func TestNodeTreeAndClone(t *testing.T) {
	root := NewNode("math", nil, nil, NewNode("mi", nil, nil, NewText("x")))
	if root.Children[0].Parent != root || root.Find("text")[0].Text != "x" {
		t.Fatal("parent links or tree search are incorrect")
	}
	clone := root.Clone()
	clone.Find("text")[0].Text = "y"
	if root.Find("text")[0].Text != "x" {
		t.Fatal("clone aliases source children")
	}
	if len(root.Find("mi")[0].Attributes.ExplicitNames()) != 0 {
		t.Fatal("new node has explicit attributes")
	}
}
