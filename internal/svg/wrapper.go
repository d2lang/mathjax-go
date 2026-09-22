// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package svg

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/d2lang/mathjax-go/internal/font"
	"github.com/d2lang/mathjax-go/internal/layout"
	"github.com/d2lang/mathjax-go/internal/mml"
)

type wrapper struct {
	renderer *renderer
	node     *mml.Node
	parent   *wrapper
	children []*wrapper

	bbox         *layout.BBox
	bboxComputed bool
	variant      font.Variant
	explicitFont bool
	fontFamily   string
	styles       *wrapperStyles
	scriptLevel  int
	displayStyle bool
	element      *Element
	dx           float64

	stretch      font.Delimiter
	hasStretch   bool
	stretchGlyph rune
	size         int
	sizeSet      bool
	surdHeight   float64
	isMathAccent bool

	// table retains CommonMtable's natural and resolved percentage-width
	// state for the lifetime of this wrapped render tree.  It is nil for every
	// non-mtable wrapper and initialized lazily after the complete wrapper tree
	// exists.
	table *tableState
}

func (r *renderer) wrap(node *mml.Node, parent *wrapper, level int, display bool) *wrapper {
	w := &wrapper{
		renderer: r, node: node, parent: parent,
		bbox: layout.ZeroBBox(), variant: font.Normal,
		scriptLevel: level, displayStyle: display,
	}
	w.inheritLayoutState()
	w.initializeStyles()
	// CommonTextNode inherits all of these properties from its token parent;
	// in particular its getScale() override is a no-op.  Giving a text wrapper
	// its own relative scale divides glyph metrics by an enclosing mathsize.
	if node.Kind != "text" {
		w.getVariant()
		w.getScale()
		w.getSpace()
	}
	for i, child := range node.Children {
		childLevel, childDisplay := w.childLayoutState(i)
		wrapped := r.wrap(child, w, childLevel, childDisplay)
		w.children = append(w.children, wrapped)
		// CommonWrapper propagates a percentage-width marker through wrappers
		// that are transparent to layout and through the root math wrapper.
		if wrapped.bbox.PWidth != "" && (node.Flags.NotParent || node.Kind == "math") {
			w.bbox.PWidth = layout.FullWidth
		}
	}
	if node.Kind == "ms" {
		w.initializeStringQuotes()
	}
	if node.Kind == "mtable" {
		initializeTablePWidth(w)
	}
	// CommonMrow has fixesPWidth=false, but its constructor explicitly marks
	// itself full-width when any child carries a percentage marker.
	if node.Kind == "mrow" {
		for _, child := range w.children {
			if child.bbox.PWidth != "" {
				w.bbox.PWidth = layout.FullWidth
				break
			}
		}
	}
	if node.Kind == "msqrt" || node.Kind == "mroot" {
		w.initializeRoot()
	}
	// CommonMunder, CommonMover, and CommonMunderover stretch their children
	// once in their constructors.  Keep this out of computeBBox: table layout
	// can subsequently resize an embellished core, and a bbox recomputation
	// must not undo that later stretch pass.
	if node.Kind == "munder" || node.Kind == "mover" || node.Kind == "munderover" {
		w.initializeMathAccent()
		w.stretchUnderOverChildren()
	}
	if node.Kind == "math" || node.Kind == "mrow" {
		w.stretchRowChildren()
	}
	return w
}

func (w *wrapper) inheritLayoutState() {
	if value, ok := numberAttribute(w.node, "scriptlevel"); ok {
		w.scriptLevel = int(math.Min(value, 2))
	}
	if value, ok := boolAttribute(w.node, "displaystyle"); ok {
		w.displayStyle = value
	}
}

