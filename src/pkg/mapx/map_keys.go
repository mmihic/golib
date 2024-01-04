// Package mapx has helper utilities for dealing with maps.
package mapx

// Keys returns the keys of the map, in any order.
func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
