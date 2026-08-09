// Copyright 2018-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2.

package svg

import (
	"math"
	"strings"

	"github.com/d2lang/mathjax-go/internal/layout"
)

// multiscriptPair is one sub/sup pair.  A nil entry is the empty mrow that
// MathJax inserts when fixMmultiscripts repairs an odd script list.
type multiscriptPair struct {
	sub, sup       *wrapper
	subbox, supbox *layout.BBox
	width          float64
}

type multiscriptData struct {
	base      *wrapper
	post, pre []multiscriptPair
	postSub   *layout.BBox
	postSup   *layout.BBox
	preSub    *layout.BBox
	preSup    *layout.BBox
}

func (w *wrapper) getMultiscriptData() multiscriptData {
	data := multiscriptData{
		postSub: layout.EmptyBBox(), postSup: layout.EmptyBBox(),
		preSub: layout.EmptyBBox(), preSup: layout.EmptyBBox(),
	}
	if len(w.children) == 0 {
		return data
	}
	data.base = w.children[0]
	prescripts, sub := false, true
	for _, child := range w.children[1:] {
		if child.node.Kind == "mprescripts" {
			prescripts, sub = true, true
			continue
		}
		pairs := &data.post
		if prescripts {
			pairs = &data.pre
		}
		if sub {
			*pairs = append(*pairs, multiscriptPair{sub: child})
		} else {
			if len(*pairs) == 0 {
				*pairs = append(*pairs, multiscriptPair{})
			}
			(*pairs)[len(*pairs)-1].sup = child
		}
		sub = !sub
	}
	data.postSub, data.postSup = combineMultiscriptPairs(data.post)
	data.preSub, data.preSup = combineMultiscriptPairs(data.pre)
	return data
}

func emptyScriptBBox() *layout.BBox { return layout.ZeroBBox() }

func scriptPairBBox(child *wrapper) *layout.BBox {
	if child == nil {
		return emptyScriptBBox()
	}
	return child.outerBBox()
}

func combineMultiscriptPairs(pairs []multiscriptPair) (*layout.BBox, *layout.BBox) {
	sub, sup := layout.EmptyBBox(), layout.EmptyBBox()
	for i := range pairs {
		pair := &pairs[i]
		pair.subbox, pair.supbox = scriptPairBBox(pair.sub), scriptPairBBox(pair.sup)
		sw := pair.subbox.W * pair.subbox.RScale
		uw := pair.supbox.W * pair.supbox.RScale
		pair.width = math.Max(sw, uw)
		sub.W += pair.width
		sup.W += pair.width
		sub.H = math.Max(sub.H, pair.subbox.H*pair.subbox.RScale)
		sub.D = math.Max(sub.D, pair.subbox.D*pair.subbox.RScale)
		sup.H = math.Max(sup.H, pair.supbox.H*pair.supbox.RScale)
		sup.D = math.Max(sup.D, pair.supbox.D*pair.supbox.RScale)
	}
	return sub, sup
}

func combineMultiscriptPrePost(pre, post *layout.BBox) *layout.BBox {
	combined := layout.NewBBox(pre.W, pre.H, pre.D)
	combined.Combine(post, 0, 0)
	return combined
}

// multiscriptPrimary is CommonScriptbase.scriptChild.  MathJax 3.2.2 keeps
// that inherited getter for mmultiscripts, so the first post-script supplies
// the initial TeX drop calculation even when the combined lists are wider.
func (w *wrapper) multiscriptPrimary() *wrapper {
	if len(w.children) < 2 {
		return nil
	}
	return w.children[1]
}

func (w *wrapper) multiscriptGetU() float64 {
	primary := w.multiscriptPrimary()
	if primary == nil {
		return 0
	}
	return w.supShift(primary)
}

func (w *wrapper) multiscriptGetV() float64 {
	primary := w.multiscriptPrimary()
	if primary == nil {
		return 0
	}
	return w.subShift(primary, w.renderer.params.Sub1)
}

