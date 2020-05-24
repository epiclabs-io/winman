package winman_test

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
