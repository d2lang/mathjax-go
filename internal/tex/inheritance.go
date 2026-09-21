// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
//
// This file is a Go translation and modification of MathJax 3.2.2.
// Sources: ts/core/MmlTree/{MmlNode,OperatorDictionary}.ts;
// MmlNodes/{math,mathchoice,mstyle,mfrac,msqrt,mroot,msubsup,munderover,
// mmultiscripts,mtable,mtr,mi,mo,maligngroup,mn,mtext,mspace,ms,mglyph}.ts;
// and ts/input/tex/FilterUtil.ts.

package tex

import (
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/ordered"
)

// inheritedAttribute is MathJax's AttributeList pair: the node kind that
// introduced an inheritable value and the value itself.
type inheritedAttribute struct {
	source string
	value  mml.Property
}

type inheritedAttributes = ordered.Map[inheritedAttribute]

// setMathMLInheritance ports AbstractMmlNode.setInheritedAttributes() and the
// kind-specific overrides used by the complete MML registry.  MathJax runs
// this as the setInherited post-filter after TeX parsing; default layers are
// not semantically useful without the same pass.
func setMathMLInheritance(root *mml.Node, display bool) {
	setInheritedAttributes(root, ordered.New[inheritedAttribute](), display, 0, false)
}

// cleanMathMLAttributes ports FilterUtil.cleanAttributes, which runs directly
// after setInherited and cleanStretchy in MathJax's TeX post-filter chain.
func cleanMathMLAttributes(root *mml.Node) {
	root.Walk(func(n *mml.Node) bool {
		if n.Attributes == nil {
			return true
		}
		keep := map[string]bool{}
		if value, ok := n.Attributes.Get("mjx-keep-attrs"); ok {
			for _, name := range strings.Fields(sourceValueString(value)) {
				keep[name] = true
			}
		}
		n.Attributes.Explicit().Delete("mjx-keep-attrs")
		for _, name := range n.Attributes.ExplicitNames() {
			if keep[name] {
				continue
			}
			explicit, _ := n.Attributes.GetExplicit(name)
			inherited, ok := n.Attributes.GetInherited(name)
			if ok && reflect.DeepEqual(explicit, inherited) {
				n.Attributes.Explicit().Delete(name)
			}
		}
		return true
	})
}

func setInheritedAttributes(n *mml.Node, attributes *inheritedAttributes, display bool, level int, prime bool) {
	if n == nil || n.Kind == "text" || n.Kind == "XML" {
		return
	}
	if n.Kind == "MathChoice" || n.Kind == "mathchoice" {
		// MmlNodes/mathchoice.ts removes this transient node during the
		// inherited-attribute pass.  Display takes the first child; otherwise
		// text, script, and scriptscript select children 1 through 3.
		selection := 0
		if !display {
			selection = level + 1
			if selection < 1 {
				selection = 1
			} else if selection > 3 {
				selection = 3
			}
		}
		var child *mml.Node
		if selection < len(n.Children) {
			child = n.Children[selection]
		}
		if child == nil {
			child = node("mrow")
		}
		if parent := n.Parent; parent != nil {
			_ = parent.ReplaceChild(child, n)
		}
		setInheritedAttributes(child, attributes, display, level, prime)
		return
	}

	// MmlMtable first forces the four indent attributes into its inherited
	// layer. Parser-produced tables do not put these attributes explicitly on
	// the table, so no deletion step is needed here.
	if n.Kind == "mtable" {
		for _, name := range []string{"indentalign", "indentalignfirst", "indentshift", "indentshiftfirst"} {
			if inherited, ok := attributes.Get(name); ok {
				n.Attributes.SetInherited(name, inherited.value)
			}
		}
	}

	defaults := n.Attributes.Defaults()
	attributes.Range(func(name string, inherited inheritedAttribute) bool {
		if !defaults.Has(name) && name != "scriptminsize" && name != "scriptsizemultiplier" {
			return true
		}
		if !blocksInheritance(inherited.source, n.Kind, name) {
			n.Attributes.SetInherited(name, inherited.value)
		}
		return true
	})
	if _, explicit := n.Attributes.GetExplicit("displaystyle"); !explicit {
		n.Attributes.SetInherited("displaystyle", display)
	}
	if _, explicit := n.Attributes.GetExplicit("scriptlevel"); !explicit {
		n.Attributes.SetInherited("scriptlevel", level)
	}
	if prime {
		n.SetProperty("texprimestyle", true)
	}

	setChildInheritedAttributes(n, attributes, display, level, prime)
	refreshDynamicFlags(n)
}

