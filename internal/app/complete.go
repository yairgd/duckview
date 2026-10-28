package app

import (
	"strings"

	"github.com/yairgd/termforge"
)

// Tab with several candidates. termforge only completes a unique match and
// posts a CompletionMsg nobody draws, so ":filter <Tab>" with many columns did
// nothing. Here the first Tab extends the text to the candidates' common prefix,
// further Tabs cycle through them, and a single row above the command line (like vim's wildmenu) lists them.

func (a *App) initCompletionPopup() {
	a.popup = termforge.NewCompletionBarWidget(a.ctx)
	// Wildmenu row: overlays the separator right above the command line.
	a.AddFloatingWidget(a.popup, func(c termforge.Canvas) termforge.Rect {
		// Layout runs rarely, so the rect cannot depend on the candidates: it is
		// always the row above the command line; the bar paints nothing when empty.
		if c.H() < 3 {
			return c.ChildRect(0, 0, 0, 0)
		}
		return c.ChildRect(0, c.H()-2, c.W(), 1)
	})
}

func (a *App) clearCompletion() {
	a.compList = nil
	if a.popup != nil {
		a.popup.Clear()
	}
}

// completeTab handles Tab on the command line. It returns false when the
// command line should handle the key itself (zero or one candidate).
func (a *App) completeTab() bool {
	text := a.cmdWidget.Text()
	if a.compList != nil && text == a.compText {
		a.compIdx = (a.compIdx + 1) % len(a.compList)
		a.apply(a.compList[a.compIdx])
		return true
	}
	a.clearCompletion()
	names := a.cmdWidget.CompletionNames()
	if len(names) < 2 {
		return false
	}
	a.compList, a.compIdx = names, -1
	if cp := commonPrefix(names); len(cp) > len(lastChunk(text)) && strings.HasPrefix(strings.ToLower(cp), strings.ToLower(lastChunk(text))) {
		a.apply(cp)
		return true
	}
	a.compIdx = 0
	a.apply(names[0])
	return true
}

// afterTab runs after the command line itself handled Tab (a unique command was
// completed): if that command takes arguments, list them in the row so the next
// Tab starts cycling, e.g. ":filter<Tab>" shows the columns.
func (a *App) afterTab() {
	names := a.cmdWidget.CompletionNames()
	if len(names) < 2 {
		return
	}
	a.compList, a.compIdx, a.compText = names, -1, a.cmdWidget.Text()
	a.popup.SetItems(names, 0)
}

func (a *App) apply(name string) {
	a.cmdWidget.ApplyCompletion(name)
	a.compText = a.cmdWidget.Text()
	a.popup.SetItems(a.compList, a.compIdx)
}

// lastChunk is the text the command line replaces: after the last space, or
// after the ':' when there is no space.
func lastChunk(text string) string {
	i := strings.LastIndex(text, " ") + 1
	if i == 0 {
		i = 1
	}
	if i > len(text) {
		return ""
	}
	return text[i:]
}

// commonPrefix is the longest case-insensitive common prefix of names, in the
// spelling of the first one.
func commonPrefix(names []string) string {
	if len(names) == 0 {
		return ""
	}
	n := len(names[0])
	for _, s := range names[1:] {
		i := 0
		for i < n && i < len(s) && strings.EqualFold(names[0][i:i+1], s[i:i+1]) {
			i++
		}
		n = i
	}
	return names[0][:n]
}
