// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package svg

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/d2lang/mathjax-go/internal/font"
	"github.com/d2lang/mathjax-go/internal/layout"
)

// tableDimensions is the source-shaped subset of CommonMtable.TableData used
// by the SVG wrapper: row heights and depths, natural (unlabelled) row sizes,
// column widths, and the label-column width.
type tableDimensions struct {
	H, D   []float64
	W      []float64
	NH, ND []float64
	L      float64
}

type tableColumnMeasure struct {
	fixed bool
	value float64
}

// tableLayout is deliberately ephemeral. MathJax caches this information on
// its wrapper subclasses, but keeping it here avoids adding table-only state to
// the shared wrapper while producing the same deterministic geometry.
type tableLayout struct {
	table *wrapper
	rows  []*wrapper
	// container is the first non-mrow, non-NotParent ancestor.  containerI is
	// the position of the transparent ancestor that contains this table.  This
	// is CommonMtable.findContainer(), retained here because percentage widths
	// and labeled subtables are relative to that wrapper rather than the
	// table's immediate (often inferred-mrow) parent.
	container  *wrapper
	containerI int

	numRows, numCols int
	hasLabels        bool
	isTop            bool

	frame      bool
	frameLine  float64
	frameSpace [2]float64

	columnSpace []float64
	rowSpace    []float64
	columnLines []float64
	rowLines    []float64
	columnStyle []string
	rowStyle    []string

	columns  []tableColumnMeasure
	computed []float64
	data     tableDimensions

	width, height float64
	h, d          float64
	left, right   float64
	pWidth        float64
	resolved      bool
}

// tableState is named separately at the wrapper boundary to make the
// persistent per-render lifetime explicit while retaining tableLayout in the
// source-shaped geometry helpers and tests.
type tableState = tableLayout

var tableDefaults = map[string]string{
	"align":           "axis",
	"rowalign":        "baseline",
	"columnalign":     "center",
	"columnwidth":     "auto",
	"width":           "auto",
	"rowspacing":      "1ex",
	"columnspacing":   ".8em",
	"rowlines":        "none",
	"columnlines":     "none",
	"frame":           "none",
	"framespacing":    "0.4em 0.5ex",
	"side":            "right",
	"minlabelspacing": "0.8em",
}

func tableString(nodeName string, w *wrapper) string {
	return stringAttribute(w.node, nodeName, tableDefaults[nodeName])
}

func tableBool(w *wrapper, name string, fallback bool) bool {
	return boolAttributeDefault(w.node, name, fallback)
}

func tableFields(w *wrapper, name string) []string {
	value := strings.TrimSpace(tableString(name, w))
	if value == "" {
		return nil
	}
	return strings.Fields(value)
}

func repeatTableField(values []string, n int, fallback string) []string {
	if n <= 0 {
		return []string{}
	}
	if len(values) == 0 {
		values = []string{fallback}
	}
	result := make([]string, n)
	for i := range result {
		result[i] = values[min(i, len(values)-1)]
	}
	return result
}

func tableLength(w *wrapper, value string, size ...float64) float64 {
	reference := 1.0
	if len(size) != 0 {
		reference = size[0]
	}
	return layout.Length2Em(value, reference, w.bbox.Scale, w.renderer.pxPerEm)
}

func tableLengths(w *wrapper, values []string) []float64 {
	result := make([]float64, len(values))
	for i, value := range values {
		result[i] = tableLength(w, value)
	}
	return result
}

func tableSum(values ...[]float64) float64 {
	total := 0.0
	for _, list := range values {
		for _, value := range list {
			total += value
		}
	}
	return total
}

func tableMax(values []float64) float64 {
	maximum := 0.0
	for _, value := range values {
		if value > maximum {
			maximum = value
		}
	}
	return maximum
}

func tableIsPercent(value string) bool {
	value = strings.TrimSpace(value)
	if !strings.HasSuffix(value, "%") {
		return false
	}
	_, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(value, "%")), 64)
	return err == nil
}

func tableRows(w *wrapper) []*wrapper {
	return append([]*wrapper(nil), w.children...)
}

func tableRowCells(row *wrapper) []*wrapper {
	if row == nil {
		return nil
	}
	start := 0
	if row.node.Kind == "mlabeledtr" && len(row.children) != 0 {
		start = 1
	}
	return row.children[start:]
}

func tableLabelCell(row *wrapper) *wrapper {
	if row != nil && row.node.Kind == "mlabeledtr" && len(row.children) != 0 {
		return row.children[0]
	}
	return nil
}

func initializeTablePWidth(w *wrapper) {
	if w == nil || w.node == nil || w.node.Kind != "mtable" {
		return
	}
	for _, row := range tableRows(w) {
		if tableLabelCell(row) != nil {
			w.bbox.PWidth = layout.FullWidth
			return
		}
	}
	if width := tableString("width", w); tableIsPercent(width) {
		w.bbox.PWidth = width
	}
}

