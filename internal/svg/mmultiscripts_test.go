// Copyright 2018-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package svg

import (
	"crypto/sha256"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/d2lang/mathjax-go/internal/jscompat"
	"github.com/d2lang/mathjax-go/internal/layout"
	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
)

func wrapperTrancheToken(kind, text string, class mml.TeXClass) *mml.Node {
	node := mml.NewNode(kind, nil, nil, mml.NewText(text))
	node.Flags.Token = true
	node.TeXClass = class
	if kind == "mo" {
		node.Flags.Embellished = true
	}
	return node
}

func wrapperTrancheRoot(child *mml.Node) *mml.Node {
	return mml.NewNode("math", nil, nil, child)
}

func wrapperTrancheWrapper(node *mml.Node) *wrapper {
	options := pipeline.DefaultOptions()
	renderer := &renderer{
		options: options,
		params:  layout.TeXParameters,
		pxPerEm: options.Ex / layout.TeXParameters.XHeight,
	}
	return renderer.wrap(node, nil, 0, false)
}

func computeTrancheWrapper(w *wrapper) {
	switch w.node.Kind {
	case "mmultiscripts":
		w.computeMultiscriptsBBox(w.bbox)
	case "mfenced":
		w.computeFencedBBox(w.bbox)
	case "mglyph":
		w.computeGlyphBBox(w.bbox)
	case "maction":
		w.computeActionBBox(w.bbox)
	case "semantics":
		w.computeSemanticsBBox(w.bbox)
	case "annotation":
		w.computeAnnotationBBox(w.bbox)
	case "annotation-xml":
		w.computeAnnotationXMLBBox(w.bbox)
	case "XML", "xml":
		w.computeXMLBBox(w.bbox)
	}
	w.bboxComputed = true
}

func renderTrancheWrapper(w *wrapper, parent *Element) {
	switch w.node.Kind {
	case "mmultiscripts":
		w.multiscriptsToSVG(parent)
	case "mfenced":
		w.fencedToSVG(parent)
	case "mglyph":
		w.glyphToSVG(parent)
	case "maction":
		w.actionToSVG(parent)
	case "semantics":
		w.semanticsToSVG(parent)
	case "annotation":
		w.annotationToSVG(parent)
	case "annotation-xml":
		w.annotationXMLToSVG(parent)
	case "XML", "xml":
		w.xmlToSVG(parent)
	}
}

// typesetTrancheWrapper is a narrow test harness used before and after the
// central wrapper dispatch is wired.  It is otherwise byte-for-byte the
// single-child path through Typesetter.Typeset.
func typesetTrancheWrapper(t *testing.T, root *mml.Node) string {
	t.Helper()
	options := pipeline.DefaultOptions()
	r := &renderer{
		options: options,
		params:  layout.TeXParameters,
		pxPerEm: options.Ex / layout.TeXParameters.XHeight,
	}
	prepareTeXClasses(root)
	rootWrapper := r.wrap(root, nil, 0, options.Display)
	if len(rootWrapper.children) != 1 {
		t.Fatalf("tranche root has %d children, want one", len(rootWrapper.children))
	}
	target := rootWrapper.children[0]
	computeTrancheWrapper(target)

	bbox := layout.EmptyBBox()
	bbox.Append(target.outerBBox())
	bbox.Clean()
	rootWrapper.bbox = bbox
	rootWrapper.bboxComputed = true
	px := options.Em / 1000
	width := math.Max(bbox.W, px)
	height := math.Max(bbox.H+bbox.D, px)

	g := NewElement("g").
		SetAttr("stroke", "currentColor").
		SetAttr("fill", "currentColor").
		SetAttr("stroke-width", "0").
		SetAttr("transform", "scale(1,-1)")
	mathElement := rootWrapper.standardSVG(g)
	renderTrancheWrapper(target, mathElement)
	target.place(target.outerBBox().L*target.outerBBox().RScale, 0, target.element)

	svg := NewElement("svg", g).
		SetStyle("vertical-align", r.ex(-bbox.D)).
		SetAttr("xmlns", namespace).
		SetAttr("width", r.ex(width)).
		SetAttr("height", r.ex(height)).
		SetAttr("role", "img").
		SetAttr("focusable", "false").
		SetAttr("viewBox", "0 "+strings.Join([]string{
			jscompat.Fixed(-bbox.H*1000, 1),
			jscompat.Fixed(width*1000, 1),
			jscompat.Fixed(height*1000, 1),
		}, " "))
	return svg.String()
}

