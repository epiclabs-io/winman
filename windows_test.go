package cview_test

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/gdamore/tcell"
	"gitlab.com/tslocum/cview"
)

type ScreenMonitor struct {
	screen   tcell.SimulationScreen
	contents []tcell.SimCell
	width    int
	height   int
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
	Symbol rune
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

func TestWindowManager(t *testing.T) {
	//wm := cview.NewWindowManager()

}

type Position struct {
	x int
	y int
}
type WindowTest struct {
	wnd          *cview.Window
	buttonClicks []Position
	lines        []string
}

var priv = NewBoringPrimitive('@')
var wtests = []WindowTest{
	{cview.NewWindow().SetRoot(priv), nil, []string{` ┌─────────────┐ `, ` │@@@@@@@@@@@@@│ `}},
	{cview.NewWindow().SetRoot(priv).AddButton(&cview.WindowButton{
		Symbol:    'A',
		Alignment: cview.AlignLeft,
	}), []Position{{3, 0}}, []string{` ┌[A]──────────┐ `, ` │@@@@@@@@@@@@@│ `}},
	{cview.NewWindow().SetRoot(priv).AddButton(&cview.WindowButton{
		Symbol:    'B',
		Alignment: cview.AlignRight,
	}), []Position{{13, 0}}, []string{` ┌──────────[B]┐ `, ` │@@@@@@@@@@@@@│ `}},
	{cview.NewWindow().SetRoot(priv).AddButton(&cview.WindowButton{
		Symbol:    'C',
		Alignment: cview.AlignRight,
	}).AddButton(&cview.WindowButton{
		Symbol:    'D',
		Alignment: cview.AlignLeft,
	}).AddButton(&cview.WindowButton{
		Symbol:    'E',
		Alignment: cview.AlignRight,
	}).AddButton(&cview.WindowButton{
		Symbol:    'F',
		Alignment: cview.AlignLeft,
	}), []Position{{13, 0}, {3, 0}, {10, 0}, {6, 0}}, []string{` ┌[D][F]─[E][C]┐ `, ` │@@@@@@@@@@@@@│ `}},
}

func TestWindow(t *testing.T) {
	wnd := cview.NewWindow()
	r := wnd.GetRoot()
	if r != nil {
		t.Fatalf("Expected to get root=nil on newly instantiated Window, got %v", r)
	}

	wnd.SetRoot(priv)

	r = wnd.GetRoot()
	if r != priv {
		t.Fatalf("Expected to get the same pritive, got %v", r)
	}

	screen := tcell.NewSimulationScreen("UTF-8")
	screen.SetSize(80, 24)
	screen.Init()
	sm := &ScreenMonitor{screen: screen}

	for _, wt := range wtests {
		screen.Clear()
		wt.wnd.SetRect(1, 0, 15, 10)
		wt.wnd.Draw(screen)
		sm.Sync()
		for y, expectedLine := range wt.lines {
			line := sm.Line(0, y, utf8.RuneCountInString(expectedLine))
			if line != expectedLine {
				t.Fatalf("Wrong rendering. Expected %q, got %q", expectedLine, line)
			}
		}

		var clickedButton int
		//var focusedPrimitive cview.Primitive
		setFocus := func(p cview.Primitive) {
			//focusedPrimitive = p
		}

		for i := 0; i < wt.wnd.ButtonCount(); i++ {
			func(i int) {
				button := wt.wnd.GetButton(i)
				button.ClickHandler = func() {
					clickedButton = i
				}
			}(i)
		}
		sw, sh := screen.Size()
		mouseHandler := wt.wnd.MouseHandler()
		for x := 0; x < sw; x++ {
			for y := 0; y < sh; y++ {
				clickedButton = -1
				//focusedPrimitive = nil
				event := tcell.NewEventMouse(x, y, tcell.Button1, tcell.ModNone)
				mouseHandler(cview.MouseLeftClick, event, setFocus)
				if clickedButton != -1 {
					fmt.Println(x, y)
					expectedPos := wt.buttonClicks[clickedButton]
					if x != expectedPos.x || y != expectedPos.y {
						t.Fatalf("Expected window button to handle click to (%d,%d), got (%d,%d)", expectedPos.x, expectedPos.y, x, y)
					}
				}
			}
		}
	}
}