func (w *wrapper) childLayoutState(index int) (int, bool) {
	level, display := w.scriptLevel, w.displayStyle
	switch w.node.Kind {
	case "msub", "msup", "msubsup":
		if index > 0 {
			level++
			display = false
		}
	case "munder", "mover", "munderover":
		if index > 0 {
			level++
			display = false
		}
	case "mfrac":
		if !display {
			level++
		}
		display = false
	case "mroot":
		if index == 1 {
			level += 2
			display = false
		}
	case "mmultiscripts":
		if index > 0 && index < len(w.node.Children) && w.node.Children[index] != nil && w.node.Children[index].Kind != "mprescripts" {
			level++
			display = false
		}
	case "mtable":
		display = false
	}
	if level > 2 {
		level = 2
	}
	return level, display
}

func (w *wrapper) getVariant() {
	if !w.node.Flags.Token && w.node.Kind != "mi" && w.node.Kind != "mn" && w.node.Kind != "mo" && w.node.Kind != "mtext" && w.node.Kind != "ms" {
		return
	}
	// The frozen Node adaptor uses OutputJax's default merrorFont (serif).
	// CommonMtext turns that into an explicit-font variant rather than using
	// MathJax's path font.  The inferred mrow in our shared tree is transparent
	// to the semantic parent relationship, so look through wrapper ancestors.
	if w.node.Kind == "mtext" && w.hasAncestor("merror") {
		w.explicitFont = true
		w.fontFamily = "serif"
		return
	}
	if w.authoredFontFamily() {
		return
	}
	if name := stringAttribute(w.node, "mathvariant", ""); name != "" && font.ValidVariant(font.Variant(name)) {
		w.variant = font.Variant(name)
	} else if w.node.Kind == "mi" {
		text := nodeText(w.node)
		if utf8.RuneCountInString(text) == 1 {
			w.variant = font.Italic
		}
	}
	if w.node.Kind == "mo" && boolAttributeDefault(w.node, "largeop", false) {
		if w.displayStyle {
			w.variant = font.LargeOp
		} else {
			w.variant = font.SmallOp
		}
	}
	if value, ok := w.node.Property("variantForm"); ok && truthy(value) {
		w.variant = font.TeXVariant
	}
}

// CommonWrapper.getVariant selects explicit-font text for an authored family.
// MathML font attributes override CSS, while an explicit mathvariant overrides
// both. Keep the selected styles on this wrapper, never on the input MathML.
func (w *wrapper) authoredFontFamily() bool {
	// Deprecated mglyph character rendering has its own family handler.
	// This path handles the five text-token kinds accepted by MmlToken.
	switch w.node.Kind {
	case "mi", "mn", "mo", "mtext", "ms":
	default:
		return false
	}
	family := stringAttribute(w.node, "fontfamily", "")
	weight := stringAttribute(w.node, "fontweight", "")
	style := stringAttribute(w.node, "fontstyle", "")
	cssFamily, cssWeight, cssStyle := "", "", ""
	if w.styles != nil {
		cssFamily = w.styles.value("font-family")
		cssWeight = w.styles.value("font-weight")
		cssStyle = w.styles.value("font-style")
	}
	if family == "" && cssFamily == "" {
		return false
	}
	if w.styles == nil {
		w.styles = &wrapperStyles{}
	}
	// The primary wrapper removes these CSS declarations before resolving
	// the variant and appends only the chosen explicit-font styles afterward.
	kept := w.styles.other[:0]
	for _, declaration := range w.styles.other {
		if declaration.name != "font-family" && declaration.name != "font-weight" && declaration.name != "font-style" {
			kept = append(kept, declaration)
		}
	}
	w.styles.other = kept
	// CommonMo selects its large-operator variant before the generic font
	// family resolution; CSS font declarations have already been removed.
	if w.node.Kind == "mo" && boolAttributeDefault(w.node, "largeop", false) {
		return false
	}
	if value, ok := w.node.Attributes.GetExplicit("mathvariant"); ok && truthy(value) {
		return false
	}
	if family == "" {
		family = cssFamily
	}
	if weight == "" {
		weight = cssWeight
	}
	if style == "" {
		style = cssStyle
	}
	if weight != "" && strings.IndexFunc(weight, func(r rune) bool { return r < '0' || r > '9' }) == -1 {
		value, _ := strconv.ParseFloat(weight, 64)
		weight = "normal"
		if value > 600 {
			weight = "bold"
		}
	}
	w.styles.setOther("font-family", family)
	if weight != "" {
		w.styles.setOther("font-weight", weight)
	}
	if style != "" {
		w.styles.setOther("font-style", style)
	}
	w.variant = font.Variant("-explicitFont")
	return true
}

