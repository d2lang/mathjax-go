// SPDX-License-Identifier: Apache-2.0
package layout

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestBBoxOriginalArithmetic(t *testing.T) {
	f, err := os.Open("../../testdata/bbox_arithmetic_mathjax_3_2_2.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	type dimensions struct{ W, H, D float64 }
	var fixture struct {
		MathjaxGitCommit, CombineSource, AppendSource string
		BBoxCases                                     []struct {
			Name, Method string
			Parent       dimensions
			Child        struct{ W, H, D, L, R, RScale float64 }
			X, Y         float64
			Expected     dimensions
		}
	}
	if err := json.NewDecoder(z).Decode(&fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.MathjaxGitCommit != "ad8f5c21cb810236551da8c6512ba733e67357ee" || fixture.CombineSource == "" || fixture.AppendSource == "" || len(fixture.BBoxCases) != 300 {
		t.Fatal("unbound original BBox arithmetic")
	}
	names := make(map[string]bool)
	for _, c := range fixture.BBoxCases {
		if c.Name == "" || names[c.Name] {
			t.Fatal("invalid original BBox case", c.Name)
		}
		names[c.Name] = true
		t.Run(c.Name, func(t *testing.T) {
			parent := NewBBox(c.Parent.W, c.Parent.H, c.Parent.D)
			child := NewBBox(c.Child.W, c.Child.H, c.Child.D)
			child.L, child.R, child.RScale = c.Child.L, c.Child.R, c.Child.RScale
			before := *child
			switch c.Method {
			case "combine":
				parent.Combine(child, c.X, c.Y)
			case "append":
				parent.Append(child)
			default:
				t.Fatal("unknown original method", c.Method)
			}
			got := dimensions{parent.W, parent.H, parent.D}
			if got != c.Expected {
				t.Fatalf("original arithmetic differs: got %#v want %#v", got, c.Expected)
			}
			if !reflect.DeepEqual(*child, before) {
				t.Fatal("child BBox mutated")
			}
		})
	}
}