func (w *wrapper) tableState() *tableState {
	if w.table == nil {
		w.table = newTableLayout(w)
	}
	return w.table
}

func newTableLayout(w *wrapper) *tableLayout {
	container, containerI := findTableContainer(w)
	t := &tableLayout{
		table: w, rows: tableRows(w),
		container: container, containerI: containerI,
		isTop: container == nil || (container.node.Kind == "math" && container.parent == nil),
	}
	t.numRows = len(t.rows)
	for _, row := range t.rows {
		if n := len(tableRowCells(row)); n > t.numCols {
			t.numCols = n
		}
		if tableLabelCell(row) != nil {
			t.hasLabels = true
		}
	}

	frame := tableString("frame", w)
	t.frame = frame != "none"
	// An explicitly empty frame is the AMS multline trick for retaining
	// framespacing without drawing (or reserving thickness for) a frame line.
	if t.frame && frame != "" {
		t.frameLine = .07
	}
	frameSpacing := repeatTableField(tableFields(w, "framespacing"), 2, "0")
	if t.frame {
		converted := tableLengths(w, frameSpacing)
		copy(t.frameSpace[:], converted)
	}

	t.columnSpace = tableLengths(w, repeatTableField(tableFields(w, "columnspacing"), max(0, t.numCols-1), ".8em"))
	t.rowSpace = tableLengths(w, repeatTableField(tableFields(w, "rowspacing"), max(0, t.numRows-1), "1ex"))
	t.columnStyle = repeatTableField(tableFields(w, "columnlines"), max(0, t.numCols-1), "none")
	t.rowStyle = repeatTableField(tableFields(w, "rowlines"), max(0, t.numRows-1), "none")
	t.columnLines = tableLineWidths(t.columnStyle)
	t.rowLines = tableLineWidths(t.rowStyle)

	// CommonMtable computes the column specifications from the natural cell
	// boxes, then stretches rows and columns, and only then measures the final
	// table.  In particular, AMS-CD relies on stretchColumn() to initialize the
	// horizontal arrows hidden under mover/munderover wrappers.
	t.data = t.measureCells()
	t.columns = t.columnMeasures()
	t.stretchRows()
	t.stretchColumns()
	// CommonMtable caches getTableData() before stretching and deliberately
	// keeps those natural row and column measurements afterward.  The stretched
	// child boxes drive their own SVG, but must not retroactively shrink the
	// table's cached column widths (observable for long AMS-CD arrows).
	t.computed = t.computedWidths()
	t.measureTable()
	t.resolveTopPercentageWidth()
	return t
}

// findTableContainer ports CommonMtable.findContainer. Inferred and explicit
// mrows, plus wrappers marked NotParent, are transparent for determining the
// table's width container.
func findTableContainer(w *wrapper) (*wrapper, int) {
	node, parent := w, w.parent
	for parent != nil && (parent.node.Kind == "mrow" || parent.node.Flags.NotParent) {
		node = parent
		parent = parent.parent
	}
	if parent == nil {
		return nil, 0
	}
	for i, child := range parent.children {
		if child == node {
			return parent, i
		}
	}
	return parent, 0
}

// resolveTopPercentageWidth ports the top-level setChildPWidths pass made by
// SVGmath.  html.convert() supplies a default containerWidth of 80ex; after
// division by the output jax's pxPerEm that is 80*x_height in MathJax layout
// ems.  The initial natural-width stretch pass above intentionally happens
// before percentage columns are finalized, matching CommonMtable's lifecycle.
func (t *tableLayout) resolveTopPercentageWidth() {
	width := tableString("width", t.table)
	if !t.isTop || !tableIsPercent(width) {
		return
	}
	t.resolvePercentageWidth(80 * t.table.renderer.params.XHeight)
}

// resolvePercentageWidth is CommonMtable.setChildPWidths().  A nested table's
// natural bbox stays unchanged; pWidth records the resolved width used for
// columns and painting.  Top tables replace their bbox width with pWidth.
func (t *tableLayout) resolvePercentageWidth(containerWidth float64) {
	width := tableString("width", t.table)
	if !tableIsPercent(width) || t.resolved {
		return
	}
	naturalWidth := t.width
	reference := math.Max(containerWidth, t.left+t.width+t.right)
	target := math.Max(t.width, tableLength(t.table, width, reference))
	if tableBool(t.table, "data-width-includes-label", false) {
		target -= t.left + t.right
	}

	var values []string
	if tableBool(t.table, "equalcolumns", false) {
		values = make([]string, t.numCols)
		value := layout.Percent(1 / float64(max(1, t.numCols)))
		for i := range values {
			values[i] = value
		}
	} else {
		values = repeatTableField(tableFields(t.table, "columnwidth"), t.numCols, "auto")
	}
	t.columns = t.fixedColumnMeasures(values, target)
	t.computed = t.computedWidths()
	t.measureTable()
	t.pWidth = t.width
	if !t.isTop {
		t.width = naturalWidth
	}
	t.resolved = true
	if !t.hasLabels {
		t.table.bbox.PWidth = ""
		if t.container != nil {
			t.container.bbox.PWidth = ""
		}
	}
}