func (w *wrapper) getScale() {
	scale := 1.0
	if w.scriptLevel != 0 {
		scale = math.Pow(1/math.Sqrt2, float64(w.scriptLevel))
		minimum := layout.Length2Em("8px", .8, 1, w.renderer.pxPerEm)
		if scale < minimum {
			scale = minimum
		}
	}
	if mathsize := stringAttribute(w.node, "mathsize", "normal"); mathsize != "1" {
		scale *= layout.Length2Em(mathsize, 1, 1, w.renderer.pxPerEm)
	}
	parentScale := 1.0
	if w.parent != nil {
		parentScale = w.parent.bbox.Scale
	}
	w.bbox.Scale = scale
	w.bbox.RScale = scale / parentScale
	if w.node.Flags.Inferred && w.parent != nil {
		w.bbox.Scale = parentScale
		w.bbox.RScale = 1
	}
}

func (w *wrapper) getSpace() {
	if w.node.Flags.Inferred || w.node.Kind == "text" {
		return
	}
	core := coreMONode(w.node)
	if core != nil && (core.Attributes.IsSet("lspace") || core.Attributes.IsSet("rspace")) {
		// CommonWrapper selects MathML spacing before consulting prevClass.
		// Only the outermost embellished wrapper owns the operator's space.
		if w.node.Flags.Embellished && (w.node.Parent == nil || !w.node.Parent.Flags.Embellished) {
			w.getMathMLSpacing(core)
		}
		return
	}
	if w.node.PrevClass == mml.TeXClassNone {
		return
	}
	level := w.scriptLevel > 0 || w.node.PrevLevel > 0
	if space := mml.TeXSpacing(w.node.PrevClass, effectiveTeXClass(w.node), level); space != "" {
		w.bbox.L = layout.Length2Em(space, 1, w.bbox.Scale, w.renderer.pxPerEm)
	}
}

// getMathMLSpacing ports CommonWrapper.getMathMLSpacing. The actual core
// parent matters: an inferred multi-child row counts, but an isolated operator
// or one directly inside a fixed-arity construct receives no outer space.
func (w *wrapper) getMathMLSpacing(core *mml.Node) {
	child := core
	for parent := child.Parent; parent != nil && parent.Kind != "math" &&
		parent.Flags.Embellished && coreMONode(parent) == core; parent = child.Parent {
		child = parent
	}
	parent := child.Parent
	if parent == nil || parent.Kind != "mrow" || len(parent.Children) <= 1 {
		return
	}
	isScript := w.scriptLevel > 0
	if level, ok := numberAttribute(core, "scriptlevel"); ok {
		isScript = level > 0
	}
	space := func(name string, fallback float64) float64 {
		if core.Attributes.IsSet(name) {
			value, _ := core.Attributes.Get(name)
			return math.Max(0, layout.Length2Em(value, 1, w.bbox.Scale, w.renderer.pxPerEm))
		}
		if isScript {
			if fallback < 2.0/18 {
				return 0
			}
			return 2.0 / 18
		}
		return fallback
	}
	w.bbox.L = space("lspace", core.OperatorLspace)
	w.bbox.R = space("rspace", core.OperatorRspace)
	index := parent.ChildIndex(child)
	if index <= 0 || !parent.Children[index-1].Flags.Embellished {
		return
	}
	// Siblings are wrapped in source order; their complete bbox is available
	// even while this wrapper's constructor is still running.
	if w.parent != nil && w.parent.node == parent && index <= len(w.parent.children) {
		previous := w.parent.children[index-1]
		w.bbox.L = math.Max(0, w.bbox.L-previous.getBBox().R)
	}
}

