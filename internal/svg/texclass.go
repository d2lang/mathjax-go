// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package svg

import "github.com/d2lang/mathjax-go/internal/mml"

// prepareTeXClasses ports the MmlNode.setTeXclass() traversal.  Its return
// value is deliberately not always the node being visited: transparent rows
// return their final child, left/right rows return their closing delimiter,
// and mfrac returns a node whose TeX class is null.  Those distinctions are
// observable as inter-atom spacing in the SVG output.
func prepareTeXClasses(node *mml.Node) {
	setTeXClass(node, nil)
}

func setTeXClass(node, previous *mml.Node) *mml.Node {
	if node == nil {
		return previous
	}
	switch node.Kind {
	case "text", "XML", "xml", "annotation", "annotation-xml", "mspace", "mprescripts", "none":
		return previous

	case "math", "mstyle", "mpadded", "mphantom":
		if len(node.Children) == 0 {
			return previous
		}
		previous = setTeXClass(node.Children[0], previous)
		updateTeXClass(node, node.Children[0])
		return previous

	case "mrow":
		return setRowTeXClass(node, previous)

	case "mfrac", "msqrt", "mroot", "mtable", "mtr", "mlabeledtr", "mtd":
		setPreviousClass(node, previous)
		for _, child := range node.Children {
			setTeXClass(child, nil)
		}
		return node

	case "menclose":
		if len(node.Children) == 0 {
			return previous
		}
		previous = setTeXClass(node.Children[0], previous)
		updateTeXClass(node, node.Children[0])
		return previous

	case "TeXAtom":
		if len(node.Children) != 0 {
			setTeXClass(node.Children[0], nil)
		}
		return adjustTeXClass(node, previous)

	case "mo":
		return adjustTeXClass(node, previous)

	case "msub", "msup", "msubsup", "munder", "mover", "munderover", "mmultiscripts", "semantics":
		return setBaseTeXClass(node, previous)

	case "maction":
		selected := actionSelection(node)
		if selected == nil {
			return previous
		}
		previous = setTeXClass(selected, previous)
		updateTeXClass(node, selected)
		return previous
	}

	setPreviousClass(node, previous)
	if node.TeXClass == mml.TeXClassNone {
		return previous
	}
	return node
}

func setRowTeXClass(node, previous *mml.Node) *mml.Node {
	_, hasOpen := node.Property("open")
	_, hasClose := node.Property("close")
	if hasOpen || hasClose {
		setPreviousClass(node, previous)
		previous = nil
	}
	for _, child := range node.Children {
		previous = setTeXClass(child, previous)
	}
	if hasOpen || hasClose {
		if node.TeXClass == mml.TeXClassNone {
			node.TeXClass = mml.TeXClassInner
		}
	} else if len(node.Children) != 0 {
		updateTeXClass(node, node.Children[0])
	}
	return previous
}

func setBaseTeXClass(node, previous *mml.Node) *mml.Node {
	setPreviousClass(node, previous)
	node.TeXClass = mml.TeXClassOrd
	if len(node.Children) == 0 || node.Children[0] == nil {
		return node
	}
	base := node.Children[0]
	if node.Flags.Embellished || base.Kind == "mi" {
		previous = setTeXClass(base, previous)
		core := base
		if node.Flags.Embellished {
			core = coreNode(node)
		}
		updateTeXClass(node, core)
	} else {
		setTeXClass(base, nil)
		previous = node
	}
	for _, child := range node.Children[1:] {
		setTeXClass(child, nil)
	}
	return previous
}

func coreNode(node *mml.Node) *mml.Node {
	for node != nil && node.Flags.Embellished && node.Kind != "mo" {
		index := node.Flags.CoreIndex
		if index < 0 || index >= len(node.Children) {
			break
		}
		node = node.Children[index]
	}
	return node
}

