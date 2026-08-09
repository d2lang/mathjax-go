// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
//
// This file is a Go translation and modification of MathJax 3.2.2.
// Source: ts/input/tex/Tags.ts, base/BaseMethods.ts, and ams/AmsMethods.ts.

package tex

import (
	"strings"
	"unicode"

	"github.com/d2lang/mathjax-go/internal/mml"
)

type amsLabel struct {
	tag string
	id  string
}

// amsTagInfo corresponds to MathJax's TagInfo.  tag is a pointer because the
// source distinguishes an unset tag (null) from the empty tag installed by
// \notag; that distinction is observable in top-level finalization.
type amsTagInfo struct {
	environment string
	taggable    bool
	defaultTags bool
	tag         *string
	tagID       string
	tagFormat   string
	noTag       bool
	labelID     string
}

// amsTagState is the per-compilation NoTags state used by D2's fixed MathJax
// setup.  NoTags suppresses automatic equation numbers but retains explicit
// \tag, labels/references, and AbstractTags' top-level finalization behavior.
type amsTagState struct {
	current *amsTagInfo
	stack   []*amsTagInfo
	history []*amsTagInfo
	labels  map[string]amsLabel
}

func (p *parser) amsTags() *amsTagState {
	if p.state.amsTags == nil {
		p.state.amsTags = &amsTagState{
			current: &amsTagInfo{},
			labels:  make(map[string]amsLabel),
		}
	}
	return p.state.amsTags
}

func (s *amsTagState) start(environment string, taggable, defaultTags bool) {
	if s.current != nil {
		s.stack = append(s.stack, s.current)
	}
	s.current = &amsTagInfo{
		environment: environment,
		taggable:    taggable,
		defaultTags: defaultTags,
	}
}

func (s *amsTagState) end() {
	if s.current != nil {
		s.history = append(s.history, s.current)
	}
	if len(s.stack) == 0 {
		s.current = &amsTagInfo{}
		return
	}
	s.current = s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
}

func (s *amsTagState) setTag(p *parser, tag string, noFormat bool) {
	s.current.tag = stringPointer(tag)
	if noFormat {
		s.current.tagFormat = tag
	} else {
		s.current.tagFormat = p.mathtoolsFormatTag(tag)
	}
	s.current.noTag = false
}

func (s *amsTagState) notag(p *parser) {
	s.setTag(p, "", true)
	s.current.noTag = true
}

func (s *amsTagState) clearTag() {
	s.current.labelID = ""
	s.current.tag = nil
	s.current.tagID = ""
	s.current.tagFormat = ""
	s.current.noTag = false
}

func stringPointer(value string) *string { return &value }

// amsTagCommand is the narrow command-dispatch seam used by parser.command.
func (p *parser) amsTagCommand(name string) (nodes []*mml.Node, handled bool, err error) {
	switch name {
	case "tag":
		err = p.amsHandleTag(name)
	case "notag", "nonumber":
		p.amsTags().notag(p)
	case "label":
		err = p.amsHandleLabel(name)
	case "ref", "refeq", "eqref":
		nodes, err = p.amsHandleReference(name, name == "eqref")
	default:
		return nil, false, nil
	}
	return nodes, true, err
}

func (p *parser) amsHandleTag(name string) error {
	state := p.amsTags()
	current := state.current
	if !current.taggable && current.environment != "" {
		return texError("CommandNotAllowedInEnv", "\\%s not allowed in %s environment", name, current.environment)
	}
	// A null or empty string is false in the source HandleTag test, so a tag
	// after \notag (or after \tag{}) is allowed.
	if current.tag != nil && *current.tag != "" {
		return texError("MultipleCommand", "Multiple \\%s", name)
	}
	star := p.readStar()
	raw, _, err := p.readArgument(name, false)
	if err != nil {
		return err
	}
	state.setTag(p, strings.TrimSpace(raw), star)
	return nil
}