// resolvedTableLayout supplies the late percentage-width pass for nested
// tables.  By SVG emission time every container has its natural bbox, so the
// source wrapper's getWrapWidth() contract can be reproduced without a
// browser measurement or shared mutable state.
func resolvedTableLayout(w *wrapper) *tableLayout {
	t := w.tableState()
	if !t.isTop && tableIsPercent(tableString("width", w)) {
		t.resolvePercentageWidth(t.containerWrapWidth(make(map[*wrapper]bool)))
	}
	return t
}

func tableCellChild(cell *wrapper) *wrapper {
	if cell == nil || len(cell.children) == 0 {
		return nil
	}
	return cell.children[0]
}

// tableCoreMO is the wrapper equivalent of MmlNode.coreMO().  CommonWrapper's
// canStretch() delegates through embellished wrappers, while the actual
// getStretchedVariant() call is made on this core mo.
func tableCoreMO(child *wrapper) *wrapper {
	for child != nil && child.node.Flags.Embellished && child.node.Kind != "mo" {
		index := child.node.Flags.CoreIndex
		if index < 0 || index >= len(child.children) {
			return nil
		}
		child = child.children[index]
	}
	if child == nil || child.node.Kind != "mo" {
		return nil
	}
	return child
}

func tableStretchCore(child *wrapper, direction font.Direction) *wrapper {
	// CommonMtable asks the cell's outer wrapper whether it stretches.  That
	// delegates through embellished movers/rows and deliberately permits a
	// second getStretchedVariant call: AMS-CD first sizes an arrow to its long
	// label, then the table column pass clamps it back to its 2.75em minsize.
	if child == nil || !child.canStretch(direction) {
		return nil
	}
	return tableCoreMO(child)
}

func invalidateStretchParents(core *wrapper) {
	for parent := core.parent; parent != nil; parent = parent.parent {
		parent.bboxComputed = false
	}
}

func (t *tableLayout) stretchRows() {
	equal := tableBool(t.table, "equalrows", false)
	equalHeight := 0.0
	if equal {
		equalHeight = t.equalRowHeight()
	}
	for i, row := range t.rows {
		var dimensions []float64
		if equal {
			h, d := t.data.H[i], t.data.D[i]
			dimensions = []float64{(equalHeight + h - d) / 2, (equalHeight - h + d) / 2}
		}
		t.stretchRow(row, dimensions)
	}
}

func (t *tableLayout) stretchRow(row *wrapper, dimensions []float64) {
	cells := tableRowCells(row)
	stretchable := make(map[*wrapper]*wrapper)
	for _, cell := range cells {
		child := tableCellChild(cell)
		if core := tableStretchCore(child, font.DirectionVertical); core != nil {
			stretchable[child] = core
		}
	}
	if len(stretchable) == 0 || len(row.children) <= 1 {
		return
	}
	if dimensions == nil {
		height, depth := 0.0, 0.0
		all := len(stretchable) > 1 && len(stretchable) == len(row.children)
		for _, cell := range cells {
			child := tableCellChild(cell)
			if child == nil {
				continue
			}
			_, stretches := stretchable[child]
			if !all && stretches {
				continue
			}
			bbox := child.outerBBox()
			height = math.Max(height, bbox.H*bbox.RScale)
			depth = math.Max(depth, bbox.D*bbox.RScale)
		}
		dimensions = []float64{height, depth}
	}
	for _, core := range stretchable {
		core.getStretchedVariant(dimensions, false)
		invalidateStretchParents(core)
	}
}

func (t *tableLayout) stretchColumns() {
	for column := 0; column < t.numCols; column++ {
		var width *float64
		if column < len(t.columns) && t.columns[column].fixed {
			value := t.columns[column].value
			width = &value
		}
		t.stretchColumn(column, width)
	}
}

func (t *tableLayout) stretchColumn(column int, width *float64) {
	stretchable := make(map[*wrapper]*wrapper)
	for _, row := range t.rows {
		cells := tableRowCells(row)
		if column >= len(cells) {
			continue
		}
		child := tableCellChild(cells[column])
		if core := tableStretchCore(child, font.DirectionHorizontal); core != nil {
			stretchable[child] = core
		}
	}
	if len(stretchable) == 0 || len(t.rows) <= 1 {
		return
	}
	target := 0.0
	if width != nil {
		target = *width
	} else {
		all := len(stretchable) > 1 && len(stretchable) == len(t.rows)
		for _, row := range t.rows {
			cells := tableRowCells(row)
			if column >= len(cells) {
				continue
			}
			child := tableCellChild(cells[column])
			if child == nil {
				continue
			}
			_, stretches := stretchable[child]
			if !all && stretches {
				continue
			}
			bbox := child.outerBBox()
			target = math.Max(target, bbox.W*bbox.RScale)
		}
	}
	for _, core := range stretchable {
		core.getStretchedVariant([]float64{target}, false)
		invalidateStretchParents(core)
	}
}

