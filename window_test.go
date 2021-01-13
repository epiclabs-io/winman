package winman_test

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/epiclabs-io/winman"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ScreenMonitor struct {
	screen   tcell.SimulationScreen
	contents []tcell.SimCell
	width    int
	height   int
}

func assertPanic(t *testing.T, f func()) {
	defer func() {
		if recover() == nil {
			t.Errorf("Expected this call to panic")
		}
	}()
	f()
}

func (sm *ScreenMonitor) Sync() {
	sm.screen.Sync()
	sm.contents, sm.width, sm.height = sm.screen.GetContents()
}

func (sm *ScreenMonitor) Char(x, y int) string {
	return string(sm.contents[x+y*sm.width].Runes)
}

func (sm *ScreenMonitor) Line(x, y, width int) string {
	b := strings.Builder{}
	s := x + y*sm.width
	for i := 0; i < width; i++ {
		b.WriteString(string(sm.contents[i+s].Runes))
	}
	return b.String()
}

type BoringPrimitive struct {
	*tview.Box
	Symbol     rune
	clickCount int
}

func NewBoringPrimitive(Symbol rune) *BoringPrimitive {
	return &BoringPrimitive{
		Box:    tview.NewBox(),
		Symbol: Symbol,
	}
}

func (bp *BoringPrimitive) Draw(screen tcell.Screen) {
	px, py, width, height := bp.GetRect()
	for x := px; x < px+width; x++ {
		for y := py; y < py+height; y++ {
			screen.SetContent(x, y, bp.Symbol, nil, tcell.StyleDefault)
		}
	}
}

func (bp *BoringPrimitive) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		x, y := event.Position()
		if bp.InRect(x, y) {
			bp.clickCount++
			return true, nil
		}
		return false, nil
	}
}

type Position struct {
	x int
	y int
}

type WindowTest struct {
	wnd          *winman.WindowBase
	buttonClicks []Position
	lines        []string
}

var priv = NewBoringPrimitive('@')
var wtests = []WindowTest{
	{winman.NewWindow().SetRoot(priv), nil, []string{` ┌─────────────┐ `, ` │@@@@@@@@@@@@@│ `}},
	{winman.NewWindow().SetRoot(priv).AddButton(&winman.Button{
		Symbol:    'A',
		Alignment: winman.ButtonLeft,
	}), []Position{{3, 0}}, []string{` ┌[A]──────────┐ `, ` │@@@@@@@@@@@@@│ `}},
	{winman.NewWindow().SetRoot(priv).AddButton(&winman.Button{
		Symbol:    'B',
		Alignment: winman.ButtonRight,
	}), []Position{{13, 0}}, []string{` ┌──────────[B]┐ `, ` │@@@@@@@@@@@@@│ `}},
	{winman.NewWindow().SetRoot(priv).AddButton(&winman.Button{
		Symbol:    'C',
		Alignment: winman.ButtonRight,
	}).AddButton(&winman.Button{
		Symbol:    'D',
		Alignment: winman.ButtonLeft,
	}).AddButton(&winman.Button{
		Symbol:    'E',
		Alignment: winman.ButtonRight,
	}).AddButton(&winman.Button{
		Symbol:    'F',
		Alignment: winman.ButtonLeft,
	}), []Position{{13, 0}, {3, 0}, {10, 0}, {6, 0}}, []string{` ┌[D][F]─[E][C]┐ `, ` │@@@@@@@@@@@@@│ `}},
}

