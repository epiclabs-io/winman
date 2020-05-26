package winman_test

import (
	"fmt"
	"testing"

	"github.com/epiclabs-io/winman"
	"github.com/gdamore/tcell"
)

func TestWindowManager(t *testing.T) {
	wm := winman.NewWindowManager()
	wndA := winman.NewWindow()
	r := wndA.GetRoot()
	if r != nil {
		t.Fatalf("Expected to get root=nil on newly instantiated Window, got %v", r)
	}

	rootA := NewBoringPrimitive('1')
	rootB := NewBoringPrimitive('2')

	wndA.SetRoot(rootA)

	r = wndA.GetRoot()
	if r != rootA {
		t.Fatalf("Expected to get the same pritive, got %v", r)
	}

	// Test WindowCount
	windowCount := wm.WindowCount()
	if windowCount != 0 {
		t.Fatalf("Expected Window Manager to have 0 windows when initialized, got %d", windowCount)
	}

	// Test Show and Hide
	wm.Show(wndA) // show window in window manager
	windowCount = wm.WindowCount()

	if windowCount != 1 {
		t.Fatalf("Expected Window Manager to have 1 window after adding 1 window, got %d", windowCount)
	}

	wm.Show(wndA) // show the same window
	windowCount = wm.WindowCount()
	if windowCount != 1 {
		t.Fatalf("Expected Window Manager to still have 1 window after adding the same window, got %d", windowCount)
	}

	wm.Hide(wndA)
	windowCount = wm.WindowCount()
	if windowCount != 0 {
		t.Fatalf("Expected Window Manager to have no windows after hiding the only window, got %d", windowCount)
	}

	// Test Z index get/set
	wm.Show(wndA) // show wndA again, should get z index 0
	wndB := winman.NewWindow().SetRoot(rootB)
	wm.Show(wndB) // show wndB, should get z index 1 since it was shown later

	z := wm.GetZ(wndA)
	if z != 0 {
		t.Fatalf("Expected wnd1's Z index to be %d, got %d", 0, z)
	}

	z = wm.GetZ(wndB)
	if z != 1 {
		t.Fatalf("Expected wnd2's Z index to be %d, got %d", 1, z)
	}

	// Test get Window by index
	wA := wm.Window(0)
	if wA != wndA {
		t.Fatalf("Expected wndA to be at index 0, got %v", wA)
	}

	wB := wm.Window(1)
	if wB != wndB {
		t.Fatalf("Expected wndB to be at index 1, got %v", wB)
	}

	//Test Focus

	// Since we did not give focus to any window, then the
	// Window Manager must not have focus:

	hasFocus := wm.HasFocus()
	if hasFocus {
		t.Fatal("Expected Window Manager to not have focus")
	}

	//Focusing on the window manager must forward the focus to the window
	// with highest z index (wndB)
	wm.Focus(delegate)
	if !wndB.HasFocus() {
		t.Fatal("Expected wndB to have focus")
	}
	if wndA.HasFocus() {
		t.Fatal("Expected wndA to not have focus")
	}

	// Since wndB has focus, then the Window Manager must have focus
	hasFocus = wm.HasFocus()
	if !hasFocus {
		t.Fatal("Expected Window Manager to have focus")
	}
}

