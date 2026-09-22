// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// Based on BaseMethods.Limits and the node-specific MmlNode.coreMO methods.
package tex

import (
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"github.com/d2lang/mathjax-go/internal/mml"
)

// limitsCore follows the source node's coreMO delegation. Its result need not
// be an operator: a nonembellished row is deliberately returned as itself.
// Keep this parser policy separate from the existing generic tree helpers.
func limitsCore(n *mml.Node) *mml.Node {
	for n != nil {
		index := 0
		switch n.Kind {
		case "mrow":
			if !n.Flags.Embellished {
				return n
			}
			index = n.Flags.CoreIndex
		case "msub", "msup", "msubsup", "munder", "mover", "munderover", "mmultiscripts", "TeXAtom", "mstyle", "mpadded", "mphantom", "semantics", "mfrac", "math", "mtd":
		case "maction":
			value, exists := n.Attributes.Get("selection")
			selection := limitsSelection(value, exists)
			selected := math.Max(1, math.Min(float64(len(n.Children)), selection)) - 1
			if math.IsNaN(selected) || selected != math.Trunc(selected) || selected < 0 || selected >= float64(len(n.Children)) || n.Children[int(selected)] == nil {
				// MmlMaction.selected returns a new empty mrow when the
				// numeric index does not address a child.
				return texMMLFactory.Create("mrow")
			}
			index = int(selected)
		default:
			return n
		}
		if index < 0 || index >= len(n.Children) || n.Children[index] == nil {
			return n
		}
		n = n.Children[index]
	}
	return nil
}

func (p *parser) setLimits(nodes []*mml.Node, name string) ([]*mml.Node, error) {
	var op *mml.Node
	if len(nodes) != 0 {
		op = nodes[len(nodes)-1]
	}
	core := limitsCore(op)
	var moves any
	if op != nil {
		moves, _ = op.Property("movesupsub")
	}
	if op == nil || core == nil || (core.TeXClass != mml.TeXClassOp && moves == nil) {
		return nil, texError("MisplacedLimits", "%s is allowed only on operators", "\\"+name)
	}
	enabled := name == "limits"
	kind := ""
	switch op.Kind {
	case "munder", "mover", "munderover":
		if !enabled {
			kind = "msubsup"
		}
	case "msub", "msup", "msubsup":
		if enabled {
			kind = "munderover"
		}
	}
	if kind != "" {
		// NodeUtil.copyChildren reuses child identities and reparents them; it
		// does not copy the old wrapper's attributes or properties.
		replacement := texMMLFactory.Create(kind)
		replacement.SetChildren(op.Children)
		refreshDynamicFlags(replacement)
		op = replacement
		nodes[len(nodes)-1] = op
	}
	op.SetProperty("movesupsub", enabled)
	applySourceObject(limitsCore(op), mjSourceObject{{Name: "movablelimits", Value: false}})
	attr, _ := op.Attributes.Get("movablelimits")
	prop, _ := op.Property("movablelimits")
	if limitsTruthy(attr) || limitsTruthy(prop) {
		applySourceObject(op, mjSourceObject{{Name: "movablelimits", Value: false}})
	}
	return nodes, nil
}

func limitsTruthy(value any) bool {
	switch v := value.(type) {
	case nil:
		return false
	case bool:
		return v
	case string:
		return v != ""
	case int:
		return v != 0
	case int64:
		return v != 0
	case float64:
		return v != 0 && v == v
	default:
		return true
	}
}