func (p *parser) amsHandleLabel(name string) error {
	label, _, err := p.readArgument(name, false)
	if err != nil {
		return err
	}
	if label == "" {
		return nil
	}
	state := p.amsTags()
	if state.current.labelID != "" {
		return texError("MultipleCommand", "Multiple \\%s", name)
	}
	state.current.labelID = label
	if _, exists := state.labels[label]; exists {
		return texError("MultipleLabel", "Label '%s' multiply defined", label)
	}
	// MathJax installs a placeholder immediately.  In NoTags mode a label on
	// an untagged equation remains ???, which is exactly what a later \eqref
	// in the same D2 expression observes.
	state.labels[label] = amsLabel{tag: "???"}
	return nil
}

func (p *parser) amsHandleReference(name string, equationReference bool) ([]*mml.Node, error) {
	label, _, err := p.readArgument(name, false)
	if err != nil {
		return nil, err
	}
	reference, exists := p.amsTags().labels[label]
	if !exists {
		reference = amsLabel{tag: "???"}
	}
	tag := reference.tag
	if equationReference {
		tag = p.mathtoolsFormatTag(tag)
	}
	result := node("mrow", textRow(tag))
	result.Attributes.Set("href", "#"+encodeURIComponent(reference.id))
	result.Attributes.Set("class", "MathJax_ref")
	return []*mml.Node{result}, nil
}

// amsTagFinalize ports AbstractTags.finalize.  It intentionally makes an
// empty label cell for a top-level \notag: the source checks tag !== null,
// rather than truthiness, on this path.
func (p *parser) amsTagFinalize(children []*mml.Node) ([]*mml.Node, error) {
	state := p.amsTags()
	if !p.display || state.current.environment != "" || state.current.tag == nil {
		return children, nil
	}
	tag := state.makeTag()
	content := row(children, true)
	return []*mml.Node{amsEnTag(content, tag)}, nil
}

// getTag is the NoTags override followed by AbstractTags.getTag.  Automatic
// numbering is deliberately absent; only a non-empty explicit tag survives.
func (s *amsTagState) getTag() *mml.Node {
	if s.current.tag == nil || *s.current.tag == "" {
		return nil
	}
	if !s.current.taggable || s.current.noTag {
		return nil
	}
	return s.makeTag()
}

func (s *amsTagState) makeTag() *mml.Node {
	current := s.current
	idSource := ""
	if current.labelID != "" {
		idSource = current.labelID
	} else if current.tag != nil {
		idSource = *current.tag
	}
	current.tagID = "mjx-eqn:" + replaceTagWhitespace(idSource)
	if current.labelID != "" {
		tag := ""
		if current.tag != nil {
			tag = *current.tag
		}
		s.labels[current.labelID] = amsLabel{tag: tag, id: current.tagID}
	}

	var cell *mml.Node
	if current.tagFormat == "" {
		cell = node("mtd")
	} else {
		cell = node("mtd", textRow(current.tagFormat))
	}
	cell.Attributes.Set("id", current.tagID)
	return cell
}

func amsEnTag(content, tag *mml.Node) *mml.Node {
	contentCell := node("mtd", content)
	labelled := node("mlabeledtr", tag, contentCell)
	table := node("mtable", labelled)
	// Preserve NodeFactory option insertion order from AbstractTags.enTag.
	table.Attributes.Set("side", "right")
	table.Attributes.Set("minlabelspacing", "0.8em")
	table.Attributes.Set("displaystyle", true)
	return table
}

func replaceTagWhitespace(value string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || r == '\uFEFF' {
			return '_'
		}
		return r
	}, value)
}

// encodeURIComponent is the UTF-8 percent encoding used by Tags.formatUrl.
// Keeping it local avoids the subtly different escaping sets of net/url's
// query and path helpers.
func encodeURIComponent(value string) string {
	const hex = "0123456789ABCDEF"
	var encoded strings.Builder
	for _, b := range []byte(value) {
		if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') ||
			strings.ContainsRune("-_.!~*'()", rune(b)) {
			encoded.WriteByte(b)
			continue
		}
		encoded.WriteByte('%')
		encoded.WriteByte(hex[b>>4])
		encoded.WriteByte(hex[b&0x0F])
	}
	return encoded.String()
}