func TestWindowManagerSetZ(t *testing.T) {
	wm := winman.NewWindowManager()

	var w []*winman.WindowBase
	// add some windows
	for i := 0; i < 10; i++ {
		wnd := winman.NewWindow()
		wm.Show(wnd)
		w = append(w, wnd)
	}

	// Check that the z index of each window equals the window id
	// given in the test
	for i, wnd := range w {
		z := wm.GetZ(wnd)
		if z != i {
			t.Fatalf("Expected window with id %d to have z index %d, got %d", i, i, z)
		}
	}

	// Move Window 0 to the top:
	wm.SetZ(w[0], winman.WindowZTop)
	// w[0] must now have z index of the top (len(w)-1)
	z := wm.GetZ(w[0])
	if z != wm.WindowCount()-1 {
		t.Fatalf("Expected w[0] to have z index %d, got %d", wm.WindowCount()-1, z)
	}

	// Move Window 3 to index 7:
	wm.SetZ(w[3], 7)
	// w[3] must now have z index 7
	z = wm.GetZ(w[3])
	if z != 7 {
		t.Fatalf("Expected w[3] to have z index 7, got %d", z)
	}

	// Move Window 5 to the bottom:
	wm.SetZ(w[5], winman.WindowZBottom)
	// w[5] must now have z index of the bottom (0)
	z = wm.GetZ(w[5])
	if z != 0 {
		t.Fatalf("Expected w[5] to have z index 0, got %d", z)
	}

	// the final z indices for all windows should be:
	indices := []int{9, 1, 2, 7, 3, 0, 4, 5, 6, 8}

	verifyIndices := func() {
		for i, wnd := range w {
			z = wm.GetZ(wnd)
			if z != indices[i] {
				t.Fatalf("Expected w[%d]'s z index to be %d, got %d", i, indices[i], z)
			}
		}
	}

	verifyIndices()

	// set the Z for a non-added window should not change the above,
	// i.e., it should be ignored:
	wm.SetZ(winman.NewWindow(), 3)

	// check nothing changed.
	verifyIndices()
}

type Rect struct {
	x int
	y int
	w int
	h int
}

func NewRect(x, y, w, h int) Rect {
	return Rect{x, y, w, h}
}

func (r Rect) String() string {
	return fmt.Sprintf("{(%d, %d) %dx%d}", r.x, r.y, r.w, r.h)
}

type WMDrawTest struct {
	i         Rect // initial Rect
	expected  Rect // expected Rect after drawing
	maximized bool
}

var minW = winman.MinWindowWidth
var minH = winman.MinWindowHeight

var WMDrawTests = []WMDrawTest{
	{Rect{5, 5, 7, 7}, Rect{5, 5, 7, 7}, false},       // window fits
	{Rect{-3, -4, 7, 7}, Rect{0, 0, 7, 7}, false},     // overflows top left
	{Rect{37, 44, 7, 7}, Rect{13, 13, 7, 7}, false},   // overflows bottom right
	{Rect{37, 44, 73, 73}, Rect{0, 0, 20, 20}, false}, // window is too large
	{Rect{5, 5, 73, 10}, Rect{0, 5, 20, 10}, false},   // too wide and overflows to the right
	{Rect{5, 5, 10, 73}, Rect{5, 0, 10, 20}, false},   // too tall and overflows to the bottom
	{Rect{-5, 5, 73, 10}, Rect{0, 5, 20, 10}, false},  // too wide and overflows to the left
	{Rect{5, -5, 10, 73}, Rect{5, 0, 10, 20}, false},  // too tall and overflows to the top
	{Rect{5, 5, 1, 1}, Rect{5, 5, minW, minH}, false}, // window is too small
	{Rect{5, 5, 7, 7}, Rect{0, 0, 20, 20}, true},      // window is maximized

}

func TestWindowManagerDraw(t *testing.T) {
	wm := winman.NewWindowManager()
	var w []*winman.WindowBase
	for _, wt := range WMDrawTests {
		wnd := winman.NewWindow() // add new window to wm
		wm.Show(wnd)
		wnd.SetRect(wt.i.x, wt.i.y, wt.i.w, wt.i.h)
		if wt.maximized {
			wnd.Maximize()
		}
		w = append(w, wnd)
	}

	// give focus to some window
	focusedWindow := w[3]
	focusedWindow.Focus(delegate)

	screen := tcell.NewSimulationScreen("UTF-8")
	screenW, screenH := 20, 20
	screen.SetSize(screenH, screenW)
	screen.Init()
	//	sm := &ScreenMonitor{screen: screen}
	wm.SetRect(0, 0, screenW, screenH)

	// Draw
	wm.Draw(screen)

	// check that the focused window got the highest Z after drawing
	z := wm.GetZ(focusedWindow)
	if z != wm.WindowCount()-1 {
		t.Fatalf("Expected the focused window to have highes z %d, got %d", wm.WindowCount()-1, z)
	}

	for i := 0; i < wm.WindowCount(); i++ {
		rect := NewRect(w[i].GetRect())
		expectedRect := WMDrawTests[i].expected
		if rect != expectedRect {
			t.Fatalf("Expected window in test %d to have rect %s, got %s", i, expectedRect, rect)
		}
	}
}
