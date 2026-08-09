// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package svg

import (
	"crypto/sha256"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/d2lang/mathjax-go/internal/layout"
	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/tex"
)

func tableTestSpace(width, height, depth string) *mml.Node {
	node := mml.NewNode("mspace", nil, nil)
	node.Attributes.Set("width", width)
	node.Attributes.Set("height", height)
	node.Attributes.Set("depth", depth)
	return node
}

func tableTestCell(width, height, depth string) *mml.Node {
	return mml.NewNode("mtd", nil, nil, tableTestSpace(width, height, depth))
}

func tableTestRow(cells ...*mml.Node) *mml.Node {
	return mml.NewNode("mtr", nil, nil, cells...)
}

func tableTestWrapper(node *mml.Node) *wrapper {
	options := pipeline.DefaultOptions()
	renderer := &renderer{
		options: options,
		params:  layout.TeXParameters,
		pxPerEm: options.Ex / layout.TeXParameters.XHeight,
	}
	return renderer.wrap(node, nil, 0, false)
}

func prepareTableWrapper(w *wrapper) *layout.BBox {
	w.computeTableBBox(w.bbox)
	w.bboxComputed = true
	return w.bbox
}

func closeTableFloat(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-12 {
		t.Errorf("%s = %.15g, want %.15g", name, got, want)
	}
}

func TestTableNaturalGeometry(t *testing.T) {
	table := mml.NewNode("mtable", nil, nil,
		tableTestRow(tableTestCell("1em", ".5em", ".2em"), tableTestCell("2em", ".8em", ".1em")),
		tableTestRow(tableTestCell("3em", ".4em", ".3em"), tableTestCell(".5em", ".2em", ".2em")),
	)
	table.SetProperty("useHeight", false)
	table.Attributes.Set("rowspacing", "0")
	table.Attributes.Set("columnspacing", "0")
	table.Attributes.Set("framespacing", "0 0")
	w := tableTestWrapper(table)
	bbox := prepareTableWrapper(w)
	closeTableFloat(t, "width", bbox.W, 5)
	closeTableFloat(t, "height", bbox.H, 1.1)
	closeTableFloat(t, "depth", bbox.D, .6)

	data := newTableLayout(w).data
	for i, want := range []float64{.8, .4} {
		closeTableFloat(t, "row height", data.H[i], want)
	}
	for i, want := range []float64{.2, .3} {
		closeTableFloat(t, "row depth", data.D[i], want)
	}
	for i, want := range []float64{3, 2} {
		closeTableFloat(t, "column width", data.W[i], want)
	}
}

func TestTableEmptyFrameRetainsSpacingWithoutLine(t *testing.T) {
	table := mml.NewNode("mtable", nil, nil,
		tableTestRow(tableTestCell("1em", ".5em", ".2em")),
	)
	table.SetProperty("useHeight", false)
	table.Attributes.Set("frame", "")
	table.Attributes.Set("framespacing", "1em .5em")
	table.Attributes.Set("rowspacing", "0")
	table.Attributes.Set("columnspacing", "0")
	w := tableTestWrapper(table)
	bbox := prepareTableWrapper(w)
	closeTableFloat(t, "empty-frame width", bbox.W, 3)
	closeTableFloat(t, "empty-frame total height", bbox.H+bbox.D, 1.7)

	root := NewElement("root")
	w.tableToSVG(root)
	if strings.Contains(root.String(), `data-frame="true"`) {
		t.Fatalf("empty frame unexpectedly emitted a frame line:\n%s", root.String())
	}
}

func TestTableFixedAndPercentageColumns(t *testing.T) {
	table := mml.NewNode("mtable", nil, nil,
		tableTestRow(
			tableTestCell("1em", ".5em", ".2em"),
			tableTestCell("2em", ".5em", ".2em"),
			tableTestCell("3em", ".5em", ".2em"),
		),
	)
	table.SetProperty("useHeight", false)
	table.Attributes.Set("width", "10em")
	table.Attributes.Set("columnwidth", "25% fit auto")
	table.Attributes.Set("columnspacing", "0")
	table.Attributes.Set("framespacing", "0 0")
	w := tableTestWrapper(table)
	bbox := prepareTableWrapper(w)
	geometry := newTableLayout(w)
	for i, want := range []float64{2.5, 4.5, 3} {
		closeTableFloat(t, "fixed column", geometry.computed[i], want)
	}
	closeTableFloat(t, "fixed table width", bbox.W, 10)

	table.Attributes.Set("width", "80%")
	table.Attributes.Set("columnwidth", "50% auto auto")
	w = tableTestWrapper(table)
	bbox = prepareTableWrapper(w)
	if bbox.PWidth != "" {
		t.Fatalf("resolved top-level percentage marker = %q, want empty", bbox.PWidth)
	}
	geometry = newTableLayout(w)
	for i, want := range []float64{14.144, 6.572, 7.572} {
		closeTableFloat(t, "resolved percentage column", geometry.computed[i], want)
	}
	closeTableFloat(t, "resolved percentage table width", bbox.W, 28.288)
}

