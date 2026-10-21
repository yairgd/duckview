package app

import (
	tcell "github.com/gdamore/tcell/v2"
	"github.com/yairgd/termforge"
	"github.com/yairgd/termforge/commands"
	"github.com/yairgd/termforge/platform"
)

func (a *App) initKeyBindings() {
	a.keyBindings = commands.NewKeyBindingRegistry()
	bind := func(name string, fn func(), keys ...string) {
		a.keyBindings.Bind(commands.NewCommand(name, func(...any) { fn() }), keys...)
	}
	bind("move-left", func() { a.focus(a.Layout().FocusLeft) }, "<C-w>h", "<C-w><Left>")
	bind("move-right", func() { a.focus(a.Layout().FocusRight) }, "<C-w>l", "<C-w><Right>")
	bind("move-up", func() { a.focus(a.Layout().FocusUp) }, "<C-w>k", "<C-w><Up>")
	bind("move-down", func() { a.focus(a.Layout().FocusDown) }, "<C-w>j", "<C-w><Down>")
	bind("escape", a.leaveCommandMode, "<Esc>")
	bind("command-mode", a.enterCommandMode, ":")
	bind("quit", a.Exit, "<C-d>")
}

func (a *App) enterCommandMode() {
	if !a.cmdWidget.Active() {
		a.cmdWidget.Activate()
	}
	a.SetMode(platform.ModeCommand)
	a.RequestFrame()
}

func (a *App) leaveCommandMode() {
	a.cmdWidget.Deativate()
	a.SetMode(platform.ModeNormal)
	a.RequestFrame()
}

func (a *App) handleNormalKey(ev *tcell.EventKey) bool {
	if a.tryKeyBindings(ev) {
		return true
	}
	if h, ok := a.focusedWidget().(termforge.FocusKeyHandler); ok {
		h.HandleFocusKey(ev)
	}
	return true
}

func (a *App) handleCommandKey(ev *tcell.EventKey) bool {
	a.cmdWidget.HandleEvent(ev)
	if ev.Key() == tcell.KeyEnter {
		a.cmdWidget.Deativate()
		if a.Mode() == platform.ModeCommand {
			a.SetMode(platform.ModeNormal)
		}
	}
	return true
}

func (a *App) tryKeyBindings(ev *tcell.EventKey) bool {
	key, ok := platform.KeyFromEvent(ev)
	if !ok {
		a.keyBindings.ResetPartial()
		return false
	}
	completed, handled := a.keyBindings.HandleKey(key)
	if !handled {
		return false
	}
	return completed || a.keyBindings.InPartial()
}

func (a *App) HandleTTYResume() {}

func (a *App) HandleInterrupt(ev *tcell.EventInterrupt) {
	if msg, ok := ev.Data().(termforge.SubmitMsg); ok && a.ctx.Bus != nil {
		a.ctx.Bus.Dispatch(msg)
	}
}

func (a *App) cmdLineRect() termforge.Rect { return a.Layout().PinnedBottomRect() }

func (a *App) HandleMouse(ev *tcell.EventMouse) {
	x, y := ev.Position()
	btn := ev.Buttons()
	primary := btn&tcell.ButtonPrimary != 0
	wheel := btn&(tcell.WheelUp|tcell.WheelDown) != 0

	inCmd := a.cmdLineRect().Contains(x, y)
	if a.Mode() == platform.ModeCommand || a.Mode() == platform.ModeCompletion {
		if btn&tcell.ButtonMiddle != 0 {
			a.cmdWidget.HandleEvent(ev)
			a.RequestFrame()
			return
		}
		if primary && inCmd {
			a.cmdWidget.SetCursorAtLocalX(x - a.cmdLineRect().X())
			a.RequestFrame()
			return
		}
		if primary || wheel {
			a.leaveCommandMode()
		} else {
			return
		}
	}
	if primary && inCmd {
		a.enterCommandMode()
		a.cmdWidget.SetCursorAtLocalX(x - a.cmdLineRect().X())
		return
	}
	if primary || wheel || btn&tcell.ButtonMiddle != 0 {
		a.Layout().FocusAt(x, y)
	}
	a.tab.HandleEvent(ev)
	// A click on a catalog row is Enter on it (press edge only, not drag/repeat).
	if primary && !a.pressed {
		if _, onRow := a.cat.HitDataRow(x, y); onRow && a.focusedWidget() == termforge.Widget(a.cat) {
			a.cat.onEnter(a.cat.SelectedRow())
		}
	}
	a.pressed = primary
	a.RequestFrame()
}
