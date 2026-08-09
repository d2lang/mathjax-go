// Copyright 2017-2022 The MathJax Consortium
// SPDX-License-Identifier: Apache-2.0
// This file is a Go translation and modification of MathJax 3.2.2 behavior.

// Package ordered provides deterministic insertion-ordered string maps. They
// model the object-key and DOM-attribute ordering observable in MathJax 3.2.2.
package ordered

// Map is an insertion-ordered map with string keys. Updating a key preserves
// its original position; deleting and reinserting it appends it at the end.
// The zero value is ready for use.
type Map[V any] struct {
	keys   []string
	values map[string]V
}

// New returns an initialized empty map.
func New[V any]() *Map[V] {
	return &Map[V]{values: make(map[string]V)}
}

func (m *Map[V]) initialize() {
	if m.values == nil {
		m.values = make(map[string]V)
	}
}

// Len returns the number of entries.
func (m *Map[V]) Len() int {
	if m == nil {
		return 0
	}
	return len(m.values)
}

// Has reports whether key is present.
func (m *Map[V]) Has(key string) bool {
	if m == nil {
		return false
	}
	_, ok := m.values[key]
	return ok
}

// Get returns the value and whether key is present.
func (m *Map[V]) Get(key string) (V, bool) {
	var zero V
	if m == nil {
		return zero, false
	}
	v, ok := m.values[key]
	return v, ok
}

// Set inserts or updates key.
func (m *Map[V]) Set(key string, value V) {
	m.initialize()
	if _, ok := m.values[key]; !ok {
		m.keys = append(m.keys, key)
	}
	m.values[key] = value
}

// Delete removes key and reports whether it was present.
func (m *Map[V]) Delete(key string) bool {
	if m == nil {
		return false
	}
	if _, ok := m.values[key]; !ok {
		return false
	}
	delete(m.values, key)
	for i, candidate := range m.keys {
		if candidate == key {
			copy(m.keys[i:], m.keys[i+1:])
			m.keys = m.keys[:len(m.keys)-1]
			break
		}
	}
	return true
}

// Keys returns a copy of the keys in insertion order.
func (m *Map[V]) Keys() []string {
	if m == nil {
		return nil
	}
	return append([]string(nil), m.keys...)
}

// Range calls yield in insertion order until it returns false.
func (m *Map[V]) Range(yield func(key string, value V) bool) {
	if m == nil {
		return
	}
	for _, key := range m.keys {
		if !yield(key, m.values[key]) {
			return
		}
	}
}

// Clone returns an independent shallow copy.
func (m *Map[V]) Clone() *Map[V] {
	clone := New[V]()
	if m == nil {
		return clone
	}
	clone.keys = append(clone.keys, m.keys...)
	for key, value := range m.values {
		clone.values[key] = value
	}
	return clone
}