func blocksInheritance(source, target, name string) bool {
	switch source {
	case "mstyle":
		if target == "mpadded" {
			switch name {
			case "width", "height", "depth", "lspace", "voffset":
				return true
			}
		}
		if target == "mtable" {
			switch name {
			case "width", "height", "depth", "align":
				return true
			}
		}
	case "maligngroup":
		if (target == "mrow" || target == "mtable") && name == "groupalign" {
			return true
		}
	}
	return false
}

func addInheritedAttributes(current *inheritedAttributes, source string, explicit *ordered.Map[mml.Property]) *inheritedAttributes {
	updated := current.Clone()
	explicit.Range(func(name string, value mml.Property) bool {
		if name != "displaystyle" && name != "scriptlevel" && name != "style" {
			updated.Set(name, inheritedAttribute{source: source, value: value})
		}
		return true
	})
	return updated
}

func addInheritedValues(current *inheritedAttributes, source string, values ...any) *inheritedAttributes {
	updated := current.Clone()
	for i := 0; i < len(values); i += 2 {
		updated.Set(values[i].(string), inheritedAttribute{source: source, value: values[i+1]})
	}
	return updated
}

func setChildInheritedAttributes(n *mml.Node, attributes *inheritedAttributes, display bool, level int, prime bool) {
	switch n.Kind {
	case "math":
		attributes = addInheritedAttributes(attributes, n.Kind, n.Attributes.Explicit())
		displaystyle, _ := n.Attributes.Get("displaystyle")
		displayValue, _ := n.Attributes.Get("display")
		display = propertyBool(displaystyle) || (!propertyBool(displaystyle) && displayValue == "block")
		n.Attributes.SetInherited("displaystyle", display)
		if scriptlevel, ok := propertyIntValue(n.Attributes, "scriptlevel"); ok {
			level = scriptlevel
		} else {
			level = 0
		}

	case "mstyle":
		if scriptlevel, ok := n.Attributes.GetExplicit("scriptlevel"); ok && scriptlevel != nil {
			if raw, isString := scriptlevel.(string); isString {
				raw = strings.TrimSpace(raw)
				parsed, err := strconv.Atoi(raw)
				if err == nil {
					if strings.HasPrefix(raw, "+") || strings.HasPrefix(raw, "-") {
						level += parsed
					} else {
						level = parsed
					}
				}
			} else if parsed, ok := propertyInt(scriptlevel); ok {
				level = parsed
			}
			prime = false
		}
		if displaystyle, ok := n.Attributes.GetExplicit("displaystyle"); ok && displaystyle != nil {
			display = propertyBool(displaystyle)
			prime = false
		}
		if cramped, ok := n.Attributes.GetExplicit("data-cramped"); ok && cramped != nil {
			prime = propertyBool(cramped)
		}
		attributes = addInheritedAttributes(attributes, n.Kind, n.Attributes.Explicit())
		if len(n.Children) != 0 {
			setInheritedAttributes(n.Children[0], attributes, display, level, prime)
		}
		return

	case "mfrac":
		if !display || level > 0 {
			level++
		}
		if len(n.Children) > 0 {
			setInheritedAttributes(n.Children[0], attributes, false, level, prime)
		}
		if len(n.Children) > 1 {
			setInheritedAttributes(n.Children[1], attributes, false, level, true)
		}
		return

	case "msqrt":
		if len(n.Children) != 0 {
			setInheritedAttributes(n.Children[0], attributes, display, level, true)
		}
		return

	case "mroot":
		if len(n.Children) > 0 {
			setInheritedAttributes(n.Children[0], attributes, display, level, true)
		}
		if len(n.Children) > 1 {
			setInheritedAttributes(n.Children[1], attributes, false, level+2, prime)
		}
		return

	case "msub", "msup", "msubsup":
		if len(n.Children) > 0 {
			setInheritedAttributes(n.Children[0], attributes, display, level, prime)
		}
		if len(n.Children) > 1 {
			subscript := n.Kind != "msup"
			setInheritedAttributes(n.Children[1], attributes, false, level+1, prime || subscript)
		}
		if len(n.Children) > 2 {
			setInheritedAttributes(n.Children[2], attributes, false, level+1, prime)
		}
		return

	case "munder", "mover", "munderover":
		inheritUnderOver(n, attributes, display, level, prime)
		return

	case "mmultiscripts":
		inheritMultiscripts(n, attributes, display, level, prime)
		return

	case "mtable":
		if scriptlevel, ok := n.Property("scriptlevel"); ok {
			if value, ok := propertyInt(scriptlevel); ok && value != 0 {
				level = value
			}
		}
		displaystyle, _ := n.Attributes.GetExplicit("displaystyle")
		if displaystyle == nil {
			displaystyle, _ = n.Attributes.GetDefault("displaystyle")
		}
		display = propertyBool(displaystyle)
		columnalign, _ := n.Attributes.Get("columnalign")
		attributes = addInheritedValues(attributes, n.Kind,
			"columnalign", columnalign,
			"rowalign", "center",
		)
		rowalign, _ := n.Attributes.Get("rowalign")
		rows := strings.Fields(propertyString(rowalign))
		cramped, _ := n.Attributes.GetExplicit("data-cramped")
		rowValue := "center"
		if inherited, ok := attributes.Get("rowalign"); ok {
			rowValue = propertyString(inherited.value)
		}
		for _, child := range n.Children {
			if len(rows) != 0 {
				rowValue, rows = rows[0], rows[1:]
			}
			childAttributes := attributes.Clone()
			childAttributes.Set("rowalign", inheritedAttribute{source: n.Kind, value: rowValue})
			setInheritedAttributes(child, childAttributes, display, level, propertyBool(cramped))
		}
		return

	case "mtr", "mlabeledtr":
		columnalign, _ := n.Attributes.Get("columnalign")
		columns := strings.Fields(propertyString(columnalign))
		if n.Kind == "mlabeledtr" && n.Parent != nil {
			side, _ := n.Parent.Attributes.Get("side")
			columns = append([]string{propertyString(side)}, columns...)
		}
		rowalign, _ := n.Attributes.Get("rowalign")
		attributes = addInheritedValues(attributes, n.Kind,
			"rowalign", rowalign,
			"columnalign", "center",
		)
		columnValue := "center"
		for _, child := range n.Children {
			if len(columns) != 0 {
				columnValue, columns = columns[0], columns[1:]
			}
			childAttributes := attributes.Clone()
			childAttributes.Set("columnalign", inheritedAttribute{source: n.Kind, value: columnValue})
			setInheritedAttributes(child, childAttributes, display, level, prime)
		}
		return

	case "mi":
		// Token nodes have only internal text children.  Plain one-character
		// identifiers acquire italic as an inherited value, not an explicit
		// parser attribute.
		if utf8.RuneCountInString(textContent(n)) == 1 && !attributes.Has("mathvariant") {
			n.Attributes.SetInherited("mathvariant", "italic")
		}
		return

	case "mo":
		applyOperatorInheritance(n)
		return

	case "mn", "mtext", "mspace", "ms", "mglyph":
		return
	}

	for _, child := range n.Children {
		setInheritedAttributes(child, attributes, display, level, prime)
	}
}

