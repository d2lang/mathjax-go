// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Sources: FilterUtil.cleanSubSup, ParseOptions.getList/inTree,
// ParseUtil.copyNode and NodeUtil.copyAttributes (MathJax 3.2.2).
package tex

import (
	"fmt"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// Only actual SubsupItem Pop delivery publishes an unfinished generic node.
// Keep the diagnostic pendingBase views read-only. The private origin marker
// follows ParseUtil.copyNode's cloned registrations, including later pruning.
const poppedScriptOrigin = "poppedScriptOrigin"

func (p *parser) publishPendingScript(attachment *scriptAttachment) *mml.Node {
	view := attachment.pendingBase()
	view.Properties = view.Properties.Clone()
	view.SetProperty(poppedScriptOrigin, true)
	view.Children = append([]*mml.Node(nil), view.Children...)
	view.SetChildren(view.Children)
	p.notePoppedScript(view)
	return view
}

func (p *parser) notePoppedScript(n *mml.Node) {
	if marked, _ := n.Property(poppedScriptOrigin); marked == true {
		p.state.poppedScripts = append(p.state.poppedScripts, n)
	}
}

// This is the source parent-chain predicate, not a traversal-membership test.
// Discarded AutoOpen contents must not be validated just because they were
// once registered; their unfinished nodes are no longer live in this root.
func scriptInTree(n, root *mml.Node) bool {
	for n != nil && n != root {
		n = n.Parent
	}
	return n != nil
}

func cleanPoppedScripts(root *mml.Node, registered []*mml.Node) (*mml.Node, error) {
	for _, family := range []struct{ generic, low, up string }{
		{"msubsup", "msub", "msup"},
		{"munderover", "munder", "mover"},
	} {
		for _, current := range registered {
			if current.Kind != family.generic || !scriptInTree(current, root) {
				continue
			}
			var low, up *mml.Node
			if len(current.Children) > 1 {
				low = current.Children[1]
			}
			if len(current.Children) > 2 {
				up = current.Children[2]
			}
			if low != nil && up != nil {
				continue
			}
			if low == nil && up == nil {
				// The original fixed-arity constructor throws here. Preserve
				// its post-successful-parse timing as a bounded API error.
				return nil, fmt.Errorf("incomplete %s has no script child", current.Kind)
			}
			kind, script := family.low, low
			if low == nil {
				kind, script = family.up, up
			}
			parent := current.Parent
			replacement := node(kind, current.Children[0], script)
			replacement.Attributes = current.Attributes
			replacement.Properties = current.Properties.Clone()
			if value, ok := current.Property("texClass"); ok {
				if class, ok := sourceInt(value); ok {
					replacement.TeXClass = mml.TeXClass(class)
				}
			}
			if parent == nil {
				root = replacement
			} else if err := parent.ReplaceChild(replacement, current); err != nil {
				return nil, err
			}
		}
	}
	return root, nil
}