func TestTableEqualRowsColumnsAndAlignmentRow(t *testing.T) {
	table := mml.NewNode("mtable", nil, nil,
		tableTestRow(tableTestCell("1em", ".7em", ".1em"), tableTestCell("2em", ".2em", ".1em")),
		tableTestRow(tableTestCell("3em", ".2em", ".3em"), tableTestCell(".5em", ".2em", ".1em")),
	)
	table.SetProperty("useHeight", false)
	table.Attributes.Set("equalrows", true)
	table.Attributes.Set("equalcolumns", true)
	table.Attributes.Set("align", "baseline 2")
	table.Attributes.Set("rowspacing", "0")
	table.Attributes.Set("columnspacing", "0")
	table.Attributes.Set("framespacing", "0 0")
	w := tableTestWrapper(table)
	bbox := prepareTableWrapper(w)
	geometry := newTableLayout(w)
	closeTableFloat(t, "equal column 0", geometry.computed[0], 3)
	closeTableFloat(t, "equal column 1", geometry.computed[1], 3)
	closeTableFloat(t, "equal-row table height", bbox.H+bbox.D, 1.6)
	// Row 2 receives h=.35, d=.45 inside its equalized .8em allocation.
	closeTableFloat(t, "alignment-row height", bbox.H, 1.15)
	closeTableFloat(t, "alignment-row depth", bbox.D, .45)
}

func TestTableSVGCellLinesFrameAndOrdering(t *testing.T) {
	table := mml.NewNode("mtable", nil, nil,
		tableTestRow(tableTestCell("1em", ".5em", ".2em"), tableTestCell("1em", ".5em", ".2em")),
		tableTestRow(tableTestCell("1em", ".5em", ".2em"), tableTestCell("1em", ".5em", ".2em")),
	)
	table.SetProperty("useHeight", false)
	table.Attributes.Set("width", "4em")
	table.Attributes.Set("columnwidth", "2em 2em")
	table.Attributes.Set("columnalign", "right left")
	table.Attributes.Set("columnspacing", "0")
	table.Attributes.Set("rowspacing", "0")
	table.Attributes.Set("framespacing", "0 0")
	table.Attributes.Set("columnlines", "dashed")
	table.Attributes.Set("rowlines", "dotted")
	table.Attributes.Set("frame", "solid")
	w := tableTestWrapper(table)
	prepareTableWrapper(w)
	root := NewElement("root")
	w.tableToSVG(root)
	output := root.String()
	for _, fragment := range []string{
		`data-mml-node="mtable"`,
		`transform="translate(1070,0)"`, // frame plus right-aligned 1em cell in a 2em column
		`data-line="v" class="mjx-dashed"`,
		`data-line="h" class="mjx-dotted"`,
		`data-frame="true" class="mjx-solid"`,
	} {
		if !strings.Contains(output, fragment) {
			t.Errorf("SVG is missing %q:\n%s", fragment, output)
		}
	}
	row := strings.Index(output, `data-mml-node="mtr"`)
	vertical := strings.Index(output, `data-line="v"`)
	horizontal := strings.Index(output, `data-line="h"`)
	frame := strings.Index(output, `data-frame="true"`)
	if !(row >= 0 && row < vertical && vertical < horizontal && horizontal < frame) {
		t.Fatalf("row/column-line/row-line/frame ordering differs:\n%s", output)
	}
}

func TestLabeledRowGeometryAndPlacement(t *testing.T) {
	label := tableTestCell(".5em", ".5em", ".2em")
	cell := tableTestCell("1em", ".5em", ".2em")
	row := mml.NewNode("mlabeledtr", nil, nil, label, cell)
	table := mml.NewNode("mtable", nil, nil, row)
	table.SetProperty("useHeight", false)
	table.Attributes.Set("side", "left")
	table.Attributes.Set("minlabelspacing", ".8em")
	table.Attributes.Set("columnspacing", "0")
	table.Attributes.Set("rowspacing", "0")
	table.Attributes.Set("framespacing", "0 0")
	w := tableTestWrapper(table)
	bbox := prepareTableWrapper(w)
	closeTableFloat(t, "left label padding", bbox.L, 1.3)
	closeTableFloat(t, "centered opposite label padding", bbox.R, 1.3)
	if bbox.PWidth != layout.FullWidth {
		t.Fatalf("labeled table percentage marker = %q, want %q", bbox.PWidth, layout.FullWidth)
	}
	root := NewElement("root")
	w.tableToSVG(root)
	output := root.String()
	for _, fragment := range []string{
		`data-mml-node="mtable" transform="translate(-1300,0)"`,
		`<svg data-table="true" preserveAspectRatio="xMidYMid"`,
		`<svg data-labels="true" preserveAspectRatio="xMinYMid"`,
		`<g data-labels="true" transform="matrix(1 0 0 -1 0 0)">`,
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("responsive left-label table is missing %q:\n%s", fragment, output)
		}
	}
}

