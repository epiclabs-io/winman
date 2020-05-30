package winman_test

import (
	"strings"
	"testing"

	"github.com/epiclabs-io/winman"
)

func dump(s winman.Stack) string {
	var b strings.Builder
	for _, e := range s {
		b.WriteRune(e.(rune))
	}
	return b.String()
}

func checkDump(t *testing.T, s winman.Stack, expected string) {
	actual := dump(s)
	if actual != expected {
		t.Fatalf("Expected stack to contain %q, got %q", expected, actual)
	}
}

func TestStack(t *testing.T) {
	var s winman.Stack

	e := s.Pop()
	if e != nil {
		t.Fatalf("Expected .Pop to return nil when the stack is empty, got %v", e)
	}

	// add an element
	s.Push('A')

	if len(s) != 1 {
		t.Fatalf("Expected the length of the stack to be 1 after adding the first element, got %d", len(s))
	}

	// add the same element
	s.Push('A')

	if len(s) != 1 {
		t.Fatalf("Expected the length of the stack to be 1 after adding the same element, got %d", len(s))
	}

	// expect panic if a nil item is added.
	assertPanic(t, func() {
		s.Push(nil)
	})

	s.Push('B')
	if len(s) != 2 {
		t.Fatalf("Expected the length of the stack to be 2 after adding a unique element, got %d", len(s))
	}

	e = s.Pop()
	if e != 'B' {
		t.Fatalf("Expected Pop to return the last pushed element, got %v", e)
	}

	if len(s) != 1 {
		t.Fatalf("Expected the length of the stack to be 1 after removing the second element, got %d", len(s))
	}

	s = nil // clear the stack

	items := []rune{'A', 'B', 'C', 'D', 'E', 'F', 'G'}
	// add a bunch of items
	for _, e := range items {
		s.Push(e)
	}

	checkDump(t, s, "ABCDEFG")

	// remove the first item
	s.Remove('A')
	checkDump(t, s, "BCDEFG")

	// remove the last item
	s.Remove('G')
	checkDump(t, s, "BCDEF")

	// remove some intermediate item
	s.Remove('E')
	checkDump(t, s, "BCDF")

	// move first item to the last position
	s.Move('B', -1)
	checkDump(t, s, "CDFB")

	// move last item to first position
	s.Move('B', 0)
	checkDump(t, s, "BCDF")

	// move intermediate item to first position
	s.Move('D', 0)
	checkDump(t, s, "DBCF")

	// move intermediate item to last position
	s.Move('C', 10000)
	checkDump(t, s, "DBFC")

	// test find
	found := s.Find(func(item interface{}) bool {
		r := item.(rune)
		return r > 'C'
	})

	if found.(rune) != 'F' {
		t.Fatalf("Expected Find to find 'F', found %c", found.(rune))
	}

	indices := map[rune]int{
		'D': 0,
		'B': 1,
		'F': 2,
		'C': 3,
	}

	for r, i := range indices {
		index := s.IndexOf(r)
		if index != i {
			t.Fatalf("Expected to find %c at index %d, got %d", r, i, index)
		}
	}
}
