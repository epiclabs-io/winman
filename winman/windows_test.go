package winman_test

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/gdamore/tcell"
	"gitlab.com/tslocum/cview"
	"gitlab.com/tslocum/cview/winman"
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
	*cview.Box
	Symbol     rune
	clickCount int
}

func NewBoringPrimitive(Symbol rune) *BoringPrimitive {
	return &BoringPrimitive{
		Box:    cview.NewBox(),
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

func (bp *BoringPrimitive) MouseHandler() func(action cview.MouseAction, event *tcell.EventMouse, setFocus func(p cview.Primitive)) (consumed bool, capture cview.Primitive) {
	return func(action cview.MouseAction, event *tcell.EventMouse, setFocus func(p cview.Primitive)) (consumed bool, capture cview.Primitive) {
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
	wnd          *winman.Window
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
			//var focusedPrimitive cview.Primitive
			setFocus := func(p cview.Primitive) {
				//focusedPrimitive = p
			}

			for i := 0; i < wt.wnd.ButtonCount(); i++ {
				func(i int) {
					button := wt.wnd.GetButton(i)
					button.ClickHandler = func() {
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
					//focusedPrimitive = nil
					event := tcell.NewEventMouse(x, y, tcell.Button1, tcell.ModNone)
					windowMouseHandler(cview.MouseLeftClick, event, setFocus)
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

		})
	}
}

func TestWindowManager(t *testing.T) {
	wm := winman.NewWindowManager()
	wnd1 := winman.NewWindow()
	r := wnd1.GetRoot()
	if r != nil {
		t.Fatalf("Expected to get root=nil on newly instantiated Window, got %v", r)
	}

	wnd1.SetRoot(priv)

	r = wnd1.GetRoot()
	if r != priv {
		t.Fatalf("Expected to get the same pritive, got %v", r)
	}

	// The following methods should panic when there is no WindowManager assigned to this window:

	assertPanic(t, func() {
		wnd1.Show()
	})
	assertPanic(t, func() {
		wnd1.Hide()
	})
	assertPanic(t, func() {
		wnd1.Maximize()
	})
	assertPanic(t, func() {
		wnd1.ShowModal()
	})
	assertPanic(t, func() {
		wnd1.Center()
	})

	screen := tcell.NewSimulationScreen("UTF-8")
	screen.SetSize(80, 24)
	screen.Init()
	//sm := &ScreenMonitor{screen: screen}

	windowCount := wm.WindowCount()
	if windowCount != 0 {
		t.Fatalf("Expected Window Manager to have 0 windows when initialized, got %d", windowCount)
	}

	wm.Show(wnd1) // show window in window manager
	windowCount = wm.WindowCount()

	if windowCount != 1 {
		t.Fatalf("Expected Window Manager to have 1 window after adding 1 window, got %d", windowCount)
	}

	wm.Show(wnd1) // show the same window
	windowCount = wm.WindowCount()
	if windowCount != 1 {
		t.Fatalf("Expected Window Manager to still have 1 window after adding the same window, got %d", windowCount)
	}

	wm.Hide(wnd1)
	windowCount = wm.WindowCount()
	if windowCount != 0 {
		t.Fatalf("Expected Window Manager to have no windows after hiding the only window, got %d", windowCount)
	}

}
