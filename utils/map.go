package utils

import (
	"iter"
	"slices"
)

// Map is a generic key-value store with various utility methods
type Map[K comparable, V any] struct {
	Data map[K]V
}

// NewMap creates a new Map instance
func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{
		Data: make(map[K]V),
	}
}

// Slice returns all values in the model as a slice
func (m Map[K, V]) Slice() []V {
	result := make([]V, 0, len(m.Data))
	for _, v := range m.Data {
		result = append(result, v)
	}
	return result
}

// Keys returns all keys in the model as a slice
func (m Map[K, V]) Keys() []K {
	result := make([]K, 0, len(m.Data))
	for k := range m.Data {
		result = append(result, k)
	}
	return result
}

// KeysSorted returns all keys in the map as a sorted slice.
// Requires a less function to compare keys.
func (m Map[K, V]) KeysSorted(less func(a, b K) int) []K {
	keys := m.Keys()
	slices.SortFunc(keys, less)
	return keys
}

// Get retrieves a value by key, returning the value and whether it exists
func (m Map[K, V]) Get(key K) (V, bool) {
	val, exists := m.Data[key]
	return val, exists
}

// Set adds or updates a key-value pair in the model
func (m *Map[K, V]) Set(key K, value V) {
	m.Data[key] = value
}

// Iter returns an iterator for the model's values
func (m Map[K, V]) Iter() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, v := range m.Data {
			if !yield(v) {
				return
			}
		}
	}
}

// SortedIter returns an iterator for the map's values ordered by sorted keys.
// Requires a less function to compare keys.
func (m Map[K, V]) SortedIter(less func(a, b K) int) iter.Seq[V] {
	return func(yield func(V) bool) {
		keys := m.KeysSorted(less)
		for _, k := range keys {
			if !yield(m.Data[k]) {
				return
			}
		}
	}
}

// Items returns an iterator for the model's key-value pairs
func (m Map[K, V]) Items() iter.Seq[struct {
	Key   K
	Value V
}] {
	return func(yield func(struct {
		Key   K
		Value V
	}) bool) {
		for k, v := range m.Data {
			if !yield(struct {
				Key   K
				Value V
			}{k, v}) {
				return
			}
		}
	}
}

// Delete removes a key-value pair from the model
func (m *Map[K, V]) Delete(key K) {
	delete(m.Data, key)
}

// Len returns the number of items in the model
func (m Map[K, V]) Len() int {
	return len(m.Data)
}

// Clear removes all items from the model
func (m *Map[K, V]) Clear() {
	clear(m.Data) // Go 1.21+ feature
	// For older versions: m.Data = make(map[K]V)
}