// multiscriptUVQ returns the superscript and subscript baseline offsets.  The
// single-script branches intentionally preserve MathJax 3.2.2's source-shaped
// CommonMmultiscripts.getUVQ behavior.
func (w *wrapper) multiscriptUVQ(subbox, supbox *layout.BBox) (u, v, q float64) {
	if subbox.H == 0 && subbox.D == 0 {
		return w.multiscriptGetU(), 0, 0
	}
	if supbox.H == 0 && supbox.D == 0 {
		return -w.multiscriptGetV(), 0, 0
	}

	primary := w.multiscriptPrimary()
	if primary == nil || w.scriptBaseCore() == nil {
		return 0, 0, 0
	}
	params := w.renderer.params
	u = w.multiscriptGetU()
	drop := w.baseCharZero(w.scriptBaseCore().outerBBox().D*w.baseScale() + params.SubDrop*subbox.RScale)
	shift := layout.Length2Em(attribute(w.node, "subscriptshift", ""), params.Sub2, w.bbox.Scale, w.renderer.pxPerEm)
	down := math.Max(drop, shift)
	minimum := 3 * params.Rule
	q = (u - supbox.D*supbox.RScale) - (subbox.H*subbox.RScale - down)
	if q < minimum {
		down += minimum - q
		p := .8*params.XHeight - (u - supbox.D*supbox.RScale)
		if p > 0 {
			u += p
			down -= p
		}
	}
	u = math.Max(layout.Length2Em(attribute(w.node, "superscriptshift", ""), u, w.bbox.Scale, w.renderer.pxPerEm), u)
	down = math.Max(layout.Length2Em(attribute(w.node, "subscriptshift", ""), down, w.bbox.Scale, w.renderer.pxPerEm), down)
	q = (u - supbox.D*supbox.RScale) - (subbox.H*subbox.RScale - down)
	return u, -down, q
}

func (w *wrapper) computeMultiscriptsBBox(bbox *layout.BBox) {
	data := w.getMultiscriptData()
	if data.base == nil {
		bbox.Empty()
		bbox.Clean()
		return
	}
	sub := combineMultiscriptPrePost(data.preSub, data.postSub)
	sup := combineMultiscriptPrePost(data.preSup, data.postSup)
	u, v, _ := w.multiscriptUVQ(sub, sup)
	scriptspace := w.renderer.params.ScriptSpace

	bbox.Empty()
	if len(data.pre) != 0 {
		bbox.Combine(data.preSup, scriptspace, u)
		bbox.Combine(data.preSub, scriptspace, v)
	}
	bbox.Append(data.base.outerBBox())
	if len(data.post) != 0 {
		x := bbox.W
		bbox.Combine(data.postSup, x, u)
		bbox.Combine(data.postSub, x, v)
		bbox.W += scriptspace
	}
	bbox.Clean()
}

func multiscriptAlign(name string, width, column float64) float64 {
	switch name {
	case "center":
		return (column - width) / 2
	case "right":
		return column - width
	default:
		return 0
	}
}

func (w *wrapper) multiscriptAlignment() (pre, post string) {
	value := "right left"
	if property, ok := w.node.Property("scriptalign"); ok {
		if text, ok := property.(string); ok && text != "" {
			value = text
		}
	}
	parts := strings.Fields(value + " " + value)
	return parts[0], parts[1]
}

func (w *wrapper) addMultiscripts(element *Element, pairs []multiscriptPair, x, u, v float64, align string) float64 {
	supRow := NewElement("g")
	subRow := NewElement("g")
	element.Append(supRow, subRow)
	w.place(x, u, supRow)
	w.place(x, v, subRow)
	dx := 0.0
	for _, pair := range pairs {
		subbox, supbox := scriptPairBBox(pair.sub), scriptPairBBox(pair.sup)
		width := math.Max(subbox.W*subbox.RScale, supbox.W*supbox.RScale)
		if pair.sub != nil {
			pair.sub.toSVG(subRow)
			pair.sub.place(dx+multiscriptAlign(align, subbox.W*subbox.RScale, width), 0, pair.sub.element)
		}
		if pair.sup != nil {
			pair.sup.toSVG(supRow)
			pair.sup.place(dx+multiscriptAlign(align, supbox.W*supbox.RScale, width), 0, pair.sup.element)
		}
		dx += width
	}
	return x + dx
}

func (w *wrapper) multiscriptsToSVG(parent *Element) {
	element := w.standardSVG(parent)
	data := w.getMultiscriptData()
	if data.base == nil {
		return
	}
	preAlign, postAlign := w.multiscriptAlignment()
	sub := combineMultiscriptPrePost(data.preSub, data.postSub)
	sup := combineMultiscriptPrePost(data.preSup, data.postSup)
	u, v, _ := w.multiscriptUVQ(sub, sup)

	x := 0.0
	if len(data.pre) != 0 {
		x = w.addMultiscripts(element, data.pre, w.renderer.params.ScriptSpace, u, v, preAlign)
	}
	data.base.toSVG(element)
	data.base.place(x, 0, data.base.element)
	x += data.base.outerBBox().W
	if len(data.post) != 0 {
		w.addMultiscripts(element, data.post, x, u, v, postAlign)
	}
}