// applyOperatorInheritance ports MmlMo.checkOperatorTable.  The form lookup is
// positional (prefix/infix/postfix), with an explicitly supplied form taking
// priority, and then falls back to the Unicode range table.
func applyOperatorInheritance(n *mml.Node) {
	forms := operatorForms(n)
	if explicit, ok := n.Attributes.GetExplicit("form"); ok {
		if form, ok := explicit.(string); ok {
			ordered := []string{form}
			for _, candidate := range forms {
				if candidate != form {
					ordered = append(ordered, candidate)
				}
			}
			forms = ordered
		}
	}
	n.Attributes.SetInherited("form", forms[0])

	text := textContent(n)
	if definition, ok := lookupOperatorDefinition(text, forms); ok {
		n.OperatorLspace = float64(definition.Lspace+1) / 18
		n.OperatorRspace = float64(definition.Rspace+1) / 18
		if _, explicitClass := n.Property("texClass"); !explicitClass {
			n.TeXClass = mml.TeXClass(definition.TexClass)
		}
		for _, property := range definition.Properties {
			n.Attributes.SetInherited(property.Name, property.Value)
			if property.Name == "stretchy" && propertyBool(property.Value) {
				if _, fixStretchy := n.Property("fixStretchy"); fixStretchy {
					n.Attributes.Set("stretchy", false)
				}
			}
		}
		n.RemoveProperty("fixStretchy")
		return
	}

	runes := []rune(text)
	if len(runes) == 0 {
		return
	}
	for _, operatorRange := range mjOperatorRanges {
		codepoint := int(runes[0])
		if codepoint < operatorRange.First {
			break
		}
		if codepoint <= operatorRange.Last {
			for _, spacing := range mjOperatorMMLSpacing {
				if spacing.TexClass == operatorRange.TexClass {
					n.OperatorLspace = float64(spacing.Lspace+1) / 18
					n.OperatorRspace = float64(spacing.Rspace+1) / 18
					break
				}
			}
			if _, explicitClass := n.Property("texClass"); !explicitClass {
				n.TeXClass = mml.TeXClass(operatorRange.TexClass)
			}
			n.RemoveProperty("fixStretchy")
			return
		}
	}
	n.RemoveProperty("fixStretchy")
}