func tableLineWidths(styles []string) []float64 {
	widths := make([]float64, len(styles))
	for i, style := range styles {
		if style != "none" {
			widths[i] = .07
		}
	}
	return widths
}

func (t *tableLayout) rowAlign(row, index int) string {
	aligns := repeatTableField(tableFields(t.table, "rowalign"), t.numRows, "baseline")
	align := aligns[index]
	if row >= 0 && row < len(t.rows) {
		if explicit := stringAttributeExplicit(t.rows[row].node, "rowalign"); explicit != "" {
			align = explicit
		}
	}
	return align
}

func (t *tableLayout) columnAlign(row *wrapper, cell *wrapper, column int) string {
	aligns := repeatTableField(tableFields(t.table, "columnalign"), t.numCols, "center")
	align := "center"
	if column >= 0 && column < len(aligns) {
		align = aligns[column]
	}
	if row != nil {
		rowAligns := strings.Fields(stringAttributeExplicit(row.node, "columnalign"))
		if len(rowAligns) != 0 {
			align = rowAligns[min(column, len(rowAligns)-1)]
		}
	}
	if cell != nil {
		if explicit := stringAttributeExplicit(cell.node, "columnalign"); explicit != "" {
			align = explicit
		}
	}
	return align
}

func (t *tableLayout) cellRowAlign(row *wrapper, cell *wrapper, rowIndex int) string {
	align := t.rowAlign(rowIndex, rowIndex)
	if cell != nil {
		if explicit := stringAttributeExplicit(cell.node, "rowalign"); explicit != "" {
			align = explicit
		}
	}
	return align
}

func (t *tableLayout) useHeight() bool {
	if value, ok := t.table.node.Property("useHeight"); ok {
		return truthy(value)
	}
	return true
}

func (t *tableLayout) updateHDW(cell *wrapper, column, row int, align string, H, D, W []float64, tallest float64) float64 {
	if cell == nil {
		return tallest
	}
	bbox := cell.outerBBox()
	scale := 1.0
	if cell.parent != nil {
		scale = cell.parent.bbox.RScale
	}
	h, d, width := bbox.H*scale, bbox.D*scale, bbox.W*scale
	if t.useHeight() {
		h = math.Max(h, .75)
		d = math.Max(d, .25)
	}
	if explicit := stringAttributeExplicit(cell.node, "rowalign"); explicit != "" {
		align = explicit
	}
	m := 0.0
	if align != "baseline" && align != "axis" {
		m = h + d
		h, d = 0, 0
	}
	H[row] = math.Max(H[row], h)
	D[row] = math.Max(D[row], d)
	tallest = math.Max(tallest, m)
	if column >= 0 && column < len(W) {
		W[column] = math.Max(W[column], width)
	}
	return tallest
}

func extendTableHD(row int, H, D []float64, tallest float64) {
	delta := (tallest - (H[row] + D[row])) / 2
	if delta < .00001 {
		return
	}
	H[row] += delta
	D[row] += delta
}

func (t *tableLayout) measureCells() tableDimensions {
	data := tableDimensions{
		H: make([]float64, t.numRows), D: make([]float64, t.numRows),
		W: make([]float64, t.numCols), NH: make([]float64, t.numRows), ND: make([]float64, t.numRows),
	}
	labelWidths := []float64{0}
	for j, row := range t.rows {
		tallest := 0.0
		align := t.rowAlign(j, j)
		for i, cell := range tableRowCells(row) {
			tallest = t.updateHDW(cell, i, j, align, data.H, data.D, data.W, tallest)
		}
		data.NH[j], data.ND[j] = data.H[j], data.D[j]
		if label := tableLabelCell(row); label != nil {
			tallest = t.updateHDW(label, 0, j, align, data.H, data.D, labelWidths, tallest)
		}
		extendTableHD(j, data.H, data.D, tallest)
		extendTableHD(j, data.NH, data.ND, tallest)
	}
	data.L = labelWidths[0]
	return data
}

func (t *tableLayout) columnMeasures() []tableColumnMeasure {
	width := tableString("width", t.table)
	if tableBool(t.table, "equalcolumns", false) {
		return t.equalColumnMeasures(width)
	}
	values := repeatTableField(tableFields(t.table, "columnwidth"), t.numCols, "auto")
	switch {
	case width == "auto":
		return t.autoColumnMeasures(values)
	case tableIsPercent(width):
		return t.percentColumnMeasures(values)
	default:
		return t.fixedColumnMeasures(values, tableLength(t.table, width))
	}
}

