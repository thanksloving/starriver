package dag

import (
	"github.com/thanksloving/starriver"
)

// Set is a set data structure.
type (
	Set map[string]GraphObject

	GraphObject = starriver.GraphObject
)

// Add adds an item to the set
func (s Set) Add(o GraphObject) {
	s[o.ID()] = o
}

// Delete removes an item from the set.
func (s Set) Delete(o GraphObject) {
	delete(s, o.ID())
}

// Include returns true/false of whether a value is in the set.
func (s Set) Include(o GraphObject) bool {
	_, ok := s[o.ID()]
	return ok
}

// Intersection computes the set intersection with other.
func (s Set) Intersection(other Set) Set {
	result := make(Set)
	if s == nil || other == nil {
		return result
	}
	// Iteration over a smaller set has better performance.
	if other.Len() < s.Len() {
		s, other = other, s
	}
	for _, v := range s {
		if other.Include(v) {
			result.Add(v)
		}
	}
	return result
}

// Difference returns a set with the elements that s has but
// other doesn't.
func (s Set) Difference(other Set) Set {
	if other == nil || other.Len() == 0 {
		return s.Copy()
	}

	result := make(Set)
	for k, v := range s {
		if _, ok := other[k]; !ok {
			result.Add(v)
		}
	}

	return result
}

// Filter returns a set that contains the elements from the receiver
// where the given callback returns true.
func (s Set) Filter(cb func(GraphObject) bool) Set {
	result := make(Set)

	for _, v := range s {
		if cb(v) {
			result.Add(v)
		}
	}

	return result
}

// Len is the number of items in the set.
func (s Set) Len() int {
	return len(s)
}

// List returns the list of set elements.
func (s Set) List() []GraphObject {
	if s == nil {
		return nil
	}

	r := make([]GraphObject, 0, len(s))
	for _, v := range s {
		r = append(r, v)
	}

	return r
}

// Copy returns a shallow copy of the set.
func (s Set) Copy() Set {
	c := make(Set, len(s))
	for k, v := range s {
		c[k] = v
	}
	return c
}