func lookupOperatorDefinition(text string, forms []string) (mjOperatorDefinition, bool) {
	for _, form := range forms {
		var dictionary []mjOperatorDictionaryEntry
		switch form {
		case "prefix":
			dictionary = mjOperatorDictionaryPrefix
		case "postfix":
			dictionary = mjOperatorDictionaryPostfix
		default:
			dictionary = mjOperatorDictionaryInfix
		}
		for _, entry := range dictionary {
			if entry.Operator == text {
				return entry.Definition, true
			}
		}
	}
	return mjOperatorDefinition{}, false
}

func operatorForms(n *mml.Node) []string {
	core := n
	parent := n.Parent
	semanticParent := n.ParentNode()
	for semanticParent != nil && semanticParent.Flags.Embellished {
		core = parent
		parent = semanticParent.Parent
		semanticParent = semanticParent.ParentNode()
	}
	if parent != nil && parent.Kind == "mrow" {
		nonSpace := make([]*mml.Node, 0, len(parent.Children))
		for _, child := range parent.Children {
			if child != nil && !child.Flags.Spacelike {
				nonSpace = append(nonSpace, child)
			}
		}
		if len(nonSpace) != 1 {
			if len(nonSpace) != 0 && nonSpace[0] == core {
				return []string{"prefix", "infix", "postfix"}
			}
			if len(nonSpace) != 0 && nonSpace[len(nonSpace)-1] == core {
				return []string{"postfix", "infix", "prefix"}
			}
		}
	}
	return []string{"infix", "prefix", "postfix"}
}

func inheritUnderOver(n *mml.Node, attributes *inheritedAttributes, display bool, level int, prime bool) {
	if len(n.Children) == 0 {
		return
	}
	over := 2
	under := 1
	if n.Kind == "mover" {
		over, under = 1, 2
	}
	setInheritedAttributes(n.Children[0], attributes, display, level, prime || over < len(n.Children))
	force := !display && effectiveCoreBool(n.Children[0], "movablelimits")
	for i := 1; i < len(n.Children); i++ {
		accent := "accentunder"
		if i == over {
			accent = "accent"
		}
		accentValue, _ := n.Attributes.Get(accent)
		childLevel := level
		if force || !propertyBool(accentValue) {
			childLevel++
		}
		setInheritedAttributes(n.Children[i], attributes, false, childLevel, prime || i == under)
	}
}

func inheritMultiscripts(n *mml.Node, attributes *inheritedAttributes, display bool, level int, prime bool) {
	if len(n.Children) == 0 {
		return
	}
	setInheritedAttributes(n.Children[0], attributes, display, level, prime)
	prescripts := false
	scriptIndex := 0
	for i := 1; i < len(n.Children); i++ {
		child := n.Children[i]
		if child.Kind == "mprescripts" {
			prescripts = true
			continue
		}
		setInheritedAttributes(child, attributes, false, level+1, prime || scriptIndex%2 == 0)
		scriptIndex++
	}
	_ = prescripts // structure is already normalized by the TeX parser
}

func effectiveCoreBool(n *mml.Node, name string) bool {
	core := n
	for core != nil && core.Kind != "mo" && len(core.Children) != 0 {
		core = core.Children[0]
	}
	if core == nil {
		return false
	}
	value, _ := core.Attributes.Get(name)
	return propertyBool(value)
}

func propertyBool(value any) bool {
	switch value := value.(type) {
	case bool:
		return value
	case string:
		parsed, err := strconv.ParseBool(value)
		return err == nil && parsed
	case int:
		return value != 0
	case float64:
		return value != 0
	}
	return false
}

func propertyString(value any) string {
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	return ""
}

func propertyInt(value any) (int, bool) {
	switch value := value.(type) {
	case int:
		return value, true
	case float64:
		return int(value), true
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		return parsed, err == nil
	}
	return 0, false
}

func propertyIntValue(attributes *mml.Attributes, name string) (int, bool) {
	value, ok := attributes.Get(name)
	if !ok {
		return 0, false
	}
	return propertyInt(value)
}