func (t *tableLayout) equalColumnMeasures(width string) []tableColumnMeasure {
	n := max(1, t.numCols)
	measure := tableColumnMeasure{}
	switch {
	case width == "auto":
		measure = tableColumnMeasure{fixed: true, value: tableMax(t.data.W)}
	case tableIsPercent(width):
		// MathJax leaves percentage columns unresolved until the parent-width pass.
		measure = tableColumnMeasure{}
	default:
		decoration := tableSum(t.columnLines, t.columnSpace) + 2*t.frameSpace[0]
		measure = tableColumnMeasure{fixed: true, value: math.Max(0, tableLength(t.table, width)-decoration) / float64(n)}
	}
	result := make([]tableColumnMeasure, t.numCols)
	for i := range result {
		result[i] = measure
	}
	return result
}

func (t *tableLayout) autoColumnMeasures(values []string) []tableColumnMeasure {
	result := make([]tableColumnMeasure, len(values))
	for i, value := range values {
		if value == "auto" || value == "fit" || tableIsPercent(value) {
			continue
		}
		result[i] = tableColumnMeasure{fixed: true, value: tableLength(t.table, value)}
	}
	return result
}

func (t *tableLayout) percentColumnMeasures(values []string) []tableColumnMeasure {
	hasFit := false
	for _, value := range values {
		hasFit = hasFit || value == "fit"
	}
	result := make([]tableColumnMeasure, len(values))
	for i, value := range values {
		switch {
		case value == "fit", tableIsPercent(value):
		case value == "auto" && !hasFit:
		case value == "auto":
			result[i] = tableColumnMeasure{fixed: true, value: t.data.W[i]}
		default:
			result[i] = tableColumnMeasure{fixed: true, value: tableLength(t.table, value)}
		}
	}
	return result
}

func (t *tableLayout) fixedColumnMeasures(values []string, width float64) []tableColumnMeasure {
	fit, auto := []int{}, []int{}
	for i, value := range values {
		if value == "fit" {
			fit = append(fit, i)
		} else if value == "auto" {
			auto = append(auto, i)
		}
	}
	n := len(fit)
	if n == 0 {
		n = len(auto)
	}
	cwidth := width - tableSum(t.columnLines, t.columnSpace) - 2*t.frameSpace[0]
	remaining := cwidth
	for i, value := range values {
		if value == "fit" || value == "auto" {
			remaining -= t.data.W[i]
		} else {
			remaining -= tableLength(t.table, value, cwidth)
		}
	}
	extra := 0.0
	if n != 0 && remaining > 0 {
		extra = remaining / float64(n)
	}
	result := make([]tableColumnMeasure, len(values))
	for i, value := range values {
		column := tableColumnMeasure{fixed: true}
		switch value {
		case "fit":
			column.value = t.data.W[i] + extra
		case "auto":
			column.value = t.data.W[i]
			if len(fit) == 0 {
				column.value += extra
			}
		default:
			column.value = tableLength(t.table, value, cwidth)
		}
		result[i] = column
	}
	return result
}

func (t *tableLayout) computedWidths() []float64 {
	widths := append([]float64(nil), t.data.W...)
	for i, column := range t.columns {
		if column.fixed {
			widths[i] = column.value
		}
	}
	if tableBool(t.table, "equalcolumns", false) {
		maximum := tableMax(widths)
		for i := range widths {
			widths[i] = maximum
		}
	}
	return widths
}

func (t *tableLayout) equalRowHeight() float64 {
	maximum := 0.0
	for i := range t.data.H {
		maximum = math.Max(maximum, t.data.H[i]+t.data.D[i])
	}
	return maximum
}

func (t *tableLayout) rowHD(index int) (float64, float64) {
	h, d := t.data.H[index], t.data.D[index]
	if tableBool(t.table, "equalrows", false) {
		hd := t.equalRowHeight()
		return (hd + h - d) / 2, (hd - h + d) / 2
	}
	return h, d
}

func (t *tableLayout) rowHalfSpacing() []float64 {
	result := make([]float64, 0, len(t.rowSpace)+2)
	result = append(result, t.frameSpace[1])
	for _, value := range t.rowSpace {
		result = append(result, value/2)
	}
	return append(result, t.frameSpace[1])
}

func (t *tableLayout) columnHalfSpacing() []float64 {
	result := make([]float64, 0, len(t.columnSpace)+2)
	result = append(result, t.frameSpace[0])
	for _, value := range t.columnSpace {
		result = append(result, value/2)
	}
	return append(result, t.frameSpace[0])
}

func (t *tableLayout) alignmentRow() (string, int, bool) {
	fields := strings.Fields(tableString("align", t.table))
	align := "axis"
	if len(fields) != 0 {
		align = fields[0]
	}
	if len(fields) < 2 {
		return align, 0, false
	}
	row, err := strconv.Atoi(fields[1])
	if err != nil {
		return align, 0, false
	}
	if row < 0 {
		row += t.numRows + 1
	}
	if row < 1 || row > t.numRows {
		return align, 0, false
	}
	return align, row - 1, true
}