func coreMONode(node *mml.Node) *mml.Node {
	for node != nil && node.Kind != "mo" && node.Flags.Embellished {
		index := node.Flags.CoreIndex
		if index < 0 || index >= len(node.Children) {
			return nil
		}
		node = node.Children[index]
	}
	if node != nil && node.Kind == "mo" {
		return node
	}
	return nil
}

func (w *wrapper) getBBox() *layout.BBox {
	if w.bboxComputed {
		return w.bbox
	}
	w.computeBBox(w.bbox)
	w.bboxComputed = true
	return w.bbox
}

func (w *wrapper) outerBBox() *layout.BBox { return w.styledOuterBBox() }

func (w *wrapper) computeBBox(bbox *layout.BBox) {
	switch w.node.Kind {
	case "text":
		w.computeTextBBox(bbox)
	case "mi":
		w.computeChildrenBBox(bbox)
		w.copySkewIC(bbox)
	case "mo":
		w.computeMoBBox(bbox)
	case "mfrac":
		w.computeFractionBBox(bbox)
	case "msub", "msup", "msubsup":
		w.computeScriptsBBox(bbox)
	case "munder", "mover", "munderover":
		w.computeUnderOverBBox(bbox)
	case "msqrt", "mroot":
		w.computeRootBBox(bbox)
	case "mmultiscripts":
		w.computeMultiscriptsBBox(bbox)
	case "mfenced":
		w.computeFencedBBox(bbox)
	case "mglyph":
		w.computeGlyphBBox(bbox)
	case "maction":
		w.computeActionBBox(bbox)
	case "semantics":
		w.computeSemanticsBBox(bbox)
	case "annotation":
		w.computeAnnotationBBox(bbox)
	case "annotation-xml":
		w.computeAnnotationXMLBBox(bbox)
	case "XML", "xml":
		w.computeXMLBBox(bbox)
	case "menclose":
		w.computeEncloseBBox(bbox)
	case "mtable":
		w.computeTableBBox(bbox)
	case "mtr", "mlabeledtr":
		w.computeRowBBox(bbox)
	case "mtd":
		w.computeCellBBox(bbox)
	case "mpadded":
		w.computePaddedBBox(bbox)
	case "TeXAtom":
		w.computeTeXAtomBBox(bbox)
	case "mspace":
		bbox.W = layout.Length2Em(attribute(w.node, "width", 0), 0, bbox.Scale, w.renderer.pxPerEm)
		bbox.H = layout.Length2Em(attribute(w.node, "height", 0), 0, bbox.Scale, w.renderer.pxPerEm)
		bbox.D = layout.Length2Em(attribute(w.node, "depth", 0), 0, bbox.Scale, w.renderer.pxPerEm)
	default:
		w.computeChildrenBBox(bbox)
	}
}

func (w *wrapper) computeChildrenBBox(bbox *layout.BBox) {
	bbox.Empty()
	for _, child := range w.children {
		bbox.Append(child.outerBBox())
	}
	bbox.Clean()
}

func (w *wrapper) computeTextBBox(bbox *layout.BBox) {
	bbox.Empty()
	if w.parent != nil && (w.parent.explicitFont || w.parent.variant == font.Variant("-explicitFont")) {
		bbox.H = .75
		bbox.D = .2
		for _, codepoint := range w.node.Text {
			if isCJK(codepoint) {
				bbox.W++
			} else {
				bbox.W += .6
			}
		}
		bbox.Clean()
		return
	}
	variant := font.Normal
	if w.parent != nil {
		variant = w.parent.variant
	}
	text := w.node.Text
	if w.parent != nil && w.parent.stretchGlyph != 0 {
		text = string(w.parent.stretchGlyph)
	}
	if w.parent != nil && w.parent.node.Kind == "mn" && strings.HasPrefix(text, "-") {
		text = "−" + strings.TrimPrefix(text, "-")
	}
	text = remapAccentText(w.parent, text)
	for _, codepoint := range text {
		glyph, ok := font.Lookup(variant, codepoint)
		if !ok {
			if isCJK(codepoint) {
				bbox.W++
			} else {
				bbox.W += .6
			}
			bbox.H = math.Max(bbox.H, .75)
			bbox.D = math.Max(bbox.D, .2)
			continue
		}
		metrics := glyph.Metrics
		bbox.W += metrics.Width
		bbox.H = math.Max(bbox.H, metrics.Height)
		bbox.D = math.Max(bbox.D, metrics.Depth)
		if metrics.HasItalicCorrection {
			bbox.IC = metrics.ItalicCorrection
		} else {
			bbox.IC = 0
		}
		if metrics.HasSkew {
			bbox.Skew = metrics.Skew
		} else {
			bbox.Skew = 0
		}
	}
	if utf8.RuneCountInString(text) > 1 {
		bbox.Skew = 0
	}
	bbox.Clean()
}

