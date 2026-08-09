// Copyright 2018-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package svg

import (
	"strings"
	"unicode/utf8"

	"github.com/d2lang/mathjax-go/internal/font"
	"github.com/d2lang/mathjax-go/internal/layout"
	"github.com/d2lang/mathjax-go/internal/mml"
)

func fencedCharacters(value string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\t', '\n', '\r':
			return -1
		default:
			return r
		}
	}, value)
}

func fencedToken(text, form string, class mml.TeXClass) *mml.Node {
	node := mml.NewNode("mo", nil, nil, mml.NewText(text))
	node.Flags.Token = true
	node.Flags.Embellished = true
	node.TeXClass = class
	if form != "" {
		node.Attributes.Set("fence", true)
		node.Attributes.Set("form", form)
		if utf8.RuneCountInString(text) == 1 {
			character, _ := utf8.DecodeRuneInString(text)
			if delimiter, ok := font.LookupDelimiter(character); ok && delimiter.Direction == font.DirectionVertical {
				node.Attributes.Set("stretchy", true)
				node.Attributes.Set("symmetric", true)
			}
		}
	} else if text == "|" {
		// The infix operator-dictionary entry for U+007C is an ordinary,
		// symmetric stretchy fence even when mfenced uses it as a separator.
		node.Attributes.Set("fence", true)
		node.Attributes.Set("stretchy", true)
		node.Attributes.Set("symmetric", true)
	}
	return node
}

func fencedSeparatorClass(text string) mml.TeXClass {
	switch text {
	case ",", ".", ";":
		return mml.TeXClassPunct
	case "|", "#", "$", "%", "/", "\\", "^", "_":
		return mml.TeXClassOrd
	case "+", "-", "*":
		return mml.TeXClassBin
	case "?":
		return mml.TeXClassClose
	default:
		// Basic Latin's fallback operator class is REL in 3.2.2's
		// OperatorDictionary; this also covers ':' and comparison signs.
		return mml.TeXClassRel
	}
}

// fencedMrow reconstructs the inferred row and fake mo nodes that MmlMfenced
// owns in MathJax.  Clones keep this compatibility layer side-effect free: the
// real children retain mfenced as their semantic parent just as upstream does
// after its temporary SVG parent swap is restored.
func (w *wrapper) fencedMrow() *wrapper {
	open := fencedCharacters(stringAttribute(w.node, "open", "("))
	close := fencedCharacters(stringAttribute(w.node, "close", ")"))
	separatorText := fencedCharacters(stringAttribute(w.node, "separators", ","))
	separators := []rune(separatorText)

	children := make([]*mml.Node, 0, 2*len(w.node.Children)+2)
	if open != "" {
		children = append(children, fencedToken(open, "prefix", mml.TeXClassOpen))
	}
	for i, child := range w.node.Children {
		if i != 0 && len(separators) != 0 {
			index := i - 1
			if index >= len(separators) {
				index = len(separators) - 1
			}
			separator := string(separators[index])
			children = append(children, fencedToken(separator, "", fencedSeparatorClass(separator)))
		}
		children = append(children, child.Clone())
	}
	if close != "" {
		children = append(children, fencedToken(close, "postfix", mml.TeXClassClose))
	}

	row := mml.NewNode("mrow", nil, nil, children...)
	row.Flags.Inferred = true
	prepareTeXClasses(row)
	return w.renderer.wrap(row, w, w.scriptLevel, w.displayStyle)
}

func (w *wrapper) computeFencedBBox(bbox *layout.BBox) {
	row := w.fencedMrow()
	bbox.UpdateFrom(row.outerBBox())
}

func (w *wrapper) fencedToSVG(parent *Element) {
	element := w.standardSVG(parent)
	row := w.fencedMrow()
	row.toSVG(element)
}