func (t *tableLayout) verticalPosition(row int, align string) float64 {
	space := t.rowHalfSpacing()
	y := t.frameLine
	for j := 0; j < row; j++ {
		h, d := t.rowHD(j)
		y += space[j] + h + d + space[j+1] + t.rowLines[j]
	}
	h, d := t.rowHD(row)
	switch align {
	case "center":
		y += space[row] + (h+d)/2
	case "bottom":
		y += space[row] + h + d + space[row+1]
	case "baseline":
		y += space[row] + h
	case "axis":
		y += space[row] + h - t.table.renderer.params.Axis
	}
	return y
}

func (t *tableLayout) bboxHD(height float64) (float64, float64) {
	align, row, hasRow := t.alignmentRow()
	if hasRow {
		h := t.verticalPosition(row, align)
		return h, height - h
	}
	half := height / 2
	switch align {
	case "top":
		return 0, height
	case "bottom":
		return height, 0
	case "axis":
		return half + t.table.renderer.params.Axis, half - t.table.renderer.params.Axis
	default:
		return half, half
	}
}

func (t *tableLayout) containerAlign() string {
	if t.isTop {
		// D2's pinned SVG output jax uses displayAlign="center" and zero
		// displayIndent.
		return "center"
	}
	if t.container == nil {
		return "left"
	}
	switch t.container.node.Kind {
	case "mtd":
		return stringAttribute(t.container.node, "columnalign", "center")
	case "mfrac":
		if boolAttributeDefault(t.container.node, "bevelled", false) {
			return "left"
		}
		if t.containerI == 0 {
			return stringAttribute(t.container.node, "numalign", "center")
		}
		return stringAttribute(t.container.node, "denomalign", "center")
	case "mpadded":
		return stringAttribute(t.container.node, "data-align", "left")
	default:
		return "left"
	}
}

func (t *tableLayout) containerWrapWidth(seen map[*wrapper]bool) float64 {
	if t.isTop || t.container == nil {
		return 80 * t.table.renderer.params.XHeight
	}
	container := t.container
	if seen[container] {
		return container.getBBox().W
	}
	seen[container] = true
	defer delete(seen, container)

	switch container.node.Kind {
	case "math":
		return container.getBBox().W
	case "mtd":
		row := container.parent
		if row == nil || row.parent == nil || row.parent.node.Kind != "mtable" {
			return container.getBBox().W
		}
		column := -1
		for i, cell := range tableRowCells(row) {
			if cell == container {
				column = i
				break
			}
		}
		if column < 0 {
			return container.getBBox().W
		}
		outer := row.parent.tableState()
		if !outer.isTop && tableIsPercent(tableString("width", row.parent)) {
			outer.resolvePercentageWidth(outer.containerWrapWidth(seen))
		}
		if column < len(outer.computed) {
			return outer.computed[column]
		}
		if column < len(outer.data.W) {
			return outer.data.W[column]
		}
	case "mfrac":
		if boolAttributeDefault(container.node, "bevelled", false) && t.containerI < len(container.children) {
			return container.children[t.containerI].outerBBox().W
		}
		width := container.getBBox().W - 2*container.fractionPad()
		if container.fractionThickness() != 0 {
			width -= .2
		}
		return width
	case "mpadded":
		return container.getBBox().W
	}
	if t.containerI >= 0 && t.containerI < len(container.children) {
		return container.children[t.containerI].getBBox().W
	}
	return container.getBBox().W
}

func (t *tableLayout) getPadAlignShift(side string) (float64, string, float64) {
	pad := t.data.L + tableLength(t.table, tableString("minlabelspacing", t.table))
	align := t.containerAlign()
	shift := 0.0
	if align == side {
		if side == "left" {
			shift = math.Max(pad, shift) - pad
		} else {
			shift = math.Min(-pad, shift) + pad
		}
	}
	return pad, align, shift
}

func (t *tableLayout) labelLR() (float64, float64) {
	if !t.hasLabels {
		return 0, 0
	}
	side := tableString("side", t.table)
	pad, align, _ := t.getPadAlignShift(side)
	labelsInWidth := boolAttributeDefault(t.table.node, "data-width-includes-label", false)
	if labelsInWidth && t.frame && t.frameSpace[0] != 0 {
		pad -= t.frameSpace[0]
	}
	if align == "center" && !labelsInWidth {
		return pad, pad
	}
	if side == "left" {
		return pad, 0
	}
	return 0, pad
}

func (t *tableLayout) measureTable() {
	if tableBool(t.table, "equalrows", false) {
		t.height = tableSum(t.rowLines, t.rowSpace) + t.equalRowHeight()*float64(t.numRows)
	} else {
		t.height = tableSum(t.data.H, t.data.D, t.rowLines, t.rowSpace)
	}
	t.height += 2 * (t.frameLine + t.frameSpace[1])
	t.width = tableSum(t.computed, t.columnLines, t.columnSpace) + 2*(t.frameLine+t.frameSpace[0])
	width := tableString("width", t.table)
	if width != "auto" && !tableIsPercent(width) {
		t.width = math.Max(t.width, tableLength(t.table, width)+2*t.frameLine)
	}
	t.h, t.d = t.bboxHD(t.height)
	t.left, t.right = t.labelLR()
}