func (w *wrapper) copySkewIC(bbox *layout.BBox) {
	if len(w.children) == 0 {
		return
	}
	first := w.children[0].getBBox()
	bbox.Skew = first.Skew
	bbox.DX = first.DX
	last := w.children[len(w.children)-1].getBBox()
	if last.IC != 0 {
		bbox.IC = last.IC
		bbox.W += last.IC
	}
}

func (w *wrapper) toSVG(parent *Element) {
	if w.node.Kind == "text" {
		w.textToSVG(parent)
		return
	}
	if w.node.Flags.Inferred {
		// SVGinferredMrow uses its parent as its element.  This matters to
		// callers that place a wrapper after rendering it and also avoids a nil
		// element for parser-produced inferred rows.
		w.element = parent
		w.addChildren(parent)
		return
	}
	if w.node.Kind == "mfrac" {
		w.fractionToSVG(parent)
		return
	}
	if w.node.Kind == "mo" {
		w.moToSVG(parent)
		return
	}
	if w.node.Kind == "msub" || w.node.Kind == "msup" || w.node.Kind == "msubsup" {
		w.scriptsToSVG(parent)
		return
	}
	if w.node.Kind == "munder" || w.node.Kind == "mover" || w.node.Kind == "munderover" {
		w.underOverToSVG(parent)
		return
	}
	if w.node.Kind == "msqrt" || w.node.Kind == "mroot" {
		w.rootToSVG(parent)
		return
	}
	if w.node.Kind == "mmultiscripts" {
		w.multiscriptsToSVG(parent)
		return
	}
	if w.node.Kind == "mfenced" {
		w.fencedToSVG(parent)
		return
	}
	if w.node.Kind == "mglyph" {
		w.glyphToSVG(parent)
		return
	}
	if w.node.Kind == "maction" {
		w.actionToSVG(parent)
		return
	}
	if w.node.Kind == "semantics" {
		w.semanticsToSVG(parent)
		return
	}
	if w.node.Kind == "annotation" {
		w.annotationToSVG(parent)
		return
	}
	if w.node.Kind == "annotation-xml" {
		w.annotationXMLToSVG(parent)
		return
	}
	if w.node.Kind == "XML" || w.node.Kind == "xml" {
		w.xmlToSVG(parent)
		return
	}
	if w.node.Kind == "menclose" {
		w.encloseToSVG(parent)
		return
	}
	if w.node.Kind == "mtable" {
		w.tableToSVG(parent)
		return
	}
	if w.node.Kind == "mtr" || w.node.Kind == "mlabeledtr" {
		w.rowToSVG(parent)
		return
	}
	if w.node.Kind == "mtd" {
		w.cellToSVG(parent)
		return
	}
	if w.node.Kind == "mpadded" {
		w.paddedToSVG(parent)
		return
	}
	if w.node.Kind == "mphantom" {
		w.phantomToSVG(parent)
		return
	}
	if w.node.Kind == "TeXAtom" {
		w.texAtomToSVG(parent)
		return
	}
	if w.node.Kind == "merror" {
		w.errorToSVG(parent)
		return
	}
	element := w.standardSVG(parent)
	w.addChildren(element)
}