// limitsSelection applies the numeric coercions used by the maction selected
// getter to the scalar attribute types supported by this parser.
func limitsSelection(value any, exists bool) float64 {
	if !exists {
		return math.NaN()
	}
	switch v := value.(type) {
	case nil:
		return 0
	case bool:
		if v {
			return 1
		}
		return 0
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case float64:
		return v
	case string:
		text := strings.TrimFunc(v, limitsNumberSpace)
		if text == "" {
			return 0
		}
		if len(text) > 2 && text[0] == '0' {
			base := 0
			switch text[1] {
			case 'x', 'X':
				base = 16
			case 'b', 'B':
				base = 2
			case 'o', 'O':
				base = 8
			}
			if base != 0 {
				for _, digit := range text[2:] {
					value := int(digit - '0')
					if digit >= 'a' && digit <= 'f' {
						value = int(digit-'a') + 10
					} else if digit >= 'A' && digit <= 'F' {
						value = int(digit-'A') + 10
					}
					if value < 0 || value >= base {
						return math.NaN()
					}
				}
				integer, ok := new(big.Int).SetString(text[2:], base)
				if !ok {
					return math.NaN()
				}
				number, _ := integer.Float64()
				return number
			}
		}
		if limitsDecimalNumber.MatchString(text) {
			// A valid overflowing decimal converts to Infinity. ParseFloat's
			// range error must not turn it into NaN.
			number, _ := strconv.ParseFloat(text, 64)
			return number
		}
	}
	return math.NaN()
}

// StringNumericLiteral accepts decimal/Infinity and unsigned radix literals;
// Go's parser additionally accepts Inf, hex floats and underscores, so validate
// the decimal grammar before conversion. Whitespace is the ECMAScript set,
// including BOM and excluding Unicode NEL (U+0085).
var limitsDecimalNumber = regexp.MustCompile(`^[+-]?(?:Infinity|(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)(?:[eE][+-]?[0-9]+)?)$`)

func limitsNumberSpace(r rune) bool {
	return r == '\t' || r == '\n' || r == '\v' || r == '\f' || r == '\r' || r == ' ' ||
		r == '\u00a0' || r == '\u1680' || (r >= '\u2000' && r <= '\u200a') ||
		r == '\u2028' || r == '\u2029' || r == '\u202f' || r == '\u205f' || r == '\u3000' || r == '\ufeff'
}

// This parser compacts script nodes as it creates them, before MathJax's final
// filter would. Remember that origin only until Compile's existing cleanup.
const limitsScriptOrigin = "_texLimitsScriptOrigin"

// parseLimits bridges only parser-created eager script nodes to the source
// handler's full family/slot representation. Authored annotation nodes enter
// setLimits unchanged, including their original child positions.
func (p *parser) parseLimits(nodes []*mml.Node, name string) ([]*mml.Node, error) {
	if len(nodes) != 0 {
		op := nodes[len(nodes)-1]
		if origin, _ := op.Property(limitsScriptOrigin); origin == true && len(op.Children) == 2 {
			children := []*mml.Node{op.Children[0], op.Children[1]}
			switch op.Kind {
			case "msub", "munder":
				children = append(children, nil)
			case "msup", "mover":
				children = []*mml.Node{op.Children[0], nil, op.Children[1]}
			default:
				children = nil
			}
			if children != nil {
				if op.Kind == "msub" || op.Kind == "msup" {
					op.Kind = "msubsup"
				} else {
					op.Kind = "munderover"
				}
				op.Flags.Arity = 3
				op.SetChildren(children)
				refreshDynamicFlags(op)
			}
		}
	}
	result, err := p.setLimits(nodes, name)
	if err != nil {
		return nil, err
	}
	op := result[len(result)-1]
	if op.Kind == "msubsup" || op.Kind == "munderover" {
		var under, over *mml.Node
		if len(op.Children) > 1 {
			under = op.Children[1]
		}
		if len(op.Children) > 2 {
			over = op.Children[2]
		}
		if len(op.Children) > 0 && (under == nil || over == nil) && (under != nil || over != nil) {
			mark := under
			if under == nil {
				mark = over
			}
			kind := "msub"
			if op.Kind == "munderover" {
				kind = "munder"
			}
			if under == nil {
				if kind == "msub" {
					kind = "msup"
				} else {
					kind = "mover"
				}
			}
			op.Kind = kind
			op.Flags.Arity = 2
			op.SetChildren([]*mml.Node{op.Children[0], mark})
			refreshDynamicFlags(op)
			op.SetProperty(limitsScriptOrigin, true)
		}
	}
	return result, nil
}