// computeTableBBox ports CommonMtable.computeBBox. Top-level percentage widths
// have already been resolved against html.convert()'s 80ex container; nested
// percentage widths remain marked for a future parent-width pass.
func (w *wrapper) computeTableBBox(bbox *layout.BBox) {
	t := w.tableState()
	bbox.Empty()
	bbox.W, bbox.H, bbox.D = t.width, t.h, t.d
	bbox.L, bbox.R = t.left, t.right
	bbox.PWidth = ""
	if t.hasLabels {
		bbox.PWidth = layout.FullWidth
	} else if width := tableString("width", w); tableIsPercent(width) && !t.resolved {
		bbox.PWidth = width
	}
	bbox.Clean()
}

func tableBackground(element *Element) *Element {
	rect, ok := elementFirstChild(element).(*Element)
	if !ok || rect.Tag != "rect" {
		return nil
	}
	for _, attr := range rect.Attributes {
		if attr.Name == "data-bgcolor" && attr.Value == "true" {
			return rect
		}
	}
	return nil
}

func setTableLineThickness(element *Element, thickness float64, style string) *Element {
	if thickness != .07 {
		element.SetAttr("stroke-thickness", fixed(thickness))
		if style != "solid" {
			dashes := fixed(2 * thickness)
			if style == "dotted" {
				dashes = "0," + dashes
			}
			element.SetAttr("stroke-dasharray", dashes)
		}
	}
	return element
}

func (t *tableLayout) makeFrame() *Element {
	style, thickness := tableString("frame", t.table), t.frameLine
	return setTableLineThickness(NewElement("rect").
		SetAttr("data-frame", "true").
		SetAttr("class", "mjx-"+style).
		SetAttr("width", fixed(t.width-thickness)).
		SetAttr("height", fixed(t.h+t.d-thickness)).
		SetAttr("x", fixed(thickness/2)).
		SetAttr("y", fixed(thickness/2-t.d)), thickness, style)
}

func (t *tableLayout) makeVLine(x float64, style string, thickness float64) *Element {
	dot := 0.0
	if style == "dotted" {
		dot = thickness / 2
	}
	x = x + thickness/2
	return setTableLineThickness(NewElement("line").
		SetAttr("data-line", "v").
		SetAttr("class", "mjx-"+style).
		SetAttr("x1", fixed(x)).
		SetAttr("y1", fixed(dot-t.d)).
		SetAttr("x2", fixed(x)).
		SetAttr("y2", fixed(t.h-dot)), thickness, style)
}

func (t *tableLayout) makeHLine(y float64, style string, thickness float64) *Element {
	dot := 0.0
	if style == "dotted" {
		dot = thickness / 2
	}
	y -= thickness / 2
	return setTableLineThickness(NewElement("line").
		SetAttr("data-line", "h").
		SetAttr("class", "mjx-"+style).
		SetAttr("x1", fixed(dot)).
		SetAttr("y1", fixed(y)).
		SetAttr("x2", fixed(t.width-dot)).
		SetAttr("y2", fixed(y)), thickness, style)
}

func (t *tableLayout) addColumnLines(element *Element) {
	space := t.columnHalfSpacing()
	x := t.frameLine
	for i, style := range t.columnStyle {
		x += space[i] + t.computed[i] + space[i+1]
		if style != "none" {
			element.Append(t.makeVLine(x, style, t.columnLines[i]))
		}
		x += t.columnLines[i]
	}
}

func (t *tableLayout) addRowLines(element *Element) {
	space := t.rowHalfSpacing()
	y := t.h - t.frameLine
	for i, style := range t.rowStyle {
		h, d := t.rowHD(i)
		y -= space[i] + h + d + space[i+1]
		if style != "none" {
			element.Append(t.makeHLine(y, style, t.rowLines[i]))
		}
		y -= t.rowLines[i]
	}
}

func (t *tableLayout) getWidth() float64 {
	if t.pWidth != 0 {
		return t.pWidth
	}
	return t.width
}

// handlePWidth ports SVGmtable.handlePWidth.  The resolved numeric width is
// intentionally separate from the nested table's natural bbox width.
func (t *tableLayout) handlePWidth(element *Element) float64 {
	if t.pWidth == 0 {
		return 0
	}
	bbox := t.table.getBBox()
	W := bbox.L + t.pWidth + bbox.R
	containerWidth := 0.0
	if t.isTop {
		containerWidth = W
	}
	containerWidth = math.Max(containerWidth, t.containerWrapWidth(make(map[*wrapper]bool))) - bbox.L - bbox.R
	dw := bbox.W - math.Min(t.pWidth, containerWidth)
	dx := dw / 2
	switch t.containerAlign() {
	case "left":
		dx = 0
	case "right":
		dx = dw
	}
	if fixed(dx) != "0" {
		children := append([]Node(nil), element.Children...)
		element.Children = nil
		group := NewElement("g", children...)
		t.table.place(dx, 0, group)
		element.Append(group)
	}
	return dx
}