func (w *wrapper) standardSVG(parent *Element) *Element {
	element := NewElement("g").SetAttr("data-mml-node", w.node.Kind)
	w.element = element
	parent = w.linkParent(parent, element)
	w.handleStyles(element)
	if w.bbox.RScale != 1 {
		element.SetAttr("transform", "scale("+fixed(w.bbox.RScale/1000, 3)+")")
	}
	if w.explicitFont {
		element.SetStyle("font-family", w.fontFamily).StylesAfterAttributes()
	}
	w.handleBorder(element)
	w.handleColor(element)
	w.handleAttributes(element)
	parent.Append(element)
	return element
}

func (w *wrapper) handleAttributes(element *Element) {
	if w.node == nil || w.node.Attributes == nil {
		return
	}
	for _, name := range w.node.Attributes.ExplicitNames() {
		// MathJax's handleAttributes copies non-MathML data and ARIA
		// attributes after styles, scale, borders, and colors.  Standard
		// MathML attributes are consumed by their wrappers instead.
		if name != "id" && !strings.HasPrefix(name, "data-") && !strings.HasPrefix(name, "aria-") {
			continue
		}
		if hasElementAttribute(element, name) {
			continue
		}
		value, _ := w.node.Attributes.GetExplicit(name)
		element.SetAttr(name, fmt.Sprint(value))
	}
	if value, ok := w.node.Attributes.Get("class"); ok {
		for _, name := range strings.Fields(fmt.Sprint(value)) {
			addElementClass(element, name)
		}
	}
}

func addElementClass(element *Element, name string) {
	for i := range element.Attributes {
		if element.Attributes[i].Name == "class" {
			element.Attributes[i].Value += " " + name
			return
		}
	}
	element.SetAttr("class", " "+name)
}

func hasElementAttribute(element *Element, name string) bool {
	for _, attribute := range element.Attributes {
		if attribute.Name == name {
			return true
		}
	}
	return false
}

func (w *wrapper) addChildren(parent *Element) {
	x := 0.0
	for _, child := range w.children {
		child.toSVG(parent)
		bbox := child.outerBBox()
		if child.element != nil {
			child.place(x+bbox.L*bbox.RScale, 0, child.element)
		}
		x += (bbox.L + bbox.W + bbox.R) * bbox.RScale
	}
}

func (w *wrapper) textToSVG(parent *Element) {
	if w.node.Text == "" {
		return
	}
	if w.parent != nil && (w.parent.explicitFont || w.parent.variant == font.Variant("-explicitFont")) {
		w.element = w.unknownText(w.node.Text, font.Variant("-explicitFont"))
		parent.Append(w.element)
		return
	}
	variant := font.Normal
	if w.parent != nil {
		variant = w.parent.variant
	}
	text := w.node.Text
	if w.parent != nil && w.parent.stretchGlyph != 0 {
		text = string(w.parent.stretchGlyph)
	}
	if w.parent != nil && w.parent.node.Kind == "mn" && strings.HasPrefix(text, "-") {
		text = "−" + strings.TrimPrefix(text, "-")
	}
	text = remapAccentText(w.parent, text)
	// SVGTextNode gives multiple text wrappers positionable groups. CommonMs
	// now has separate opening/body/closing wrappers; keep this correction
	// scoped to those string tokens rather than alter unrelated text output.
	if w.parent != nil && w.parent.node.Kind == "ms" && len(w.parent.children) > 1 {
		w.element = NewElement("g").SetAttr("data-mml-node", "text")
		parent.Append(w.element)
		parent = w.element
	}
	x := 0.0
	for _, codepoint := range text {
		x += w.placeChar(codepoint, x, 0, parent, variant)
	}
}

func (w *wrapper) hasAncestor(kind string) bool {
	for parent := w.parent; parent != nil; parent = parent.parent {
		if parent.node.Kind == kind {
			return true
		}
	}
	return false
}

