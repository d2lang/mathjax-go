// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
//
// This file is a Go translation and modification of MathJax 3.2.2.
// Sources: ts/input/tex/{NodeFactory,NodeUtil}.ts and
// ts/core/MmlTree/{MmlNode,OperatorDictionary}.ts;
// MmlNodes/{TeXAtom,maction,math,menclose,merror,mfenced,mglyph,mi,
// mmultiscripts,mn,mo,mpadded,mphantom,mroot,mrow,ms,mspace,msqrt,mstyle,
// msubsup,mtable,mtd,mtext,munderover}.ts.

package tex

import (
	"github.com/d2lang/mathjax-go/internal/mml"
)

func node(kind string, children ...*mml.Node) *mml.Node {
	if hasInferredMrow(kind) {
		if len(children) != 1 || children[0] == nil || children[0].Kind != "mrow" || !children[0].Flags.Inferred {
			children = []*mml.Node{forcedRow(children, true)}
		}
	}
	// AbstractMmlNode.appendChild converts an inferred mrow to an explicit
	// mrow when it is inserted into a fixed-arity node (mfrac, mroot,
	// scripts, ...).  This is observable in SVG: inferred rows borrow their
	// parent's element, while explicit rows own the transform that places a
	// multi-token numerator or denominator.
	if definition, ok := texMMLFactory.Definition(kind); ok && definition.Flags.Arity >= 0 && definition.Flags.Arity != mmlUnboundedArity {
		for i, child := range children {
			if child == nil || !child.Flags.Inferred {
				continue
			}
			explicit := texMMLFactory.Create("mrow", child.Children...)
			explicit.Attributes = child.Attributes.Clone()
			explicit.Properties = child.Properties.Clone()
			explicit.TeXClass = child.TeXClass
			explicit.PrevClass = child.PrevClass
			explicit.PrevLevel = child.PrevLevel
			refreshDynamicFlags(explicit)
			children[i] = explicit
		}
	}
	n := texMMLFactory.Create(kind, children...)
	switch kind {
	case "mi", "mn", "mtext", "ms", "mglyph", "msqrt", "mroot", "merror", "mphantom", "menclose", "mtable":
		n.TeXClass = mml.TeXClassOrd
	case "mfenced":
		n.TeXClass = mml.TeXClassInner
	}
	if kind == "mtable" {
		n.SetProperty("useHeight", true)
	} else if kind == "annotation" {
		n.SetProperty("isChars", true)
	}
	refreshDynamicFlags(n)
	return n
}

// MathJax's MML classes insert an inferred mrow for the variable-arity
// content models below. Keeping it in the Go tree is observable through
// MmlNode.toString(), inherited attributes, and wrapper construction.
func hasInferredMrow(kind string) bool {
	switch kind {
	case "math", "TeXAtom", "mstyle", "merror", "mpadded", "mphantom", "menclose", "mtd", "msqrt":
		return true
	}
	return false
}

func token(kind, text string) *mml.Node {
	n := node(kind, mml.NewText(text))
	n.Flags.Token = true
	switch kind {
	case "mi", "mn", "mtext", "ms":
		n.TeXClass = mml.TeXClassOrd
	case "mo":
		n.Flags.Embellished = true
		n.TeXClass = operatorClass(text)
	}
	return n
}

func operator(text string, class mml.TeXClass, attributes map[string]any) *mml.Node {
	n := token("mo", text)
	n.TeXClass = class
	// Parser-created operators pass a TeX class explicitly through
	// NodeUtil.setProperties.  Preserve that distinction so MmlMo's dynamic
	// dictionary lookup does not overwrite it during inherited-attribute setup.
	n.SetProperty("texClass", class)
	for name, value := range attributes {
		n.Attributes.Set(name, value)
	}
	return n
}

func row(nodes []*mml.Node, inferred bool) *mml.Node {
	if len(nodes) == 1 {
		return nodes[0]
	}
	n := node("mrow", nodes...)
	n.Flags.Inferred = inferred
	n.Flags.NotParent = inferred
	return n
}

func forcedRow(nodes []*mml.Node, inferred bool) *mml.Node {
	n := node("mrow", nodes...)
	n.Flags.Inferred = inferred
	n.Flags.NotParent = inferred
	return n
}

func texAtom(child *mml.Node, class mml.TeXClass) *mml.Node {
	n := node("TeXAtom", child)
	n.TeXClass = class
	n.SetProperty("texClass", class)
	return n
}

