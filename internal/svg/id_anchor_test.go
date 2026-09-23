// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package svg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"testing"

	"github.com/d2lang/mathjax-go/internal/layout"
	"github.com/d2lang/mathjax-go/internal/mml"
	"github.com/d2lang/mathjax-go/internal/pipeline"
	"github.com/d2lang/mathjax-go/internal/tex"
)

type idAnchorNumber float64

func (n *idAnchorNumber) UnmarshalJSON(data []byte) error {
	var marker bytes.Buffer
	if json.Compact(&marker, data) == nil && marker.String() == `{"negativeZero":true}` {
		*n = idAnchorNumber(math.Copysign(0, -1))
		return nil
	}
	var value float64
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*n = idAnchorNumber(value)
	return nil
}

type idAnchorCall struct {
	X, Y            idAnchorNumber
	ElementArgument string
}

type idAnchorSpec struct {
	Name, Method                       string
	X, Y                               idAnchorNumber
	Dx, H, Rscale                      float64
	DOMID                              *string
	ExistingTransform, ElementArgument string
	Children                           []json.RawMessage
	Calls                              json.RawMessage
	NodeID                             struct {
		Present, AttributesObjectAbsent bool
		Channel                         string
		Value                           any
	}
}

type idAnchorMovement struct {
	BeforePath, AfterPath []int
}

type idAnchorPrivateCase struct {
	Name          string
	Spec          idAnchorSpec
	ExpectedScene string
	Calls         []struct {
		BeforeScene, AfterScene string
		OriginalElements        []idAnchorMovement
	}
	FirstChildPath      *[]int
	TableBackgroundPath *[]int
}

func idAnchorPrivateCases(t *testing.T) []idAnchorPrivateCase {
	t.Helper()
	var fixture struct {
		MathJaxGitCommit string
		Cases            []idAnchorPrivateCase
	}
	data, err := os.ReadFile("testdata/id_anchor_private_mathjax_3_2_2.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathJaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || len(fixture.Cases) != 38 {
		t.Fatal("unbound primary ID-anchor private fixtures")
	}
	return fixture.Cases
}

func idAnchorChild(t *testing.T, raw json.RawMessage) *Element {
	t.Helper()
	var name string
	if json.Unmarshal(raw, &name) == nil {
		switch name {
		case "alignmentText":
			return NewElement("text", Text("")).SetAttr("data-id-align", "true")
		case "emptyMarkerText":
			return NewElement("text", Text("")).SetAttr("data-id-align", "")
		case "hitbox":
			return NewElement("rect").SetAttr("data-hitbox", "true").SetAttr("fill", "none").SetAttr("stroke", "none").SetAttr("width", "100").SetAttr("height", "200")
		case "background":
			return NewElement("rect").SetAttr("data-bgcolor", "true").SetAttr("fill", "yellow").SetAttr("width", "100").SetAttr("height", "200")
		case "path-A":
			return NewElement("path").SetAttr("data-fixture", name).SetAttr("d", "M0 0L1 1")
		default:
			return NewElement("g").SetAttr("data-fixture", name)
		}
	}
	var group struct{ IDBox []json.RawMessage }
	if err := json.Unmarshal(raw, &group); err != nil {
		t.Fatal(err)
	}
	element := NewElement("g").SetAttr("data-idbox", "true")
	for _, child := range group.IDBox {
		element.Append(idAnchorChild(t, child))
	}
	return element
}