// isCJK matches NodeMixin.cjkPattern from MathJax 3.2.2.  LiteAdaptor uses
// this to assign deterministic widths when browser text measurement is not
// available.
func isCJK(codepoint rune) bool {
	switch {
	case codepoint >= 0x1100 && codepoint <= 0x115F:
		return true
	case codepoint == 0x2329 || codepoint == 0x232A:
		return true
	case codepoint >= 0x2E80 && codepoint <= 0x303E:
		return true
	case codepoint >= 0x3040 && codepoint <= 0x3247:
		return true
	case codepoint >= 0x3250 && codepoint <= 0x4DBF:
		return true
	case codepoint >= 0x4E00 && codepoint <= 0xA4C6:
		return true
	case codepoint >= 0xA960 && codepoint <= 0xA97C:
		return true
	case codepoint >= 0xAC00 && codepoint <= 0xD7A3:
		return true
	case codepoint >= 0xF900 && codepoint <= 0xFAFF:
		return true
	case codepoint >= 0xFE10 && codepoint <= 0xFE19:
		return true
	case codepoint >= 0xFE30 && codepoint <= 0xFE6B:
		return true
	case codepoint >= 0xFF01 && codepoint <= 0xFF60:
		return true
	case codepoint >= 0xFFE0 && codepoint <= 0xFFE6:
		return true
	case codepoint >= 0x1B000 && codepoint <= 0x1B001:
		return true
	case codepoint >= 0x1F200 && codepoint <= 0x1F251:
		return true
	case codepoint >= 0x20000 && codepoint <= 0x3FFFD:
		return true
	default:
		return false
	}
}

func (w *wrapper) placeChar(codepoint rune, x, y float64, parent *Element, variant font.Variant) float64 {
	codepoint = font.RemapCodepoint(variant, codepoint)
	glyph, ok := font.Lookup(variant, codepoint)
	if !ok {
		text := w.unknownText(string(codepoint), variant)
		parent.Append(text)
		w.place(x, y, text)
		if isCJK(codepoint) {
			return 1
		}
		return .6
	}
	if glyph.Content != "" {
		group := NewElement("g").SetAttr("data-c", strings.ToUpper(strconv.FormatInt(int64(glyph.Codepoint), 16)))
		parent.Append(group)
		w.place(x, y, group)
		offset := 0.0
		for _, content := range glyph.Content {
			offset += w.placeChar(content, offset, y, group, variant)
		}
		return glyph.Metrics.Width
	}
	path := ""
	if glyph.Path != "" {
		path = "M" + glyph.Path + "Z"
	}
	element := NewElement("path").
		SetAttr("data-c", strings.ToUpper(strconv.FormatInt(int64(glyph.Codepoint), 16))).
		SetAttr("d", path)
	parent.Append(element)
	w.place(x, y, element)
	return glyph.Metrics.Width
}

func (w *wrapper) unknownText(text string, variant font.Variant) *Element {
	scale := w.renderer.params.XHeight / w.renderer.options.Ex * w.renderer.options.Em * 1000
	element := NewElement("text", Text(text)).
		SetAttr("data-variant", string(variant)).
		SetAttr("transform", "scale(1,-1)").
		SetAttr("font-size", jsFixed(scale, 1)+"px")
	if variant == font.Variant("-explicitFont") {
		return element
	}
	// A mathematical-alphanumeric character already encodes its font style.
	// SVG.unknownText leaves its CSS font attributes unset in this case.
	if r, size := utf8.DecodeRuneInString(text); size == len(text) && r >= 0x1D400 && r <= 0x1D7FF {
		return element
	}
	family, italic, bold := "serif", false, false
	switch variant {
	case font.Script, font.BoldScript:
		family = "cursive"
	case font.SansSerif, font.BoldSansSerif, font.SansSerifItalic, font.SansSerifBoldItalic:
		family = "sans-serif"
	case font.Monospace:
		family = "monospace"
	}
	switch variant {
	case font.Italic, font.BoldItalic, font.SansSerifItalic, font.SansSerifBoldItalic:
		italic = true
	}
	switch variant {
	case font.Bold, font.BoldItalic, font.DoubleStruck, font.BoldFraktur, font.BoldScript, font.BoldSansSerif, font.SansSerifBoldItalic:
		bold = true
	}
	element.SetAttr("font-family", family)
	if italic {
		element.SetAttr("font-style", "italic")
	}
	if bold {
		element.SetAttr("font-weight", "bold")
	}
	return element
}

