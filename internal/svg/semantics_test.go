// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package svg

import (
	"strings"
	"testing"

	"github.com/d2lang/mathjax-go/internal/layout"
	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
)

func semanticsFixture() *mml.Node {
	annotation := mml.NewNode("annotation", nil, nil, mml.NewText("ignored"))
	annotation.Attributes.Set("encoding", "application/x-tex")
	node := mml.NewNode("semantics", nil, nil,
		wrapperTrancheToken("mi", "x", mml.TeXClassOrd), annotation,
	)
	node.Flags.NotParent = true
	node.TeXClass = mml.TeXClassOrd
	return node
}

func TestSemanticsFrozenSourceShapedSVG(t *testing.T) {
	got := typesetTrancheWrapper(t, wrapperTrancheRoot(semanticsFixture()))
	const want = "ec398ba7562e5550a62d4d5342bb278495a27e2421472be351f29fe64f18ed6a"
	requireFrozenWrapperHash(t, got, want)
}

func TestSemanticsRendersOnlyPresentationChild(t *testing.T) {
	w := wrapperTrancheWrapper(semanticsFixture())
	w.computeSemanticsBBox(w.bbox)
	w.bboxComputed = true
	root := NewElement("root")
	w.semanticsToSVG(root)
	output := root.String()
	if strings.Contains(output, "annotation") || strings.Contains(output, "ignored") {
		t.Fatalf("semantic annotation leaked into presentation SVG:\n%s", output)
	}
	if !strings.Contains(output, `data-mml-node="mi"`) {
		t.Fatalf("presentation child is missing:\n%s", output)
	}
}

func TestAnnotationAndXMLLiteDOMBehavior(t *testing.T) {
	annotation := mml.NewNode("annotation", nil, nil, mml.NewText("note"))
	w := wrapperTrancheWrapper(annotation)
	w.computeAnnotationBBox(w.bbox)
	w.bboxComputed = true
	if w.bbox.W != 0 || w.bbox.H != 0 || w.bbox.D != 0 {
		t.Fatalf("annotation bbox = %+v, want zero", w.bbox)
	}
	root := NewElement("root")
	w.annotationToSVG(root)
	if !strings.Contains(root.String(), `data-mml-node="annotation"><path`) {
		t.Fatalf("direct annotation did not use generic child output:\n%s", root.String())
	}

	options := pipeline.DefaultOptions()
	renderer := &renderer{
		options: options, params: layout.TeXParameters,
		pxPerEm: options.Ex / layout.TeXParameters.XHeight,
	}
	xml := mml.NewNode("XML", nil, nil)
	xmlWrapper := renderer.wrap(xml, nil, 0, false)
	xmlWrapper.computeXMLBBox(xmlWrapper.bbox)
	xmlWrapper.bboxComputed = true
	root = NewElement("root")
	xmlWrapper.xmlToSVG(root)
	const foreignObject = `<foreignObject data-mjx-xml="true" y="0px" width="0px" height="0px" transform="scale(62.5) matrix(1 0 0 -1 0 0)"></foreignObject>`
	if root.String() != "<root>"+foreignObject+"</root>" {
		t.Fatalf("Lite XML container differs:\n%s", root.String())
	}

	annotationXML := mml.NewNode("annotation-xml", nil, nil, xml)
	w = renderer.wrap(annotationXML, nil, 0, false)
	w.computeAnnotationXMLBBox(w.bbox)
	w.bboxComputed = true
	root = NewElement("root")
	w.annotationXMLToSVG(root)
	if !strings.Contains(root.String(), foreignObject) {
		t.Fatalf("annotation-xml did not contain XML foreignObject:\n%s", root.String())
	}
}
