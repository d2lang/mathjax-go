// Copyright 2018-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Source: ts/core/MmlTree/MmlNodes/mfenced.ts.

package mml

import (
	"fmt"
	"strings"
)

// CreateFencedNodes ports MmlMfenced.addFakeNodes/fakeNode. Its caller must
// apply the original incoming inheritance context to the returned operators,
// rather than copying the resolved attributes of their mfenced parent.
func (f *Factory) CreateFencedNodes(node *Node) {
	characters := func(name, fallback string) string {
		text := fallback
		if value, ok := node.Attributes.Get(name); ok {
			text = fmt.Sprint(value)
		}
		return strings.Map(func(r rune) rune {
			switch r {
			case ' ', '\t', '\n', '\r':
				return -1
			}
			return r
		}, text)
	}
	fake := func(text, form string, class TeXClass) *Node {
		op := f.Create("mo", f.Text(text))
		if form != "" {
			op.Attributes.Set("fence", true)
			op.Attributes.Set("form", form)
		}
		// The source sets the field only. Operator inheritance may replace it;
		// a texClass property would incorrectly suppress that source behavior.
		op.TeXClass = class
		op.Parent = node
		return op
	}
	if node.Fenced == nil {
		node.Fenced = &FencedNodes{}
	}
	fences := node.Fenced
	if open := characters("open", "("); open != "" {
		fences.Open = fake(open, "prefix", TeXClassOpen)
	}
	if separators := []rune(characters("separators", ",")); len(separators) != 0 {
		index := 0
		for _, child := range node.Children[min(1, len(node.Children)):] {
			if child == nil {
				continue
			}
			separator := separators[min(index, len(separators)-1)]
			fences.Separators = append(fences.Separators, fake(string(separator), "", TeXClassNone))
			index++
		}
	}
	if close := characters("close", ")"); close != "" {
		fences.Close = fake(close, "postfix", TeXClassClose)
	}
}
