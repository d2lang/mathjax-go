// Copyright (c) 2020-2022 MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Source: ts/input/tex/mathtools/MathtoolsConfiguration.ts, fixPrescripts.
package tex

import "github.com/d2lang/mathjax-go/internal/mml"

func fixPrescripts(root *mml.Node) {
	root.Walk(func(n *mml.Node) bool {
		if n.Kind != "mmultiscripts" {
			return true
		}
		if marked, _ := n.Property("fixPrescript"); !limitsTruthy(marked) {
			return true
		}
		children := append([]*mml.Node(nil), n.Children...)
		absent := 0
		for _, index := range []int{1, 2} {
			if children[index] == nil {
				children[index] = node("none")
				absent++
			}
		}
		for _, index := range []int{4, 5} {
			child := children[index]
			if child.Kind == "mrow" && len(child.Children) == 0 {
				children[index] = node("none")
			}
		}
		if absent == 2 {
			children = append(children[:1], children[3:]...)
		}
		n.SetChildren(children)
		refreshDynamicFlags(n)
		return true
	})
}
