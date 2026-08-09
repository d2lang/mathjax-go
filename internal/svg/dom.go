// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

// Package svg implements MathJax 3.2.2's SVG output jax.
package svg

import (
	"strings"
)

// Node is a deterministic in-memory SVG node.
type Node interface {
	appendXML(*strings.Builder)
}

// Attribute is one ordered XML attribute.
type Attribute struct {
	Name  string
	Value string
}

// Style is one ordered CSS declaration.
type Style struct {
	Name  string
	Value string
}

// Element is an SVG element. Attribute and style insertion order is retained
// because LiteAdaptor serialization order is part of the frozen D2 oracle.
type Element struct {
	Tag        string
	Attributes []Attribute
	Styles     []Style
	Children   []Node

	stylesAfterAttributes bool
}

// NewElement creates an element and appends children in order.
func NewElement(tag string, children ...Node) *Element {
	return &Element{Tag: tag, Children: append([]Node(nil), children...)}
}

// SetAttr adds or replaces an attribute without changing its position.
func (e *Element) SetAttr(name, value string) *Element {
	for i := range e.Attributes {
		if e.Attributes[i].Name == name {
			e.Attributes[i].Value = value
			return e
		}
	}
	e.Attributes = append(e.Attributes, Attribute{Name: name, Value: value})
	return e
}

// RemoveAttr removes an attribute, retaining the order of all others.
func (e *Element) RemoveAttr(name string) bool {
	for i := range e.Attributes {
		if e.Attributes[i].Name == name {
			copy(e.Attributes[i:], e.Attributes[i+1:])
			e.Attributes = e.Attributes[:len(e.Attributes)-1]
			return true
		}
	}
	return false
}

// SetStyle adds or replaces a CSS declaration without changing its position.
func (e *Element) SetStyle(name, value string) *Element {
	for i := range e.Styles {
		if e.Styles[i].Name == name {
			e.Styles[i].Value = value
			return e
		}
	}
	e.Styles = append(e.Styles, Style{Name: name, Value: value})
	return e
}

// StylesAfterAttributes makes the serialized style attribute follow the
// ordinary attributes.  LiteAdaptor preserves the order in which properties
// and attributes are installed, and a few MathJax wrappers (notably mtext's
// explicit-font path) install their style after standardSVGnode has added
// data-mml-node.
func (e *Element) StylesAfterAttributes() *Element {
	e.stylesAfterAttributes = true
	return e
}

// Append appends children.
func (e *Element) Append(children ...Node) *Element {
	e.Children = append(e.Children, children...)
	return e
}

// Prepend inserts children before the existing children.
func (e *Element) Prepend(children ...Node) *Element {
	all := make([]Node, 0, len(children)+len(e.Children))
	all = append(all, children...)
	all = append(all, e.Children...)
	e.Children = all
	return e
}

// Text is escaped XML character data.
type Text string

func (t Text) appendXML(builder *strings.Builder) { escapeText(builder, string(t)) }

func (e *Element) appendXML(builder *strings.Builder) {
	builder.WriteByte('<')
	builder.WriteString(e.Tag)
	if !e.stylesAfterAttributes {
		e.appendStyle(builder)
	}
	for _, attribute := range e.Attributes {
		builder.WriteByte(' ')
		builder.WriteString(attribute.Name)
		builder.WriteString(`="`)
		escapeAttribute(builder, attribute.Value)
		builder.WriteByte('"')
	}
	if e.stylesAfterAttributes {
		e.appendStyle(builder)
	}
	builder.WriteByte('>')
	for _, child := range e.Children {
		if child != nil {
			child.appendXML(builder)
		}
	}
	builder.WriteString("</")
	builder.WriteString(e.Tag)
	builder.WriteByte('>')
}

func (e *Element) appendStyle(builder *strings.Builder) {
	if len(e.Styles) == 0 {
		return
	}
	builder.WriteString(` style="`)
	for i, style := range e.Styles {
		if i != 0 {
			builder.WriteByte(' ')
		}
		builder.WriteString(style.Name)
		builder.WriteString(": ")
		escapeAttribute(builder, style.Value)
		builder.WriteByte(';')
	}
	builder.WriteByte('"')
}

// String serializes e exactly once without indentation or a trailing newline.
func (e *Element) String() string {
	var builder strings.Builder
	e.appendXML(&builder)
	return builder.String()
}

func escapeText(builder *strings.Builder, value string) {
	for _, r := range value {
		switch r {
		case '&':
			builder.WriteString("&amp;")
		case '<':
			builder.WriteString("&lt;")
		case '>':
			builder.WriteString("&gt;")
		default:
			builder.WriteRune(r)
		}
	}
}

func escapeAttribute(builder *strings.Builder, value string) {
	for _, r := range value {
		switch r {
		case '&':
			builder.WriteString("&amp;")
		case '"':
			builder.WriteString("&quot;")
		case '<':
			builder.WriteString("&lt;")
		case '>':
			builder.WriteString("&gt;")
		default:
			builder.WriteRune(r)
		}
	}
}
