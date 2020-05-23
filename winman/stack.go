package winman

type Stack []*Window

func (ws Stack) Push(window *Window) {
	for _, wnd := range ws {
		if wnd == window {
			return
		}
	}
	ws = append(ws, window)
}

func (ws Stack) Pop() *Window {
	lenWs := len(ws)
	if lenWs == 0 {
		return nil
	}
	var wnd *Window
	wnd, ws = ws[lenWs-1], ws[:lenWs-1]
	return wnd
}

func (ws Stack) Remove(window *Window) {
	for i, wnd := range ws {
		if wnd == window {
			ws = append(ws[:i], ws[i+1:]...)
			break
		}
	}
}

func (ws Stack) Index(window *Window) int {
	for i, wnd := range ws {
		if wnd == window {
			return i
		}
	}
	return -1
}

func (ws Stack) Move(window *Window, targetIndex int) {
	oldIndex := ws.Index(window)
	lenW := len(ws)
	if oldIndex == -1 {
		return
	}

	if targetIndex < 0 || targetIndex >= lenW {
		targetIndex = lenW - 1
	}

	newWindows := make([]*Window, lenW)
	for i, j := 0, 0; i < lenW; j++ {
		if j == oldIndex {
			j++
		}
		if i == targetIndex {
			j--
		} else {
			newWindows[i] = ws[j]
		}
		i++
	}

	newWindows[targetIndex] = window
	ws = newWindows
}

func (ws Stack) Find(f func(window *Window) bool) *Window {
	for i := len(ws) - 1; i >= 0; i-- {
		wnd := ws[i]
		if f(wnd) {
			return wnd
		}
	}
	return nil
}

func (ws Stack) Modal() *Window {
	return ws.Find(func(wnd *Window) bool {
		return wnd.modal
	})
}