func actionSelection(node *mml.Node) *mml.Node {
	if len(node.Children) == 0 {
		return nil
	}
	selection := 1
	if value, ok := numberAttribute(node, "selection"); ok {
		selection = int(value)
	}
	if selection < 1 || selection > len(node.Children) {
		selection = 1
	}
	return node.Children[selection-1]
}

func setPreviousClass(node, previous *mml.Node) {
	node.PrevClass = mml.TeXClassNone
	node.PrevLevel = 0
	if previous == nil {
		return
	}
	node.PrevClass = previous.TeXClass
	if level, ok := numberAttribute(previous, "scriptlevel"); ok {
		node.PrevLevel = int(level)
	}
}

func updateTeXClass(node, core *mml.Node) {
	if node == nil || core == nil {
		return
	}
	node.PrevClass = core.PrevClass
	node.PrevLevel = core.PrevLevel
	core.PrevClass = mml.TeXClassNone
	core.PrevLevel = 0
	node.TeXClass = core.TeXClass
}

func adjustTeXClass(node, previous *mml.Node) *mml.Node {
	class := effectiveTeXClass(node)
	if class == mml.TeXClassNone {
		return previous
	}
	previousClass := mml.TeXClassNone
	if previous != nil {
		previousClass = previous.TeXClass
		if auto, ok := previous.Property("autoOP"); ok && truthy(auto) &&
			(class == mml.TeXClassBin || class == mml.TeXClassRel) {
			previous.TeXClass = mml.TeXClassOrd
			previousClass = mml.TeXClassOrd
		}
	}
	node.PrevClass = previousClass
	node.PrevLevel = 0
	if previous != nil {
		if level, ok := numberAttribute(previous, "scriptlevel"); ok {
			node.PrevLevel = int(level)
		}
	}

	if class == mml.TeXClassBin &&
		(previousClass == mml.TeXClassNone || previousClass == mml.TeXClassBin ||
			previousClass == mml.TeXClassOp || previousClass == mml.TeXClassRel ||
			previousClass == mml.TeXClassOpen || previousClass == mml.TeXClassPunct) {
		node.TeXClass = mml.TeXClassOrd
		class = mml.TeXClassOrd
	} else if previousClass == mml.TeXClassBin &&
		(class == mml.TeXClassRel || class == mml.TeXClassClose || class == mml.TeXClassPunct) {
		previous.TeXClass = mml.TeXClassOrd
		node.PrevClass = mml.TeXClassOrd
	} else if class == mml.TeXClassBin && isLastTeXAtom(node) {
		node.TeXClass = mml.TeXClassOrd
	}
	return node
}

func isLastTeXAtom(node *mml.Node) bool {
	child := node
	parent := child.Parent
	for parent != nil && parent.Parent != nil && parent.Flags.Embellished &&
		(len(parent.Children) == 1 || (parent.Kind != "mrow" && parent.Core() == child)) {
		child = parent
		parent = parent.Parent
	}
	return parent != nil && len(parent.Children) != 0 && parent.Children[len(parent.Children)-1] == child
}

func effectiveTeXClass(node *mml.Node) mml.TeXClass {
	if node == nil {
		return mml.TeXClassNone
	}
	if node.TeXClass != mml.TeXClassNone {
		return node.TeXClass
	}
	// MmlNode.texClass can be supplied as an internal property on any node,
	// not only TeXAtom.  The invisible ApplyFunction mo uses an explicit -1;
	// preserving it prevents spurious OP-to-ORD spacing.
	if value, ok := node.Property("texClass"); ok {
		switch class := value.(type) {
		case mml.TeXClass:
			return class
		case int:
			return mml.TeXClass(class)
		case int64:
			return mml.TeXClass(class)
		case float64:
			return mml.TeXClass(int(class))
		}
	}
	switch node.Kind {
	case "text", "mspace", "mprescripts", "none", "annotation", "annotation-xml", "XML", "xml":
		return mml.TeXClassNone
	default:
		// AbstractMmlNode.texSpacing() treats a null texClass as ORD via
		// `this.texClass || TEXCLASS.ORD`.
		return mml.TeXClassOrd
	}
}
