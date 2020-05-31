/*

Package winman implements a basic yet powerful window manager that can be used
with tview (github.com/rivo/tview).

It supports floating windows that can be dragged, resized and maximized.
Windows can have buttons on the title bar, for example to close them,
help commands or maximize / minimize.

Windows can also be modal, meaning that other windows don't receive input while
a modal window is on top.

You can control whether the user can drag or resize windows around the screen.

Windows can overlap each other by setting their Z-index

Any tview.Primitive can be added to a window.


*/
package winman