func TestWindow(t *testing.T) {

	screen := tcell.NewSimulationScreen("UTF-8")
	screen.SetSize(80, 24)
	screen.Init()
	sm := &ScreenMonitor{screen: screen}

	for num, wt := range wtests {
		t.Run(fmt.Sprintf("Window %d", num), func(t *testing.T) {
			// Start with a new screen
			screen.Clear()

			//Position window
			wt.wnd.SetRect(1, 0, 15, 10)
			//Draw this window on the screen
			wt.wnd.Draw(screen)
			sm.Sync() // sync so the screen is readable now

			// 1.- Check whether the first lines of the window have rendered correctly:
			for y, expectedLine := range wt.lines {
				line := sm.Line(0, y, utf8.RuneCountInString(expectedLine))
				if line != expectedLine {
					t.Fatalf("Wrong rendering. Expected %q, got %q", expectedLine, line)
				}
			}

			// 2.- Test that GetButton() returns nil for out of range button numbers:
			btn := wt.wnd.GetButton(-1)
			if btn != nil {
				t.Fatalf("Expected GetButton() to return nil, got %v", btn)
			}
			btn = wt.wnd.GetButton(1000)
			if btn != nil {
				t.Fatalf("Expected GetButton() to return nil, got %v", btn)
			}

			// 3.- Set up window button handlers and click all cells in the screen to see
			// if handlers are called.
			var clickedButton int
			clickCounter := make(map[int]int)
			//var focusedPrimitive tview.Primitive
			setFocus := func(p tview.Primitive) {
				//focusedPrimitive = p
			}

			for i := 0; i < wt.wnd.ButtonCount(); i++ {
				func(i int) {
					button := wt.wnd.GetButton(i)
					button.OnClick = func() {
						clickedButton = i
						clickCounter[i] = clickCounter[i] + 1
					}
				}(i)
			}
			windowMouseHandler := wt.wnd.MouseHandler()
			// The following code virtually clicks each cell of the entire screen
			// As a consequence, each window button should be clicked exactly once.
			priv.clickCount = 0

			sw, sh := screen.Size()
			for x := 0; x < sw; x++ {
				for y := 0; y < sh; y++ {
					clickedButton = -1
					event := tcell.NewEventMouse(x, y, tcell.Button1, tcell.ModNone)
					windowMouseHandler(tview.MouseLeftClick, event, setFocus)
					if clickedButton != -1 {
						expectedPos := wt.buttonClicks[clickedButton]
						if x != expectedPos.x || y != expectedPos.y {
							t.Fatalf("Expected window button to handle click to (%d,%d), got (%d,%d)", expectedPos.x, expectedPos.y, x, y)
						}
					}
				}
			}
			if len(clickCounter) != wt.wnd.ButtonCount() {
				t.Fatalf("Expected only %d different buttons to be clicked, got %d", wt.wnd.ButtonCount(), len(clickCounter))
			}
			for i, clicks := range clickCounter {
				if clicks != 1 {
					t.Fatalf("Expected each button to be clicked exactly once. Got %d clicks in button %d", clicks, i)
				}
			}

			_, _, width, height := priv.GetRect()
			if priv.clickCount != width*height {
				t.Fatalf("Expected root primitive to have received exactly %d clicks, got %d", width*height, priv.clickCount)
			}

			wt.wnd.SetBorder(false)
			for x := 0; x < sw; x++ {
				for y := 0; y < sh; y++ {
					clickedButton = -1
					event := tcell.NewEventMouse(x, y, tcell.Button1, tcell.ModNone)
					windowMouseHandler(tview.MouseLeftClick, event, setFocus)
					if clickedButton != -1 {
						t.Fatalf("Expected no window button to be clicked, since border is turned off")
					}
				}
			}

		})
	}
}

func delegate(p tview.Primitive) {
	p.Focus(delegate)
}

func TestFocusDelegation(t *testing.T) {
	root := NewBoringPrimitive('%')
	wnd := winman.NewWindow()

	hasFocus := wnd.HasFocus()
	if hasFocus == true {
		t.Fatal("Expected a newly created window without root to not have focus, since Focus() was not called")
	}

	// set focus and check if windows shows as focused
	wnd.Focus(delegate)
	hasFocus = wnd.HasFocus()
	if hasFocus == false {
		t.Fatal("Expected to have focus")
	}

	wnd.SetRoot(root)
	hasFocus = wnd.HasFocus()
	if hasFocus == true {
		t.Fatalf("Expected not to have focus, since root does not have focus")
	}

	// set focus. Focus should be passed on to root and retained.
	wnd.Focus(delegate)
	hasFocus = wnd.HasFocus()
	if hasFocus == false {
		t.Fatal("Expected window to have focus")
	}

	hasFocus = root.HasFocus()
	if hasFocus == false {
		t.Fatal("Expected root to have focus")
	}

}

func TestWindowSettings(t *testing.T) {
	wnd := winman.NewWindow()

	if wnd.IsModal() {
		t.Fatal("Expected window to be non-modal by default")
	}

	wnd.SetModal(true)

	if !wnd.IsModal() {
		t.Fatal("Expected window to be modal after setting modal to true")
	}

	if wnd.IsResizable() {
		t.Fatal("Expected window to be non-resizable by default")
	}

	wnd.SetResizable(true)

	if !wnd.IsResizable() {
		t.Fatal("Expected window to be resizable after setting resizable to true")
	}

	if wnd.IsDraggable() {
		t.Fatal("Expected window to be non-draggable by default")
	}

	wnd.SetDraggable(true)

	if !wnd.IsDraggable() {
		t.Fatal("Expected window to be draggable after setting draggable to true")
	}

	if wnd.GetTitle() != "" {
		t.Fatal("Expected window to not have a title by default")
	}

	wnd.SetTitle("Hello")

	if wnd.GetTitle() != "Hello" {
		t.Fatalf("Expected window to have the expecte title, got %s", wnd.GetTitle())
	}

}