func requireFrozenWrapperHash(t *testing.T, got, want string) {
	t.Helper()
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(got)))
	if hash != want {
		t.Fatalf("frozen MathJax 3.2.2 SVG SHA-256 = %s, want %s\n%s", hash, want, got)
	}
}

func multiscriptFixture() *mml.Node {
	script := func(node *mml.Node) *mml.Node {
		node.Attributes.SetInherited("scriptlevel", 1)
		node.Attributes.SetInherited("displaystyle", false)
		return node
	}
	node := mml.NewNode("mmultiscripts", nil, nil,
		wrapperTrancheToken("mi", "x", mml.TeXClassOrd),
		script(wrapperTrancheToken("mi", "i", mml.TeXClassOrd)),
		script(wrapperTrancheToken("mn", "2", mml.TeXClassOrd)),
		script(wrapperTrancheToken("mi", "j", mml.TeXClassOrd)),
		script(wrapperTrancheToken("mn", "3", mml.TeXClassOrd)),
		mml.NewNode("mprescripts", nil, nil),
		script(wrapperTrancheToken("mi", "k", mml.TeXClassOrd)),
		script(wrapperTrancheToken("mn", "4", mml.TeXClassOrd)),
		script(mml.NewNode("none", nil, nil)),
		script(wrapperTrancheToken("mn", "5", mml.TeXClassOrd)),
	)
	node.TeXClass = mml.TeXClassOrd
	node.SetProperty("scriptalign", "center right")
	return node
}

func TestMultiscriptsFrozenSourceShapedSVG(t *testing.T) {
	got := typesetTrancheWrapper(t, wrapperTrancheRoot(multiscriptFixture()))
	const want = "a22ca7741abb15d9af0349e1dd9f431badb2de42fab2a5a414bc673315e4e8d1"
	requireFrozenWrapperHash(t, got, want)
}

func TestPrescriptsOnlyFrozenSourceShapedSVG(t *testing.T) {
	script := func(node *mml.Node) *mml.Node {
		node.Attributes.SetInherited("scriptlevel", 1)
		node.Attributes.SetInherited("displaystyle", false)
		return node
	}
	node := mml.NewNode("mmultiscripts", nil, nil,
		wrapperTrancheToken("mi", "X", mml.TeXClassOrd),
		mml.NewNode("mprescripts", nil, nil),
		script(wrapperTrancheToken("mi", "b", mml.TeXClassOrd)),
		script(wrapperTrancheToken("mi", "a", mml.TeXClassOrd)),
	)
	node.TeXClass = mml.TeXClassOrd
	got := typesetTrancheWrapper(t, wrapperTrancheRoot(node))
	const want = "522c3b8e5959877ab8935bfd84db9c6a8d7e37e675315401b705cb247811af54"
	requireFrozenWrapperHash(t, got, want)
}

func TestMultiscriptsMarkersAndAlignment(t *testing.T) {
	w := wrapperTrancheWrapper(multiscriptFixture())
	w.computeMultiscriptsBBox(w.bbox)
	w.bboxComputed = true
	if w.bbox.W <= w.children[0].outerBBox().W {
		t.Fatalf("multiscript width %g does not include pre/post scripts", w.bbox.W)
	}
	root := NewElement("root")
	w.multiscriptsToSVG(root)
	output := root.String()
	if strings.Contains(output, `data-mml-node="mprescripts"`) {
		t.Fatalf("mprescripts marker was rendered:\n%s", output)
	}
	if !strings.Contains(output, `data-mml-node="none"`) {
		t.Fatalf("empty script placeholder was not rendered:\n%s", output)
	}
	if strings.Count(output, `<g transform="translate(`) < 4 {
		t.Fatalf("pre/post sup/sub rows are missing:\n%s", output)
	}
}
