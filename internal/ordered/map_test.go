// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0

package ordered

import (
	"reflect"
	"testing"
)

func TestMapOrder(t *testing.T) {
	m := New[int]()
	m.Set("b", 1)
	m.Set("a", 2)
	m.Set("b", 3)
	if got, want := m.Keys(), []string{"b", "a"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("keys = %v, want %v", got, want)
	}
	m.Delete("b")
	m.Set("b", 4)
	if got, want := m.Keys(), []string{"a", "b"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("keys after reinsertion = %v, want %v", got, want)
	}
}

func TestZeroMap(t *testing.T) {
	var m Map[string]
	m.Set("x", "y")
	if got, ok := m.Get("x"); !ok || got != "y" {
		t.Fatalf("Get(x) = %q, %v", got, ok)
	}
}
