// Copyright (c) 2009-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Sources: ts/input/tex/{FilterUtil,ParseOptions,ParseUtil}.ts and
// ts/core/MmlTree/MmlNode.ts and ts/core/Tree/Node.ts.

package tex

import (
	"strings"

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/ordered"
)

// The inherited-attribute pass already handles cleanStretchy's stretchy
// attribute and removes its temporary flag. Retain the registered candidates
// first so its ORD wrapper can be added at the source postfilter boundary,
// after moveLimits. ParseUtil.copyNode preserves and re-registers in-lists.
func stretchyCleanupOperators(operators []*mml.Node) []*mml.Node {
	var result []*mml.Node
	seen := make(map[*mml.Node]bool)
	for _, n := range operators {
		fix, _ := n.Property("fixStretchy")
		lists, _ := n.Property("in-lists")
		if !seen[n] && propertyBool(fix) && strings.Contains(","+propertyString(lists)+",", ",fixStretchy,") {
			result = append(result, n)
			// The source removes fixStretchy after the first registered event.
			seen[n] = true
		}
	}
	return result
}

func cleanStretchy(root *mml.Node, operators []*mml.Node) {
	for _, mo := range operators {
		// ParseOptions.getList discards detached construction events.
		attached := false
		for n := mo; n != nil; n = n.Parent {
			if n == root {
				attached = true
				break
			}
		}
		if !attached {
			continue
		}
		symbol, found := lookupOperatorDefinition(textContent(mo), operatorForms(mo))
		if found {
			for _, attribute := range symbol.Properties {
				if attribute.Name == "stretchy" && propertyBool(attribute.Value) {
					mo.Attributes.Set("stretchy", false)
				}
			}
		}
		if mo.TeXClass != mml.TeXClassOrd || found && symbol.TexClass != int(mml.TeXClassOrd) {
			continue
		}
		parent := mo.Parent
		atom := node("TeXAtom", mo)
		_ = parent.ReplaceChild(atom, mo)
		// AbstractNode.replaceChild clears the old child's parent even after
		// nodeFactory has inserted that child into the new TeXAtom. Preserve
		// that postfilter-specific state: inheritance and later core-parent
		// queries see a parentless operator. The general Go replacement API
		// deliberately retains ownership when a caller has already moved it.
		mo.Parent = nil
		// AbstractMmlNode.inheritAttributesFrom copies only the current size,
		// display/script level and prime style, then inherits through the new
		// wrapper. Font/color attributes remain on the original token.
		attributes := ordered.New[inheritedAttribute]()
		if mo.Attributes.IsSet("mathsize") {
			size, _ := mo.Attributes.Get("mathsize")
			attributes.Set("mathsize", inheritedAttribute{source: "math", value: size})
		}
		display, _ := mo.Attributes.Get("displaystyle")
		level, _ := propertyIntValue(mo.Attributes, "scriptlevel")
		prime, _ := mo.Property("texprimestyle")
		setInheritedAttributes(atom, attributes, propertyBool(display), level, propertyBool(prime))
		for n := parent; n != nil; n = n.Parent {
			refreshDynamicFlags(n)
		}
	}
}