func TestNestedPercentageTableUsesContainingColumnWidth(t *testing.T) {
	inner := mml.NewNode("mtable", nil, nil,
		tableTestRow(tableTestCell("1em", ".5em", ".2em")),
		tableTestRow(tableTestCell("2em", ".5em", ".2em")),
	)
	inner.SetProperty("useHeight", false)
	inner.Attributes.Set("width", "50%")
	inner.Attributes.Set("columnalign", "right")
	inner.Attributes.Set("columnspacing", "0")
	inner.Attributes.Set("rowspacing", "0")
	inner.Attributes.Set("framespacing", "0 0")
	outerCell := mml.NewNode("mtd", nil, nil, inner)
	outer := mml.NewNode("mtable", nil, nil, tableTestRow(outerCell))
	outer.SetProperty("useHeight", false)
	outer.Attributes.Set("width", "100%")
	outer.Attributes.Set("frame", "")
	outer.Attributes.Set("framespacing", "1em 0")
	outer.Attributes.Set("columnspacing", "0")
	outer.Attributes.Set("rowspacing", "0")

	outerWrapper := tableTestWrapper(outer)
	innerWrapper := outerWrapper.children[0].children[0].children[0]
	outerLayout := newTableLayout(outerWrapper)
	closeTableFloat(t, "outer resolved column", outerLayout.computed[0], 33.36)
	innerLayout := resolvedTableLayout(innerWrapper)
	closeTableFloat(t, "inner natural width", innerLayout.width, 2)
	closeTableFloat(t, "inner percentage width", innerLayout.pWidth, 16.68)
	closeTableFloat(t, "inner resolved column", innerLayout.computed[0], 16.68)

	root := NewElement("root")
	innerWrapper.tableToSVG(root)
	output := root.String()
	for _, fragment := range []string{
		`transform="translate(15680,`,
		`transform="translate(14680,`,
		`<g transform="translate(-7340,0)">`,
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("nested percentage SVG is missing %q:\n%s", fragment, output)
		}
	}
}

func TestLabeledTableResponsiveRootFrozenSVG(t *testing.T) {
	root, err := tex.NewCompiler().Compile(`\begin{equation}a=b\tag{T}\end{equation}`, true)
	if err != nil {
		t.Fatal(err)
	}
	got, err := NewTypesetter().Typeset(root, pipeline.DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	const want = "5a220d38903b66e14b3f77846582f1571e43f1ace17ccd5454b4a56d47445dab"
	if hash := fmt.Sprintf("%x", sha256.Sum256([]byte(got))); hash != want {
		t.Fatalf("responsive labeled table SVG SHA-256 = %s, want %s\n%s", hash, want, got)
	}
}

func TestAMSCDLongFrozenSVG(t *testing.T) {
	const source = `\begin{CD} B @>{\text{very long label}}>> C S^{{\mathcal{W}}_\Lambda}\otimes T @>j>> T\\ @VVV V \end{CD}`
	root, err := tex.NewCompiler().Compile(source, true)
	if err != nil {
		t.Fatal(err)
	}
	got, err := NewTypesetter().Typeset(root, pipeline.DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	const want = "d0404e1a275e586c57420e6f3ecc6f7bf6584598fb052f3bc723ed806f13d958"
	if hash := fmt.Sprintf("%x", sha256.Sum256([]byte(got))); hash != want {
		t.Fatalf("long AMS-CD SVG SHA-256 = %s, want %s\n%s", hash, want, got)
	}
}

// This fixture repairs only the parser's currently-known missing AMS
// alignment marker, then drives the source-shaped MathML directly through the
// SVG wrappers. It keeps the table gate actionable while the parser owner
// fixes the same node shape upstream of this package.
func TestAlignedTableFrozenMMLSVG(t *testing.T) {
	root, err := tex.NewCompiler().Compile(`\begin{aligned}a&=b+c\\d&=e\end{aligned}`, true)
	if err != nil {
		t.Fatal(err)
	}
	tables := root.Find("mtable")
	if len(tables) != 1 || len(tables[0].Children) != 2 {
		t.Fatalf("aligned fixture table shape = %#v", tables)
	}
	// AMS's aligned environment overrides the generic matrix cadence and forces
	// display style. Keep those source attributes explicit in this SVG fixture.
	tables[0].Attributes.Set("displaystyle", true)
	tables[0].Attributes.Set("rowspacing", "3pt")
	for _, row := range tables[0].Children {
		if len(row.Children) < 2 || len(row.Children[1].Children) != 1 {
			t.Fatalf("aligned fixture row shape = %#v", row)
		}
		contents := row.Children[1].Children[0]
		if len(contents.Children) == 0 || contents.Children[0].Kind != "mi" || nodeText(contents.Children[0]) != "" {
			empty := mml.NewNode("mi", nil, nil, mml.NewText(""))
			empty.Flags.Token = true
			empty.TeXClass = mml.TeXClassOrd
			contents.SetChildren(append([]*mml.Node{empty}, contents.Children...))
		}
	}

	got, err := NewTypesetter().Typeset(root, pipeline.DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	const want = "beea3e49eedda60558d71b4588c4c512c549123baa428013cefe1effbbd3e514"
	if hash := fmt.Sprintf("%x", sha256.Sum256([]byte(got))); hash != want {
		t.Fatalf("source-shaped aligned SVG SHA-256 = %s, want %s\n%s", hash, want, got)
	}
}