func setAttributes(n *mml.Node, attributes map[string]any) *mml.Node {
	for name, value := range attributes {
		n.Attributes.Set(name, value)
	}
	return n
}

func textContent(n *mml.Node) string {
	if n == nil {
		return ""
	}
	if n.Kind == "text" {
		return n.Text
	}
	var text string
	for _, child := range n.Children {
		text += textContent(child)
	}
	return text
}

// refreshDynamicFlags evaluates the MML class getters that depend on children
// or attributes.  The shared tree stores their current values explicitly so
// SVG wrappers do not need a parallel class hierarchy.
func refreshDynamicFlags(n *mml.Node) {
	if n == nil {
		return
	}
	switch n.Kind {
	case "mrow":
		n.Flags.Spacelike = true
		n.Flags.CoreIndex = 0
		core := -1
		validCore := true
		for i, child := range n.Children {
			if child == nil {
				continue
			}
			if !child.Flags.Spacelike {
				n.Flags.Spacelike = false
			}
			if child.Flags.Embellished {
				if core >= 0 {
					validCore = false
				}
				core = i
			} else if !child.Flags.Spacelike {
				validCore = false
			}
		}
		n.Flags.Embellished = validCore && core >= 0
		if n.Flags.Embellished {
			n.Flags.CoreIndex = core
		}

	case "math", "mstyle", "mpadded", "mphantom":
		if len(n.Children) != 0 && n.Children[0] != nil {
			n.Flags.Spacelike = n.Children[0].Flags.Spacelike
			n.Flags.Embellished = n.Children[0].Flags.Embellished
			n.Flags.CoreIndex = 0
		}
		if n.Kind == "mstyle" {
			n.Flags.NotParent = hasSingleInferredChild(n)
		}

	case "TeXAtom":
		if len(n.Children) != 0 && n.Children[0] != nil {
			n.Flags.Embellished = n.Children[0].Flags.Embellished
			n.Flags.CoreIndex = 0
		}
		n.Flags.NotParent = hasSingleInferredChild(n)

	case "msub", "msup", "msubsup", "munder", "mover", "munderover", "mmultiscripts", "mtd":
		if len(n.Children) != 0 && n.Children[0] != nil {
			n.Flags.Embellished = n.Children[0].Flags.Embellished
			n.Flags.CoreIndex = 0
		}

	case "maction":
		if len(n.Children) != 0 && n.Children[0] != nil {
			n.Flags.Embellished = n.Children[0].Flags.Embellished
			n.Flags.Spacelike = n.Children[0].Flags.Spacelike
			n.Flags.CoreIndex = 0
		}

	case "mo":
		linebreak, _ := n.Attributes.Get("linebreak")
		n.Flags.HasNewline = linebreak == "newline"

	case "mspace":
		linebreak, _ := n.Attributes.Get("linebreak")
		_, width := n.Attributes.GetExplicit("width")
		_, height := n.Attributes.GetExplicit("height")
		_, depth := n.Attributes.GetExplicit("depth")
		n.Flags.HasNewline = !width && !height && !depth && linebreak == "newline"
	}
}

func hasSingleInferredChild(n *mml.Node) bool {
	return len(n.Children) != 0 && n.Children[0] != nil && n.Children[0].Flags.Inferred && len(n.Children[0].Children) == 1
}

func operatorClass(text string) mml.TeXClass {
	switch text {
	case "(", "[", "{", "⟨", "⌈", "⌊", "|", "‖":
		return mml.TeXClassOpen
	case ")", "]", "}", "⟩", "⌉", "⌋":
		return mml.TeXClassClose
	case ",", ";":
		return mml.TeXClassPunct
	case "=", "<", ">", "≤", "≥", "≠", "≈", "∼", "∈", "∉", "⊂", "⊃", "⊆", "⊇", "←", "→", "↔", "⇐", "⇒", "⇔", "⟶":
		return mml.TeXClassRel
	case "+", "−", "±", "∓", "×", "÷", "⋅", "∗", "∪", "∩", "⊕", "⊗", "∧", "∨":
		return mml.TeXClassBin
	case "∑", "∏", "∫", "∬", "∭", "∐", "⋃", "⋂":
		return mml.TeXClassOp
	}
	return mml.TeXClassOrd
}
