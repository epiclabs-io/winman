package winman_test

import (
	"testing"

	"github.com/epiclabs-io/winman"
)

func TestRect(t *testing.T) {
	r := winman.NewRect(1, 2, 3, 4)
	x, y, w, h := r.Rect()
	if !(x == 1 && y == 2 && w == 3 && h == 4) {
		t.Fatalf("Expected x=1, y=2, w=3 and h=4. Got x=%d, y=%d, h=%d, w=%d", x, y, w, h)
	}

	if r.Contains(35, 35) {
		t.Fatal("Expected the given coordinates to not be contained in r")
	}

	if !r.Contains(1, 2) {
		t.Fatal("Expected the given coordinates to be contained in r")
	}

	if !r.Contains(2, 3) {
		t.Fatal("Expected the given coordinates to be contained in r")
	}

	if !r.Contains(3, 4) {
		t.Fatal("Expected the given coordinates to be contained in r")
	}

	s := r.String()
	expectedString := "{(1, 2) 3x4}"
	if s != expectedString {
		t.Fatalf("Expected String() to return %q, got %q", expectedString, s)
	}

}