func (t *tableLayout) topTable(element, labels *Element, side string) {
	bbox := t.table.getBBox()
	W := bbox.L + t.getWidth() + bbox.R
	_, align, shift := t.getPadAlignShift(side)
	dx := shift + bbox.L
	if align == "right" {
		dx -= W
	} else if align == "center" {
		dx -= W / 2
	}

	const matrix = "matrix(1 0 0 -1 0 0)"
	scale := jsFixed((t.table.renderer.params.XHeight*1000)/t.table.renderer.options.Ex, 2)
	transform := fmt.Sprintf("translate(0 %s) %s scale(%s)", fixed(bbox.H), matrix, scale)

	children := append([]Node(nil), element.Children...)
	element.Children = nil
	tableContents := NewElement("g", children...).SetAttr("transform", matrix)
	tableSVG := NewElement("svg", tableContents).
		SetAttr("data-table", "true").
		SetAttr("preserveAspectRatio", map[string]string{
			"left": "xMinYMid", "right": "xMaxYMid",
		}[align]).
		SetAttr("viewBox", strings.Join([]string{
			fixed(-dx), fixed(-bbox.H), "1", fixed(bbox.H + bbox.D),
		}, " "))
	if align != "left" && align != "right" {
		tableSVG.SetAttr("preserveAspectRatio", "xMidYMid")
	}
	labelX := "0"
	labelAspect := "xMinYMid"
	if side != "left" {
		labelX = fixed(t.data.L)
		labelAspect = "xMaxYMid"
	}
	labelsSVG := NewElement("svg", labels).
		SetAttr("data-labels", "true").
		SetAttr("preserveAspectRatio", labelAspect).
		SetAttr("viewBox", strings.Join([]string{
			labelX, fixed(-bbox.H), "1", fixed(bbox.H + bbox.D),
		}, " "))
	element.Append(NewElement("g", tableSVG, labelsSVG).SetAttr("transform", transform))
	// The parent adds bbox.L when appending this wrapper.  Remove it here so
	// the responsive table and labels share the source wrapper's origin.
	t.table.place(-bbox.L, 0, element)
}

func (t *tableLayout) subTable(element, labels *Element, side string, dx float64) {
	bbox := t.table.getBBox()
	W := bbox.L + t.getWidth() + bbox.R
	containerWidth := math.Max(W, t.containerWrapWidth(make(map[*wrapper]bool)))
	x := 0.0
	align := t.containerAlign()
	if side == "left" {
		switch align {
		case "right":
			x = W - containerWidth + dx
		case "center":
			x = (W-containerWidth)/2 + dx
		}
		x -= bbox.L
	} else {
		switch align {
		case "left":
			x = containerWidth
		case "right":
			x = W + dx
		default:
			x = (containerWidth+W)/2 + dx
		}
		x -= bbox.L + t.data.L
	}
	t.table.place(x, 0, labels)
	element.Append(labels)
}

func (t *tableLayout) handleLabels(element, labels *Element, dx float64) {
	if !t.hasLabels || labels == nil || len(labels.Children) == 0 {
		return
	}
	side := tableString("side", t.table)
	if t.isTop {
		t.topTable(element, labels, side)
	} else {
		t.subTable(element, labels, side, dx)
	}
}

// tableToSVG ports SVGmtable.toSVG, including responsive top-level labels and
// explicit nested-label placement.
func (w *wrapper) tableToSVG(parent *Element) {
	element := w.standardSVG(parent)
	t := resolvedTableLayout(w)
	if background := tableBackground(element); background != nil {
		background.SetAttr("width", fixed(t.getWidth()))
	}
	labels := NewElement("g").SetAttr("data-labels", "true")
	if t.isTop {
		labels.SetAttr("transform", "matrix(1 0 0 -1 0 0)")
	}
	space := t.rowHalfSpacing()
	y := t.h - t.frameLine
	for i, row := range t.rows {
		h, d := t.rowHD(i)
		topLine, bottomLine := t.frameLine, t.frameLine
		if i > 0 {
			topLine = t.rowLines[i-1]
		}
		if i < len(t.rowLines) {
			bottomLine = t.rowLines[i]
		}
		row.renderTableRow(element, t, i, h, d, space[i], space[i+1], topLine, bottomLine, labels)
		baseline := y - space[i] - h
		row.place(0, baseline)
		y -= space[i] + h + d + space[i+1] + bottomLine
	}
	t.addColumnLines(element)
	t.addRowLines(element)
	if t.frame && t.frameLine != 0 {
		element.Append(t.makeFrame())
	}
	dx := t.handlePWidth(element)
	t.handleLabels(element, labels, dx)
}
