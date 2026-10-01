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

func TestJavaScriptKeys(t *testing.T) {
	insertion := []string{"z", "4294967295", "10", "01", "2", "0", "00", "-0", "+2", "2.0", "2e0", "4294967294", "a"}
	m := New[int]()
	for i, key := range insertion {
		m.Set(key, i)
	}
	m.Set("10", 20)
	want := []string{"0", "2", "10", "4294967294", "z", "4294967295", "01", "00", "-0", "+2", "2.0", "2e0", "a"}
	if got := m.JavaScriptKeys(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Object.keys order = %v, want %v", got, want)
	}
	if got := m.Keys(); !reflect.DeepEqual(got, insertion) {
		t.Fatalf("insertion order changed: %v, want %v", got, insertion)
	}
	keys := m.JavaScriptKeys()
	keys[0] = "changed"
	if !reflect.DeepEqual(m.JavaScriptKeys(), want) {
		t.Fatal("JavaScriptKeys does not return an independent copy")
	}
	m.Delete("2")
	m.Set("2", 30)
	if !reflect.DeepEqual(m.JavaScriptKeys(), want) {
		t.Fatal("reinserted array index lost numeric ordering")
	}
}
