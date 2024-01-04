// Package genericlist contains an implementation of a doubly-linked list
// that uses generics. The list is not go-routine safe.
package genericlist

// Element is an element of a doubly linked list
type Element[V any] struct {
	Value      V
	list       *List[V]
	next, prev *Element[V]
}

func (e *Element[V]) unlink() {
	if e.next != nil {
		e.next.prev = e.prev
	}

	if e.prev != nil {
		e.prev.next = e.next
	}
}

// Next returns the next list element or nil.
func (e *Element[V]) Next() *Element[V] {
	return e.next
}

// Prev returns the previous list element or nil.
func (e *Element[V]) Prev() *Element[V] {
	return e.prev
}

// List is a doubly linked list
type List[V any] struct {
	numElements int
	front, back *Element[V]
}

// New returns a new empty list.
func New[V any]() *List[V] {
	return &List[V]{}
}

// Front returns the first element of list l or nil if the list is empty.
func (l *List[V]) Front() *Element[V] {
	return l.front
}

// Back returns the last element of list l or nil if the list is empty.
func (l *List[V]) Back() *Element[V] {
	return l.back
}

// MoveAfter moves element e to its new position after mark. If e or
// mark is not an element of l, or e == mark, the list is not modified.
// The element and mark must not be nil.
func (l *List[V]) MoveAfter(e, mark *Element[V]) {
	if mark == nil || e.list != l || e == mark || mark.list != l || l.numElements == 0 {
		return
	}

	l.moveAfterInternal(e, mark)
}

func (l *List[V]) moveAfterInternal(e *Element[V], mark *Element[V]) {
	if mark == l.back {
		l.MoveToBack(e)
		return
	}

	e.unlink()

	// [mark] <-> [mark.next] becomes
	//
	// [e] <- [mark.next], [mark] -> [mark.next]
	if mark.next != nil {
		mark.next.prev = e
	}

	// [e] <-> [mark.next], [mark] -> [mark.next]
	e.next = mark.next

	// [mark] -> [e] <-> [mark.next]
	mark.next = e

	// [mark] <-> [e] <-> [mark.next]
	e.prev = mark
}

// MoveBefore moves element e to its new position before mark. If e or
// mark is not an element of l, or e == mark, the list is not modified.
// The element and mark must not be nil.
func (l *List[V]) MoveBefore(e, mark *Element[V]) {
	if mark == nil || e.list != l || e == mark || mark.list != l || l.numElements == 0 {
		return
	}

	l.moveBeforeInternal(e, mark)
}

func (l *List[V]) moveBeforeInternal(e *Element[V], mark *Element[V]) {
	if mark == l.front {
		l.MoveToFront(e)
		return
	}

	e.unlink()

	// [mark.prev] <-> [mark] becomes
	// [mark.prev] -> [e], [mark.prev] <- [mark]
	if mark.prev != nil {
		mark.prev.next = e
	}

	// [mark.prev] <-> [e], [mark.prev] <- [mark]
	e.prev = mark.prev

	// [mark.prev] <-> [e] <- [mark]
	mark.prev = e

	// [mark.prev] <-> [e] <-> [mark]
	e.next = mark
}

// MoveToBack moves element e to the back of list l. If e is not an
// element of l, the list is not modified. The element must not be nil.
func (l *List[V]) MoveToBack(e *Element[V]) {
	if e.list != l || e == l.back || l.numElements == 0 {
		return
	}

	l.moveToBackInternal(e)
}

func (l *List[V]) moveToBackInternal(e *Element[V]) {
	e.unlink()
	if l.back != nil {
		l.back.next = e
	}

	e.prev = l.back
	l.back = e

	if l.front == nil {
		l.front = e
	}
}

// MoveToFront moves element e to the front of list l. If e is not an
// element of l, the list is not modified. The element must not be nil.
func (l *List[V]) MoveToFront(e *Element[V]) {
	if e.list != l || e == l.front || l.numElements == 0 {
		return
	}

	l.moveToFrontInternal(e)
}

func (l *List[V]) moveToFrontInternal(e *Element[V]) {
	e.unlink()
	if l.front != nil {
		l.front.prev = e
	}

	e.next = l.front
	l.front = e

	if l.back == nil {
		l.back = e
	}
}

// Remove removes e from l if e is an element of list l. It returns the element
// value e.Value. The element must not be nil.
func (l *List[V]) Remove(e *Element[V]) V {
	if e.list != l || l.numElements == 0 {
		var noop V
		return noop
	}

	if e == l.front {
		l.front = l.front.next
	}

	if e == l.back {
		l.back = l.back.prev
	}

	e.unlink()
	l.numElements--
	return e.Value
}

// Len returns the number of elements of list l. The complexity is O(1).
func (l *List[V]) Len() int {
	return l.numElements
}

// PushFront inserts a new element e with value v at the front of list l and returns e.
func (l *List[V]) PushFront(v V) *Element[V] {
	e := &Element[V]{
		Value: v,
		list:  l,
	}

	l.numElements++
	l.moveToFrontInternal(e)
	if l.back == nil {
		l.back = e
	}
	return e
}

// PushBack inserts a new element e with value v at the back of list l and returns e.
func (l *List[V]) PushBack(v V) *Element[V] {
	e := &Element[V]{
		Value: v,
		list:  l,
	}

	l.numElements++
	l.moveToBackInternal(e)
	if l.front == nil {
		l.front = e
	}
	return e
}

// InsertBefore inserts a new element e with value v before the mark and returns e
func (l *List[V]) InsertBefore(v V, mark *Element[V]) *Element[V] {
	if mark == nil || mark.list != l {
		return nil
	}

	e := &Element[V]{
		Value: v,
		list:  l,
	}

	l.numElements++
	if l.numElements == 0 {
		l.front = e
		l.back = e
		return e
	}

	l.moveBeforeInternal(e, mark)
	return e
}

// InsertAfter inserts a new element e with value v after the mark and returns e
func (l *List[V]) InsertAfter(v V, mark *Element[V]) *Element[V] {
	if mark == nil || mark.list != l {
		return nil
	}

	e := &Element[V]{
		Value: v,
		list:  l,
	}

	l.numElements++
	if l.numElements == 0 {
		l.front = e
		l.back = e
		return e
	}

	l.moveAfterInternal(e, mark)
	return e
}