func idAnchorPrivateWrapper(t *testing.T, s idAnchorSpec) (*wrapper, *mml.Node, *Element, *Element) {
	t.Helper()
	root, err := tex.NewCompiler().Compile("x", true)
	if err != nil {
		t.Fatal(err)
	}
	prepareTeXClasses(root)
	options := pipeline.DefaultOptions()
	renderer := &renderer{options: options, params: layout.TeXParameters, pxPerEm: options.Ex / layout.TeXParameters.XHeight}
	rootWrapper := renderer.wrap(root, nil, 0, true)
	var w *wrapper
	var find func(*wrapper)
	find = func(candidate *wrapper) {
		if (!s.NodeID.AttributesObjectAbsent && candidate.node.Kind == "mi") ||
			(s.NodeID.AttributesObjectAbsent && candidate.node.Kind == "text") {
			w = candidate
		}
		for _, child := range candidate.children {
			find(child)
		}
	}
	find(rootWrapper)
	if w == nil {
		t.Fatal("actual mi/text wrapper is absent")
	}
	// The primary text node has no Attributes object; Go's genuine text node
	// retains an empty one. Both paths have no effective id, without fabricating
	// a replacement node or adding a production-only test API.
	if s.NodeID.Present {
		if s.NodeID.Channel == "inherited" {
			w.node.Attributes.SetInherited("id", s.NodeID.Value)
		} else {
			w.node.Attributes.Set("id", s.NodeID.Value)
		}
	}
	own := NewElement("g")
	for _, child := range s.Children {
		own.Append(idAnchorChild(t, child))
	}
	w.element = own
	if s.DOMID != nil {
		own.SetAttr("id", *s.DOMID)
	}
	if s.ExistingTransform != "" {
		own.SetAttr("transform", s.ExistingTransform)
	}
	other := NewElement("g", NewElement("path").SetAttr("data-fixture", "other-path")).SetAttr("data-fixture", "other")
	scene := NewElement("svg", own, other)
	w.dx = s.Dx
	if s.Method == "place" {
		w.bbox.H = s.H
		w.bbox.RScale = s.Rscale
		w.bboxComputed = true
	}
	return w, root, scene, other
}

func idAnchorElements(root *Element) map[string]*Element {
	out := map[string]*Element{}
	var visit func(*Element, []int)
	visit = func(element *Element, path []int) {
		out[fmt.Sprint(path)] = element
		for index, child := range element.Children {
			if child, ok := child.(*Element); ok {
				visit(child, append(append([]int(nil), path...), index))
			}
		}
	}
	visit(root, nil)
	return out
}

func idAnchorSameElements(before, after map[string]*Element) bool {
	if len(before) != len(after) {
		return false
	}
	for path, element := range before {
		if after[path] != element {
			return false
		}
	}
	return true
}

func idAnchorAt(t *testing.T, root *Element, path *[]int) Node {
	t.Helper()
	if path == nil {
		return nil
	}
	var current Node = root
	for _, index := range *path {
		element, ok := current.(*Element)
		if !ok || index < 0 || index >= len(element.Children) {
			t.Fatalf("invalid primary identity path %v", *path)
		}
		current = element.Children[index]
	}
	return current
}

type idAnchorModelIdentity struct {
	Node, Parent *mml.Node
	Children     []*mml.Node
}

func idAnchorModelGraph(root *mml.Node) []idAnchorModelIdentity {
	var graph []idAnchorModelIdentity
	root.Walk(func(n *mml.Node) bool {
		graph = append(graph, idAnchorModelIdentity{n, n.Parent, append([]*mml.Node(nil), n.Children...)})
		return true
	})
	return graph
}

func idAnchorSameModelGraph(before, after []idAnchorModelIdentity) bool {
	if len(before) != len(after) {
		return false
	}
	for i, n := range before {
		other := after[i]
		if n.Node != other.Node || n.Parent != other.Parent || len(n.Children) != len(other.Children) {
			return false
		}
		for j, child := range n.Children {
			if child != other.Children[j] {
				return false
			}
		}
	}
	return true
}

func idAnchorSameWrapper(before wrapper, after *wrapper) bool {
	if before.node != after.node || before.parent != after.parent || before.renderer != after.renderer ||
		before.element != after.element || before.bbox != after.bbox || len(before.children) != len(after.children) {
		return false
	}
	for i, child := range before.children {
		if child != after.children[i] {
			return false
		}
	}
	return reflect.DeepEqual(before, *after)
}

