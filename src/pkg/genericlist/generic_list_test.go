package genericlist

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mmihic/golib/src/pkg/slices"
)

func TestList_NotMyList(t *testing.T) {
	l1, l2 := New[string](), New[string]()

	e1 := l1.PushFront("my-string")
	e2 := l2.PushFront("my-string-in-another-list")

	// Inserting at an element from another list should do nothing
	require.Nil(t, l2.InsertBefore("my-other-string", e1))
	require.Nil(t, l2.InsertAfter("my-other-string", e1))

	// Moving before or after an element from another list should do nothing
	l2.MoveAfter(e2, e1)
	require.Equal(t, 1, l2.Len())
	require.Equal(t, "my-string-in-another-list", l2.Front().Value)
	require.Equal(t, "my-string-in-another-list", l2.Back().Value)

	l2.MoveBefore(e2, e1)
	require.Equal(t, 1, l2.Len())
	require.Equal(t, "my-string-in-another-list", l2.Front().Value)
	require.Equal(t, "my-string-in-another-list", l2.Back().Value)

	// Moving an element from another list should do nothing
	l2.MoveAfter(e1, e2)
	require.Equal(t, 1, l2.Len())
	require.Equal(t, "my-string-in-another-list", l2.Front().Value)
	require.Equal(t, "my-string-in-another-list", l2.Back().Value)

	l2.MoveBefore(e1, e2)
	require.Equal(t, 1, l2.Len())
	require.Equal(t, "my-string-in-another-list", l2.Front().Value)
	require.Equal(t, "my-string-in-another-list", l2.Back().Value)

	// Removing an element from another list should do nothing
	l2.Remove(e1)
	require.Equal(t, 1, l2.Len())
	require.Equal(t, "my-string-in-another-list", l2.Front().Value)
	require.Equal(t, "my-string-in-another-list", l2.Back().Value)

	// Moving an element to the front or back should do nothing if it doesn't
	// come from this list
	l2.MoveToFront(e1)
	require.Equal(t, 1, l2.Len())
	require.Equal(t, "my-string-in-another-list", l2.Front().Value)
	require.Equal(t, "my-string-in-another-list", l2.Back().Value)

	l2.MoveToBack(e1)
	require.Equal(t, 1, l2.Len())
	require.Equal(t, "my-string-in-another-list", l2.Front().Value)
	require.Equal(t, "my-string-in-another-list", l2.Back().Value)
}

func TestList_EmptyList(t *testing.T) {
	l := New[string]()

	// Empty list
	require.Equal(t, 0, l.Len())
	require.Nil(t, l.Front())
	require.Nil(t, l.Back())
}

func TestList_PushFront(t *testing.T) {
	input := []string{"foo", "bar", "zed", "mork", "ork", "banana"}

	l := New[string]()
	testPushElement(t, l, l.PushFront, input, slices.Reversed(input))
}

func TestList_PushBack(t *testing.T) {
	input := []string{"foo", "bar", "zed", "mork", "ork", "banana"}

	l := New[string]()
	testPushElement(t, l, l.PushBack, input, input)
}

func testPushElement(t *testing.T, l *List[string],
	pushFn func(string) *Element[string], input, fromFront []string) {

	// Pushing the first element should set both front and back
	e := pushFn(input[0])
	require.Equal(t, 1, l.Len())
	require.NotNil(t, e)
	require.Equal(t, l.Front(), e)
	require.Equal(t, l.Back(), e)
	require.Nil(t, e.Prev())
	require.Nil(t, e.Next())

	// Push a bunch of others and then conform they are in the right order
	for i := 1; i < len(input); i++ {
		val := input[i]
		e = pushFn(val)
		require.Equal(t, i+1, l.Len())
		require.NotNil(t, e)
		require.Equal(t, val, e.Value)
	}

	requireListEquals(t, l, fromFront)
}

func TestList_InsertBefore(t *testing.T) {
	l := New[string]()
	l.PushFront("foo")

	// Insert before the front, should become the new front
	l.InsertBefore("bar", l.Front())
	require.Equal(t, 2, l.Len())
	require.Equal(t, "bar", l.Front().Value)
	require.Equal(t, "foo", l.Back().Value)
	requireListEquals(t, l, []string{"bar", "foo"})

	// Insert before the tail
	l.InsertBefore("zed", l.Back())
	require.Equal(t, 3, l.Len())
	require.Equal(t, "bar", l.Front().Value)
	require.Equal(t, "foo", l.Back().Value)
	requireListEquals(t, l, []string{"bar", "zed", "foo"})

	// Insert somewhere in the middle
	l.InsertBefore("quork", l.Front().Next())
	require.Equal(t, 4, l.Len())
	require.Equal(t, "bar", l.Front().Value)
	require.Equal(t, "foo", l.Back().Value)
	requireListEquals(t, l, []string{"bar", "quork", "zed", "foo"})
}

func TestList_InsertAfter(t *testing.T) {
	l := New[string]()
	l.PushFront("foo")

	// Insert after the back, should become the new back
	l.InsertAfter("bar", l.Back())
	require.Equal(t, "foo", l.Front().Value)
	require.Equal(t, "bar", l.Back().Value)
	requireListEquals(t, l, []string{"foo", "bar"})

	// Insert after the front, should insert into the middle
	l.InsertAfter("zed", l.Front())
	require.Equal(t, "foo", l.Front().Value)
	require.Equal(t, "bar", l.Back().Value)
	requireListEquals(t, l, []string{"foo", "zed", "bar"})

	// Insert somewhere in the middle
	l.InsertAfter("quork", l.Front().Next())
	require.Equal(t, "foo", l.Front().Value)
	require.Equal(t, "bar", l.Back().Value)
	requireListEquals(t, l, []string{"foo", "zed", "quork", "bar"})
}

func requireListEquals[V any](t *testing.T, l *List[V], fromFront []V) {
	// Confirm result iterating from front
	e := l.Front()
	for i, expected := range fromFront {
		require.NotNil(t, e, "element %d is nil", i)
		require.Equal(t, expected, e.Value, "unexpected value for %d", i)
		e = e.Next()
	}
	require.Nil(t, e, "more elements than expected")

	// Confirm result iterating from back
	fromBack := slices.Reversed(fromFront)
	e = l.Back()
	for i, expected := range fromBack {
		require.NotNil(t, e, "element %d is nil", i)
		require.Equal(t, expected, e.Value, "unexpected value for %d", i)
		e = e.Prev()
	}
	require.Nil(t, e, "more elements than expected")
}