func jsFixed(value float64, digits int) string {
	if math.Abs(value) < .0006 {
		return "0"
	}
	result := strconv.FormatFloat(value, 'f', digits, 64)
	result = strings.TrimRight(strings.TrimRight(result, "0"), ".")
	return result
}

func (w *wrapper) place(x, y float64, element *Element) {
	if element == nil {
		return
	}
	x += w.dx
	if fixed(x) == "0" && fixed(y) == "0" {
		return
	}
	translate := fmt.Sprintf("translate(%s,%s)", fixed(x), fixed(y))
	for i := range element.Attributes {
		if element.Attributes[i].Name == "transform" {
			element.Attributes[i].Value = translate + " " + element.Attributes[i].Value
			return
		}
	}
	element.SetAttr("transform", translate)
}

func (w *wrapper) handleColor(element *Element) {
	color := stringAttributeExplicit(w.node, "mathcolor")
	if color == "" {
		color = stringAttributeExplicit(w.node, "color")
	}
	if color != "" {
		element.SetAttr("fill", color).SetAttr("stroke", color)
	}
	background := stringAttributeExplicit(w.node, "mathbackground")
	if background == "" {
		background = stringAttributeExplicit(w.node, "background")
	}
	if background == "" && w.styles != nil {
		background = w.styles.value("background-color")
	}
	if background != "" && background != "transparent" {
		bbox := w.outerBBox()
		rect := NewElement("rect").
			SetAttr("fill", background).
			SetAttr("x", fixed(-w.dx)).
			SetAttr("y", fixed(-bbox.D)).
			SetAttr("width", fixed(bbox.W)).
			SetAttr("height", fixed(bbox.H+bbox.D)).
			SetAttr("data-bgcolor", "true")
		element.Prepend(rect)
	}
}

func nodeText(node *mml.Node) string {
	if node == nil {
		return ""
	}
	if node.Kind == "text" {
		return node.Text
	}
	var builder strings.Builder
	for _, child := range node.Children {
		builder.WriteString(nodeText(child))
	}
	return builder.String()
}

func attribute(node *mml.Node, name string, fallback any) any {
	if node != nil && node.Attributes != nil {
		if value, ok := node.Attributes.Get(name); ok {
			return value
		}
	}
	return fallback
}

func stringAttribute(node *mml.Node, name, fallback string) string {
	value := attribute(node, name, fallback)
	if text, ok := value.(string); ok {
		return text
	}
	return fmt.Sprint(value)
}

func stringAttributeExplicit(node *mml.Node, name string) string {
	if node == nil || node.Attributes == nil {
		return ""
	}
	value, ok := node.Attributes.GetExplicit(name)
	if !ok || value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func boolAttribute(node *mml.Node, name string) (bool, bool) {
	if node == nil || node.Attributes == nil {
		return false, false
	}
	value, ok := node.Attributes.Get(name)
	if !ok {
		return false, false
	}
	return truthy(value), true
}

func boolAttributeDefault(node *mml.Node, name string, fallback bool) bool {
	value, ok := boolAttribute(node, name)
	if !ok {
		return fallback
	}
	return value
}

func truthy(value any) bool {
	switch x := value.(type) {
	case bool:
		return x
	case string:
		return x != "" && x != "false" && x != "0"
	case int:
		return x != 0
	case int64:
		return x != 0
	case float64:
		return x != 0 && !math.IsNaN(x)
	case nil:
		return false
	default:
		return true
	}
}

func numberAttribute(node *mml.Node, name string) (float64, bool) {
	if node == nil || node.Attributes == nil {
		return 0, false
	}
	value, ok := node.Attributes.Get(name)
	if !ok {
		return 0, false
	}
	switch x := value.(type) {
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case float64:
		return x, true
	case string:
		parsed, err := strconv.ParseFloat(x, 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}