func TestIDAnchorPlacementPrimary(t *testing.T) {
	count, calls := 0, 0
	for _, c := range idAnchorPrivateCases(t) {
		if c.Spec.Method != "place" {
			continue
		}
		count++
		t.Run(c.Name, func(t *testing.T) {
			w, root, scene, other := idAnchorPrivateWrapper(t, c.Spec)
			model, graph := root.Clone(), idAnchorModelGraph(root)
			own := w.element
			steps := []idAnchorCall{{c.Spec.X, c.Spec.Y, c.Spec.ElementArgument}}
			if len(c.Spec.Calls) != 0 {
				if err := json.Unmarshal(c.Spec.Calls, &steps); err != nil {
					t.Fatal(err)
				}
			}
			if len(steps) != len(c.Calls) {
				t.Fatal("primary placement call count differs")
			}
			for index, step := range steps {
				calls++
				want := c.Calls[index]
				if scene.String() != want.BeforeScene {
					t.Fatalf("call %d input scene differs from primary", index)
				}
				before := idAnchorElements(scene)
				x, y := float64(step.X), float64(step.Y)
				if c.Name == "own-negative-zero" {
					// Retain the recorded -0/-0 call without a numeric tolerance.
					if !math.Signbit(x) {
						t.Fatal("negative zero input lost its sign")
					}
					y = math.Copysign(0, -1)
				}
				switch step.ElementArgument {
				case "omitted", "undefined":
					w.place(x, y)
				case "null":
					w.place(x, y, nil)
				case "sameOwnElement":
					w.place(x, y, own)
				case "otherElement":
					w.place(x, y, other)
				default:
					t.Fatalf("unknown element argument %q", step.ElementArgument)
				}
				if scene.String() != want.AfterScene {
					t.Fatalf("call %d complete primary scene differs\ngot: %s\nwant: %s", index, scene.String(), want.AfterScene)
				}
				after := idAnchorElements(scene)
				if len(before) != len(want.OriginalElements) {
					t.Fatal("incomplete primary original-element identity inventory")
				}
				for _, movement := range want.OriginalElements {
					original := before[fmt.Sprint(movement.BeforePath)]
					if original == nil || after[fmt.Sprint(movement.AfterPath)] != original {
						t.Fatalf("call %d replaced/reordered original object at %v -> %v", index, movement.BeforePath, movement.AfterPath)
					}
				}
				if w.element != own || !idAnchorSameModelGraph(graph, idAnchorModelGraph(root)) || !reflect.DeepEqual(model, root.Clone()) {
					t.Fatal("placement changed owned element identity or MathML objects/state")
				}
			}
			if scene.String() != c.ExpectedScene {
				t.Fatal("complete final primary scene differs")
			}
		})
	}
	if count != 30 || calls != 33 {
		t.Fatalf("placement coverage changed: %d scenes, %d calls", count, calls)
	}
}

func TestIDAnchorFirstChildPrimary(t *testing.T) {
	count := 0
	for _, c := range idAnchorPrivateCases(t) {
		if c.Spec.Method != "firstChild" {
			continue
		}
		count++
		t.Run(c.Name, func(t *testing.T) {
			w, root, scene, _ := idAnchorPrivateWrapper(t, c.Spec)
			model, graph := root.Clone(), idAnchorModelGraph(root)
			before, elements := scene.String(), idAnchorElements(scene)
			state, bbox := *w, w.bbox.Clone()
			wantChild := idAnchorAt(t, w.element, c.FirstChildPath)
			if got := w.firstChild(); got != wantChild {
				t.Fatalf("firstChild returned a different object: got %p, want %p", got, wantChild)
			}
			if before != scene.String() || !idAnchorSameElements(elements, idAnchorElements(scene)) ||
				!idAnchorSameWrapper(state, w) || !reflect.DeepEqual(bbox, w.bbox) ||
				!idAnchorSameModelGraph(graph, idAnchorModelGraph(root)) || !reflect.DeepEqual(model, root.Clone()) {
				t.Fatal("firstChild mutated SVG, wrapper, or MathML state/identity")
			}
			// tableBackground has an additional rect/data-bgcolor predicate.
			// Its source-bound expected path is separate from primary firstChild.
			wantBackground := idAnchorAt(t, w.element, c.TableBackgroundPath)
			gotBackground := tableBackground(w.element)
			if (gotBackground == nil) != (wantBackground == nil) || (gotBackground != nil && Node(gotBackground) != wantBackground) {
				t.Fatal("separate tableBackground predicate/identity differs")
			}
			if before != scene.String() || before != c.ExpectedScene ||
				!idAnchorSameElements(elements, idAnchorElements(scene)) || !idAnchorSameWrapper(state, w) ||
				!reflect.DeepEqual(bbox, w.bbox) || !idAnchorSameModelGraph(graph, idAnchorModelGraph(root)) ||
				!reflect.DeepEqual(model, root.Clone()) {
				t.Fatal("tableBackground mutated SVG, wrapper, or MathML state/identity")
			}
		})
	}
	if count != 8 {
		t.Fatalf("firstChild coverage changed: %d", count)
	}
}
