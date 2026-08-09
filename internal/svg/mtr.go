// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package svg

import (
	"math"

	"github.com/d2lang/mathjax-go/internal/layout"
)

// computeRowBBox retains the normal inferred-row geometry. Table geometry is
// computed from the individual mtd boxes so that labels can be excluded from
// the data columns, matching CommonMtr/CommonMlabeledtr.
func (w *wrapper) computeRowBBox(bbox *layout.BBox) {
	w.computeChildrenBBox(bbox)
}

func (t *tableLayout) rowIndex(row *wrapper) int {
	for i, candidate := range t.rows {
		if candidate == row {
			return i
		}
	}
	return -1
}

func (t *tableLayout) rowBaseline(index int) float64 {
	space := t.rowHalfSpacing()
	y := t.h - t.frameLine
	for i := 0; i < index; i++ {
		h, d := t.rowHD(i)
		y -= space[i] + h + d + space[i+1] + t.rowLines[i]
	}
	h, _ := t.rowHD(index)
	return y - space[index] - h
}

func rowBackground(element *Element) *Element { return tableBackground(element) }

func (w *wrapper) renderTableRow(parent *Element, table *tableLayout, index int, H, D, topSpace, bottomSpace, topLine, bottomLine float64, labels *Element) {
	element := w.standardSVG(parent)
	rowScale := w.outerBBox().RScale
	if rowScale == 0 {
		rowScale = 1
	}
	scale := 1 / rowScale
	columnSpace := table.columnHalfSpacing()
	columnLines := make([]float64, 0, len(table.columnLines)+2)
	columnLines = append(columnLines, table.frameLine)
	columnLines = append(columnLines, table.columnLines...)
	columnLines = append(columnLines, table.frameLine)

	x := columnLines[0] * scale
	for column, cell := range tableRowCells(w) {
		if column >= len(table.computed) {
			break
		}
		cell.cellToSVG(element)
		leftSpace := columnSpace[column] * scale
		rightSpace := columnSpace[column+1] * scale
		width := table.computed[column] * scale
		leftLine := columnLines[column] * scale
		rightLine := columnLines[column+1] * scale
		columnAlign := table.columnAlign(w, cell, column)
		rowAlign := table.cellRowAlign(w, cell, index)
		dx, dy := cell.placeTableCell(x+leftSpace, 0, width, H*scale, D*scale, columnAlign, rowAlign)
		span := leftSpace + width + rightSpace
		cell.resizeCellBackground(-(dx + leftSpace + leftLine/2), -(D*scale + bottomSpace*scale + dy), span+(leftLine+rightLine)/2, (H+D+topSpace+bottomSpace)*scale)
		x += span + rightLine
	}

	if background := rowBackground(element); background != nil {
		background.SetAttr("y", fixed(-(D+bottomSpace+bottomLine/2)*scale))
		background.SetAttr("width", fixed(table.getWidth()*scale))
		background.SetAttr("height", fixed((topLine/2+topSpace+H+D+bottomSpace+bottomLine/2)*scale))
	}

	if label := tableLabelCell(w); label != nil && labels != nil {
		label.cellToSVG(labels)
		align := tableString("side", table.table)
		if explicit := stringAttributeExplicit(label.node, "columnalign"); explicit != "" {
			align = explicit
		}
		rowAlign := table.cellRowAlign(w, label, index)
		dx, dy := label.placeTableCell(0, table.rowBaseline(index), table.data.L, H, D, align, rowAlign)
		label.resizeCellBackground(-dx, -(D + bottomSpace + dy), table.data.L, H+D+topSpace+bottomSpace)
	}
}

// rowToSVG is the standalone dispatch target. In normal rendering mtable owns
// row placement so it can supply the common row dimensions and label group.
func (w *wrapper) rowToSVG(parent *Element) {
	if w.parent == nil || w.parent.node.Kind != "mtable" {
		element := w.standardSVG(parent)
		w.addChildren(element)
		return
	}
	table := w.parent.tableState()
	index := table.rowIndex(w)
	if index < 0 {
		element := w.standardSVG(parent)
		w.addChildren(element)
		return
	}
	h, d := table.rowHD(index)
	space := table.rowHalfSpacing()
	topLine, bottomLine := table.frameLine, table.frameLine
	if index > 0 {
		topLine = table.rowLines[index-1]
	}
	if index < len(table.rowLines) {
		bottomLine = table.rowLines[index]
	}
	w.renderTableRow(parent, table, index, h, d, space[index], space[index+1], topLine, bottomLine, nil)
}

func tableAlignX(width float64, bbox *layout.BBox, align string) float64 {
	switch align {
	case "right":
		return width - (bbox.W+bbox.R)*bbox.RScale
	case "left":
		return bbox.L * bbox.RScale
	default:
		return (width - bbox.W*bbox.RScale) / 2
	}
}

func tableAlignY(H, D, h, d float64, align string) float64 {
	switch align {
	case "top":
		return H - h
	case "bottom":
		return d - D
	case "center":
		return ((H - h) - (D - d)) / 2
	default:
		return 0 // baseline and axis
	}
}

func (w *wrapper) placeTableCell(x, y, width, H, D float64, columnAlign, rowAlign string) (float64, float64) {
	bbox := w.outerBBox()
	h := math.Max(bbox.H*bbox.RScale, .75)
	d := math.Max(bbox.D*bbox.RScale, .25)
	dx := tableAlignX(width, bbox, columnAlign)
	dy := tableAlignY(H, D, h, d, rowAlign)
	w.place(x+dx, y+dy, w.element)
	return dx, dy
}
