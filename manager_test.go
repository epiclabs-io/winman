package winman_test

import (
	"fmt"
	"testing"

	"github.com/epiclabs-io/winman"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Focuser returns a setFocus function that remembers the primitive
// that had focus before
func Focuser(focusedPrimitive *tview.Primitive) func(tview.Primitive) {
	var setFocus func(tview.Primitive)
	setFocus = func(p tview.Primitive) {

		if *focusedPrimitive != nil {
			(*focusedPrimitive).Blur()
		}
		*focusedPrimitive = p
		if p != nil {
			p.Focus(func(p tview.Primitive) {
				setFocus(p)
			})
		}
	}
	return setFocus
}

func TestWindowManagerFocus(t *testing.T) {
	var focusedPrimitive tview.Primitive
	setFocus := Focuser(&focusedPrimitive)

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

	// Test add and remove
	wm.AddWindow(wndA)
	windowCount = wm.WindowCount()

	if windowCount != 1 {
		t.Fatalf("Expected Window Manager to have 1 window after adding 1 window, got %d", windowCount)
	}

	wm.AddWindow(wndA) // add the same window
	windowCount = wm.WindowCount()
	if windowCount != 1 {
		t.Fatalf("Expected Window Manager to still have 1 window after adding the same window, got %d", windowCount)
	}

	wm.RemoveWindow(wndA)
	windowCount = wm.WindowCount()
	if windowCount != 0 {
		t.Fatalf("Expected Window Manager to have no windows after hiding the only window, got %d", windowCount)
	}

	// Test Z index get/set
	wm.AddWindow(wndA) // add wndA again, should get z index 0
	wndB := winman.NewWindow().SetRoot(rootB)
	wm.AddWindow(wndB) // add wndB, should get z index 1 since it was added later

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

	wX := wm.Window(-1)
	if wX != nil {
		t.Fatalf("Expected no window be returned with a negative index, got %v", wX)
	}

	wX = wm.Window(1000)
	if wX != nil {
		t.Fatalf("Expected no window be returned with an out of bounds index, got %v", wX)
	}
	//Test Focus

	// Since we did not give focus to any window, then the
	// Window Manager must not have focus:

	hasFocus := wm.HasFocus()
	if hasFocus {
		t.Fatal("Expected Window Manager to not have focus")
	}

	//Focusing on the window manager must forward the focus to the window
	// with highest z index that is visible, however both windows are hidden

	setFocus(wm)

	if wndA.HasFocus() || wndB.HasFocus() {
		t.Fatal("Expected no window to get focus, since both are hidden")
	}

	wndC := wm.NewWindow() // add a third window which now will have highest z index, but it is hidden

	// now show wndB and try again
	wndB.Show()
	setFocus(wm)
	if !wndB.HasFocus() {
		t.Fatal("Expected wndB to have focus")
	}

	// now show wndC and have wm choose it as it is got higher Z
	wndC.Show()
	setFocus(wm)
	if !wndC.HasFocus() {
		t.Fatal("Expected wndC to have focus")
	}
	if wndB.HasFocus() {
		t.Fatal("Expected wndB to not have focus")
	}

	// Since wndC has focus, then the Window Manager must have focus
	hasFocus = wm.HasFocus()
	if !hasFocus {
		t.Fatal("Expected Window Manager to have focus")
	}

	// if we hide wndC, then wndB should get the focus again
	wndC.Hide()
	setFocus(wm)
	if wndC.HasFocus() {
		t.Fatal("Expected wndC to not have focus")
	}
	if !wndB.HasFocus() {
		t.Fatal("Expected wndB to  have focus")
	}

}

func TestWindowManagerSetZ(t *testing.T) {
	wm := winman.NewWindowManager()

	var windows []*winman.WindowBase
	// add some windows
	for i := 0; i < 10; i++ {
		wnd := winman.NewWindow()
		wm.AddWindow(wnd)
		windows = append(windows, wnd)
	}

	// Check that the z index of each window equals the window id
	// given in the test
	for i, wnd := range windows {
		z := wm.GetZ(wnd)
		if z != i {
			t.Fatalf("Expected window with id %d to have z index %d, got %d", i, i, z)
		}
	}

	// Move Window 0 to the top:
	wm.SetZ(windows[0], winman.WindowZTop)
	// windows[0] must now have z index of the top (len(w)-1)
	z := wm.GetZ(windows[0])
	if z != wm.WindowCount()-1 {
		t.Fatalf("Expected windows[0] to have z index %d, got %d", wm.WindowCount()-1, z)
	}

	// Move Window 3 to index 7:
	wm.SetZ(windows[3], 7)
	// windows[3] must now have z index 7
	z = wm.GetZ(windows[3])
	if z != 7 {
		t.Fatalf("Expected windows[3] to have z index 7, got %d", z)
	}

	// Move Window 5 to the bottom:
	wm.SetZ(windows[5], winman.WindowZBottom)
	// windows[5] must now have z index of the bottom (0)
	z = wm.GetZ(windows[5])
	if z != 0 {
		t.Fatalf("Expected windows[5] to have z index 0, got %d", z)
	}

	// the final z indices for all windows should be:
	indices := []int{9, 1, 2, 7, 3, 0, 4, 5, 6, 8}

	verifyIndices := func() {
		for i, wnd := range windows {
			z = wm.GetZ(wnd)
			if z != indices[i] {
				t.Fatalf("Expected windows[%d]'s z index to be %d, got %d", i, indices[i], z)
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

type Rect = winman.Rect

type WMDrawTest struct {
	initial   Rect // initial Rect
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
	var windows []*winman.WindowBase
	for _, wt := range WMDrawTests {
		wnd := winman.NewWindow() // create a new window. Windows are not visible by default.
		wm.AddWindow(wnd)         // add new window to wm.
		wnd.SetRect(wt.initial.Rect())
		if wt.maximized {
			wnd.Maximize()
		}
		windows = append(windows, wnd)
	}

	// Add an additional window to test that
	// giving it focus will move it to the top.
	var testWindow winman.Window = wm.NewWindow()

	wm.SetZ(testWindow, 3)
	z := wm.GetZ(testWindow)
	if z != 3 {
		t.Fatalf("Expected testWindow to have z 3, got %d", z)
	}

	// actually set the focus to this window
	// This will also make it visible.
	testWindow.Focus(delegate)

	screen := tcell.NewSimulationScreen("UTF-8")
	screenW, screenH := 20, 20
	screen.SetSize(screenH, screenW)
	screen.Init()
	wm.SetRect(0, 0, screenW, screenH)

	// Draw
	wm.Draw(screen)

	// check that the focused window got the highest Z after drawing
	z = wm.GetZ(testWindow)
	if z != wm.WindowCount()-1 {
		t.Fatalf("Expected the focused window to have highest z %d, got %d", wm.WindowCount()-1, z)
	}

	// remove the test Window
	wm.RemoveWindow(testWindow)

	// check that only the WMDrawTests windows remain
	expectedCount := len(windows)
	actualCount := wm.WindowCount()
	if actualCount != expectedCount {
		t.Fatalf("Expected window manager to only contain %d windows, got %d", expectedCount, actualCount)
	}

	// since WMDrawTests are not visible, their Rects should be the original ones.
	// Draw() must not adjust hidden windows.
	for i, wnd := range windows {
		rect := winman.NewRect(wnd.GetRect())
		expectedRect := WMDrawTests[i].initial
		if rect != expectedRect {
			t.Fatalf("Expected window in test %d to have rect %s, got %s", i, expectedRect, rect)
		}
	}

	// now make all windows visible:
	for _, wnd := range windows {
		wnd.Show()
	}

	// Now draw again and check if window rects were adjusted to fit:
	wm.Draw(screen)
	for i, wnd := range windows {
		rect := winman.NewRect(wnd.GetRect())
		expectedRect := WMDrawTests[i].expected
		if rect != expectedRect {
			t.Fatalf("Expected window in test %d to have rect %s, got %s", i, expectedRect, rect)
		}
	}
}

type TestMouseWindow struct {
	initial   Rect
	final     Rect
	visible   bool
	border    bool
	draggable bool
	resizable bool
}

type ClickTest struct {
	pos         Position
	action      tview.MouseAction
	expectedWnd int // expected index in testWindowRects of the window where interactions should be forwarded to
	expectFocus bool
}

func TestWindowManagerMouse(t *testing.T) {
	wm := winman.NewWindowManager()
	wm.SetRect(0, 0, 20, 20)
	screen := tcell.NewSimulationScreen("UTF-8")
	screen.SetSize(20, 20)
	screen.Init()

	testWindowRects := []TestMouseWindow{
		{Rect{2, 2, 5, 5}, Rect{0, 2, 20, 18}, true, true, true, true},      // window 0, behind window 1
		{Rect{0, 0, 12, 12}, Rect{0, 0, 14, 17}, true, true, true, true},    // window 1, behind window 2, overlapping in the bottom right
		{Rect{8, 8, 12, 12}, Rect{8, 8, 12, 12}, true, false, false, false}, // window 2, on top, overlapping on the top left with window 1
		{Rect{0, 15, 5, 5}, Rect{0, 15, 5, 5}, false, false, false, false},  //window 3, hidden, should not receive events
	}

	testClicks := []ClickTest{
		{Position{2, 2}, tview.MouseMove, 1, false},        // see if #1 receives mouse
		{Position{0, 19}, tview.MouseMove, -1, false},      // see if no window receives mouse. This is over #3, but it is hidden
		{Position{19, 0}, tview.MouseMove, -1, false},      // see if now window receives mouse. No windows here.
		{Position{10, 10}, tview.MouseMove, 2, false},      // click the overlap area of #1 and #2. #2 should receive it since it is on top
		{Position{2, 2}, tview.MouseLeftClick, 1, true},    // click on window #1, should go to the top and get focus
		{Position{10, 10}, tview.MouseMove, 1, true},       // clicking the overlap area now should go to  #1
		{Position{3, 0}, tview.MouseLeftDown, -1, true},    // begin drag on window #1
		{Position{8, 0}, tview.MouseMove, -1, true},        // move window #1 to the right
		{Position{8, 0}, tview.MouseLeftUp, 1, true},       // finish dragging #1
		{Position{2, 2}, tview.MouseLeftClick, 0, true},    // click now goes to window #0, which was behind #1 before dragging
		{Position{18, 18}, tview.MouseLeftClick, 2, true},  //click to #2, should get focus
		{Position{10, 8}, tview.MouseLeftDown, 2, true},    // begin drag on window #2
		{Position{0, 8}, tview.MouseMove, -1, false},       // move window #2 to the left. However, #2 is not draggable so this does nothing
		{Position{0, 8}, tview.MouseLeftUp, -1, false},     // finish drag operation, however we didn't really drag
		{Position{0, 17}, tview.MouseLeftClick, -1, false}, // see if no window receives mouse. This is over #3, but it is hidden. If #2 was dragged, then it would receive the click instead
		{Position{2, 2}, tview.MouseLeftClick, 0, true},    // click #0 to have it get focus.
		{Position{2, 4}, tview.MouseLeftDown, -1, true},    // begin resize on left border of window #0
		{Position{0, 4}, tview.MouseMove, -1, true},        // drag left border all the way to the left
		{Position{0, 4}, tview.MouseLeftUp, 0, false},      // finish dragging left border
		{Position{6, 4}, tview.MouseLeftDown, -1, true},    // begin resize on right border of window #0
		{Position{19, 4}, tview.MouseMove, -1, true},       // drag right border all the way to the right
		{Position{19, 4}, tview.MouseLeftUp, 0, false},     // finish dragging right border
		{Position{4, 6}, tview.MouseLeftDown, -1, true},    // begin resize on bottom border of window #0
		{Position{4, 19}, tview.MouseMove, -1, true},       // drag bottom border all the way to the bottom
		{Position{4, 19}, tview.MouseLeftUp, 0, false},     // finish dragging bottom border
		{Position{6, 1}, tview.MouseLeftClick, 1, true},    // click #1 to have it get focus.
		{Position{5, 11}, tview.MouseLeftDown, -1, true},   // begin resize on bottom left border of window #1
		{Position{0, 19}, tview.MouseMove, -1, true},       // drag bottom left corner all the way to the bottom left corner of the screen
		{Position{0, 19}, tview.MouseLeftUp, 1, false},     // finish dragging bottom border
		{Position{16, 19}, tview.MouseLeftDown, -1, true},  // begin resize on bottom right border of window #1
		{Position{13, 16}, tview.MouseMove, -1, true},      // drag bottom right corner 3 units up and 3 units left
		{Position{13, 16}, tview.MouseLeftUp, 1, false},    // finish dragging bottom right corner of #1
	}

	var clickedId int
	var windows []*winman.WindowBase
	var focusedPrimitive tview.Primitive
	setFocus := Focuser(&focusedPrimitive)

	for id, tw := range testWindowRects {
		wnd := wm.NewWindow()
		wnd.SetRect(tw.initial.Rect())
		if tw.visible {
			wnd.Show()
		}
		wnd.Draggable = tw.draggable
		wnd.Resizable = tw.resizable
		wnd.SetBorder(tw.border)
		func(id int) {
			wnd.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
				clickedId = id
				return action, event
			})
		}(id)
		windows = append(windows, wnd)
	}

	handler := wm.MouseHandler()
	for testId, click := range testClicks {
		wm.Draw(screen)
		clickedId = -1
		handler(click.action, tcell.NewEventMouse(click.pos.x, click.pos.y, tcell.Button1, tcell.ModNone), setFocus)

		if clickedId != click.expectedWnd {
			var expectedStr string
			if click.expectedWnd == -1 {
				expectedStr = "no window"
			} else {
				expectedStr = fmt.Sprintf("window with id %d", click.expectedWnd)
			}
			t.Fatalf("click test #%d: Expected %s to receive event, got %d", testId, expectedStr, clickedId)
		}

		if clickedId == -1 {
			continue
		}

		wnd := windows[clickedId]
		if wnd == nil {
			t.Fatalf("click test #%d: cannot find window with id %d", testId, clickedId)
		}

		if click.expectFocus && !wnd.HasFocus() {
			t.Fatalf("click test #%d: Expected window with id %d to have focus as a result of the click", testId, clickedId)
		}
	}

	for i, wnd := range windows {
		finalRect := winman.NewRect(wnd.GetRect())
		if finalRect != testWindowRects[i].final {
			t.Fatalf("Expected window #%d to have a final rect of %s, got %s", i, testWindowRects[i].final, finalRect)
		}
	}
}

type KeyTestPrimitive struct {
	tview.Box
}

var lastPrimitive *KeyTestPrimitive

func (ktp *KeyTestPrimitive) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		lastPrimitive = ktp
	}
}

func TestWindowManagerKeyboard(t *testing.T) {
	wm := winman.NewWindowManager()
	wm.SetRect(0, 0, 20, 20)
	var windows []*winman.WindowBase
	// add a few windows
	// windows start all hidden
	for i := 0; i < 5; i++ {
		window := wm.NewWindow()
		window.SetRect(0, 0, 20, 20)
		root := &KeyTestPrimitive{}
		window.SetRoot(root)
		windows = append(windows, window)
	}

	inputHandler := wm.InputHandler()
	if inputHandler == nil {
		t.Fatal("Window manager must return an input handler")
	}

	testEventKey := tcell.NewEventKey(tcell.KeyF1, 'a', tcell.ModCtrl)
	testSetFocus := func(tview.Primitive) {}

	// send first keypress. All windows are hidden, so no primitive should get the key
	inputHandler(testEventKey, testSetFocus)

	if lastPrimitive != nil {
		t.Fatal("No primitive should have got the key press since all windows are hidden")
	}

	// Show window #3. It should get the next keypress.
	wnd3 := windows[3]
	wnd3.Show()
	wnd3.GetRoot().Focus(nil)
	inputHandler(testEventKey, testSetFocus)
	if lastPrimitive != wnd3.GetRoot() {
		t.Fatal("Expected last keypress to have gone to window3's root")
	}

	// Show Window #4, which has a higher Z than #3, should now get the keypress
	wnd4 := windows[4]
	wnd4.Show()
	wnd4.GetRoot().Focus(nil)
	inputHandler(testEventKey, testSetFocus)
	if lastPrimitive != wnd4.GetRoot() {
		t.Fatal("Expected last keypress to have gone to window4's root")
	}

	// Change #3's Z to be highest, so it should get the keypress
	wm.SetZ(wnd3, winman.WindowZTop)
	inputHandler(testEventKey, testSetFocus)
	if lastPrimitive != wnd3.GetRoot() {
		t.Fatal("Expected last keypress to have gone to window3's root")
	}

}

type TestCenterWindow struct {
	initial Rect
	final   Rect
}

func TestCenter(t *testing.T) {
	wm := winman.NewWindowManager()
	wm.SetRect(0, 0, 20, 20)

	testWindowRects := []TestCenterWindow{
		{Rect{2, 2, 5, 5}, Rect{7, 7, 5, 5}},
		{Rect{0, 0, 4, 12}, Rect{8, 4, 4, 12}},
		{Rect{8, 8, 12, 9}, Rect{4, 5, 12, 9}},
		{Rect{0, 15, 2, 5}, Rect{9, 7, 2, 5}},
	}

	for i, tw := range testWindowRects {
		wnd := wm.NewWindow()
		wnd.SetRect(tw.initial.Rect())
		wm.Center(wnd)
		finalRect := winman.NewRect(wnd.GetRect())
		if finalRect != tw.final {
			t.Fatalf("Expected window #%d to have a final rect of %s, got %s", i, tw.final, finalRect)
		}
	}

}

func TestMaximizeRestore(t *testing.T) {
	wm := winman.NewWindowManager()
	wm.SetRect(0, 0, 20, 20)

	screen := tcell.NewSimulationScreen("UTF-8")
	screen.SetSize(20, 20)
	screen.Init()

	initialRect := winman.NewRect(1, 2, 5, 9)

	wndA := wm.NewWindow()
	wndA.SetRect(initialRect.Rect())
	wndA.Show()
	wm.Draw(screen)

	wndA.Maximize()
	wm.Draw(screen)

	maximizedRect := winman.NewRect(wndA.GetRect())
	wmSize := winman.NewRect(wm.GetInnerRect())
	if maximizedRect != wmSize {
		t.Fatalf("Expected wndA maximized rect to be %s, got %s", wmSize, maximizedRect)
	}

	wndA.Restore()
	wm.Draw(screen)

	restoredRect := winman.NewRect(wndA.GetRect())
	if restoredRect != initialRect {
		t.Fatalf("Expected wndA restored rect to be the initial rect %s, got %s", initialRect, restoredRect)
	}

}

func TestModal(t *testing.T) {
	wm := winman.NewWindowManager()
	wm.SetRect(0, 0, 100, 100)
	screen := tcell.NewSimulationScreen("UTF-8")
	screen.SetSize(100, 100)
	screen.Init()
	handler := wm.MouseHandler()
	var focusedPrimitive tview.Primitive
	setFocus := Focuser(&focusedPrimitive)

	// define 10 windows
	clicked := make([]bool, 10)
	for i := 0; i < 10; i++ {
		wnd := wm.NewWindow()
		wnd.SetRect(i*10, 0, 10, 10)
		wnd.Show()
		func(i int) {
			wnd.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
				clicked[i] = true
				return action, event
			})
		}(i)
	}

	for i := 0; i < 10; i++ {
		handler(tview.MouseLeftClick, tcell.NewEventMouse(i*10, 0, tcell.Button1, tcell.ModNone), setFocus)
	}

	count := 0
	for i := 0; i < 10; i++ {
		if clicked[i] {
			count++
		}
	}

	if count != 10 {
		t.Fatalf("Expected each window to have been clicked, got only %d clicks", count)
	}

	// now mark window 4 as modal.
	// only window 4 should get clicks
	w4 := wm.Window(4)
	w4.(*winman.WindowBase).SetModal(true)
	setFocus(w4)
	wm.Draw(screen)

	clicked = make([]bool, 10)
	for i := 0; i < 10; i++ {
		handler(tview.MouseLeftClick, tcell.NewEventMouse(i*10, 0, tcell.Button1, tcell.ModNone), setFocus)
	}

	count = 0
	for i := 0; i < 10; i++ {
		if clicked[i] {
			count++
		}
	}

	if count != 1 || !clicked[4] {
		t.Fatalf("Expected only window 4 to have been clicked. Got %d clicks", count)
	}

}
