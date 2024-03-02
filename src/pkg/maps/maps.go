// Package maps has helper utilities for dealing with maps.
package maps

// Keys returns the keys of the map, in any order.
func Keys[K comparable, V any, M ~map[K]V](m M) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Values returns all values in a map.
func Values[K comparable, V any, M ~map[K]V](m map[K]V) []V {
	vals := make([]V, 0, len(m))
	for _, v := range m {
		vals = append(vals, v)
	}

	return vals
}
