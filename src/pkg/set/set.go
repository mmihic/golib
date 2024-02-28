// Package set contains a basic generic Set type.
package set

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

// A Set is an unordered collection of values with the ability
// to test for inclusion and exclusion.
type Set[E comparable] map[E]struct{}

// NewSet creates a new set from a collection of values.
func NewSet[E comparable](vals ...E) Set[E] {
	set := make(Set[E], len(vals))
	for _, v := range vals {
		set.Add(v)
	}
	return set
}

// Add adds values to the set.
func (set Set[E]) Add(vals ...E) Set[E] {
	for _, v := range vals {
		set[v] = struct{}{}
	}
	return set
}

// Del removes values from the set.
func (set Set[E]) Del(vals ...E) Set[E] {
	for _, v := range vals {
		delete(set, v)
	}
	return set
}

// Has returns true if the set contains the given value.
func (set Set[E]) Has(val E) bool {
	_, ok := set[val]
	return ok
}

// All returns all values in the set.
func (set Set[E]) All() []E {
	result := make([]E, 0, len(set))
	for k := range set {
		result = append(result, k)
	}
	return result
}

// Len returns the size of the set.
func (set Set[E]) Len() int {
	return len(set)
}

// Intersect computes the intersection of two sets, producing
// a set that only contains the values present in both sets.
func (set Set[E]) Intersect(other Set[E]) Set[E] {
	result := NewSet[E]()
	for v := range set {
		if other.Has(v) {
			result.Add(v)
		}
	}

	return result
}

// Union computes the union of two sets, producing a set
// that contains the values present in either set.
func (set Set[E]) Union(other Set[E]) Set[E] {
	result := NewSet[E]()
	for v := range set {
		result.Add(v)
	}

	for v := range other {
		result.Add(v)
	}

	return result
}

// Contains returns true if one set contains all
// elements of the other.
func (set Set[E]) Contains(other Set[E]) bool {
	for v := range other {
		if !set.Has(v) {
			return false
		}
	}

	return true
}

// Equal compares two sets for exact equality.
func (set Set[E]) Equal(other Set[E]) bool {
	// Two sets exactly match if both are the same size and one
	// contains all elements of the other
	return len(set) == len(other) && set.Contains(other)
}

// UnmarshalJSON unmarshals a set from a JSON array.
func (set *Set[E]) UnmarshalJSON(b []byte) error {
	var val []E
	if err := json.Unmarshal(b, &val); err != nil {
		return err
	}

	*set = NewSet(val...)
	return nil
}

// MarshalJSON marshals a set as a JSON array.
func (set Set[E]) MarshalJSON() ([]byte, error) {
	return json.Marshal(set.All())
}

// UnmarshalYAML unmarshals a set from a YAML array.
func (set *Set[E]) UnmarshalYAML(n *yaml.Node) error {
	var val []E
	if err := n.Decode(&val); err != nil {
		return err
	}

	*set = NewSet(val...)
	return nil
}

// MarshalYAML marshals a set as a YAML array.
func (set Set[E]) MarshalYAML() (any, error) {
	return set.All(), nil
}
