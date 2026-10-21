package app

import (
	"fmt"
	"strings"

	tcell "github.com/gdamore/tcell/v2"
	"github.com/yairgd/duckview/internal/db"
	"github.com/yairgd/termforge"
	"github.com/yairgd/termforge/platform"
)

const pageSize = 1000

// dataPane is a table over a db.Source. It holds one window of pageSize rows;
// moving past either edge loads the neighbouring window.
type dataPane struct {
	*termforge.TableWidget
	app   *App
	src   db.Source
	cols  []string
	types []string
	total int64
	off   int
	err   string

	onEnter func(row int) // Enter on a row (catalog: open it)
}

func (a *App) newDataPane(src db.Source) *dataPane {
	w := termforge.NewTableWidget(a.ctx)
	w.Table().SetShowHeader(true)
	w.InitSelectionKeyBindings()
	w.SetClipboard(a.ClipboardIO())
	p := &dataPane{TableWidget: w, app: a, src: src}
	w.SetRowStyleFunc(p.rowStyle)
	p.reload()
	return p
}

func (p *dataPane) rowStyle(row int) tcell.Style {
	return selStyle(row == p.SelectedRow(), p.Focused())
}

// selStyle marks the selected row: slate blue while its pane has focus, gray otherwise.
func selStyle(selected, focused bool) tcell.Style {
	if !selected {
		return tcell.StyleDefault
	}
	bg := tcell.ColorGray
	if focused {
		bg = tcell.ColorDarkSlateBlue
	}
	return tcell.StyleDefault.Bold(true).Background(bg).Foreground(platform.ContrastColor(bg))
}

// reload counts and loads the window at p.off. On failure it keeps the old
// rows and shows the error in the status line.
func (p *dataPane) reload() {
	total, err := p.app.db.Count(p.src)
	if err != nil {
		p.fail(err)
		return
	}
	p.total = total
	p.off = max(min(p.off, int(total)-1), 0)
	p.load()
}

func (p *dataPane) load() {
	cols, types, rows, err := p.app.db.Page(p.src, pageSize, p.off)
	if err != nil {
		p.fail(err)
		return
	}
	p.err, p.cols, p.types = "", cols, types
	t := p.Table()
	t.ClearColumns()
	for _, c := range cols {
		t.AddColumn(c)
	}
	for _, r := range rows {
		t.AddRow(r...)
	}
	p.SetSelectedRow(0)
	p.ResetHorizontalPan()
	p.status(len(rows))
}

func (p *dataPane) fail(err error) {
	p.err = strings.Join(strings.Fields(err.Error()), " ")
	p.status(p.Table().NumRows())
}

func (p *dataPane) status(n int) {
	s := fmt.Sprintf("%s  rows %d-%d of %d", p.src.Name, min(p.off+1, int(p.total)), p.off+n, p.total)
	if p.src.Where != "" {
		s += "  where " + p.src.Where
	}
	if p.src.Order != "" {
		s += "  order " + p.src.Order
	}
	if p.err != "" {
		s += "  ⚠ " + p.err
	}
	p.PaneName = s
}

// apply sets a new filter/sort. A bad one is rolled back and reported.
func (p *dataPane) apply(next db.Source) {
	prev := p.src
	p.src, p.off = next, 0
	p.reload()
	if p.err != "" {
		msg := p.err
		p.src = prev
		p.fail(fmt.Errorf("%s", msg))
	}
}

// goTo loads the window at off and selects row sel (-1 = last).
func (p *dataPane) goTo(off, sel int) {
	off = max(min(off, (int(p.total)-1)/pageSize*pageSize), 0)
	if off == p.off && p.Table().NumRows() > 0 {
		p.SetSelectedRow(sel)
		return
	}
	prev := p.off
	p.off = off
	p.load()
	if p.err != "" {
		p.off = prev
		return
	}
	if sel < 0 {
		sel = p.Table().NumRows() - 1
	}
	p.SetSelectedRow(sel)
	p.EnsureRowVisible()
}

func (p *dataPane) HandleFocusKey(ev *tcell.EventKey) bool {
	sel, n := p.SelectedRow(), p.Table().NumRows()
	r := ev.Rune()
	isRune := ev.Key() == tcell.KeyRune
	down := ev.Key() == tcell.KeyDown || ev.Key() == tcell.KeyPgDn || ev.Key() == tcell.KeyCtrlF || (isRune && r == 'j')
	up := ev.Key() == tcell.KeyUp || ev.Key() == tcell.KeyPgUp || ev.Key() == tcell.KeyCtrlB || (isRune && r == 'k')
	switch {
	case down && sel == n-1:
		p.goTo(p.off+pageSize, 0)
	case up && sel == 0:
		p.goTo(p.off-pageSize, -1)
	case ev.Key() == tcell.KeyHome || (isRune && r == 'g'):
		p.goTo(0, 0)
	case ev.Key() == tcell.KeyEnd || (isRune && r == 'G'):
		p.goTo(int(p.total), -1)
	case ev.Key() == tcell.KeyEnter && p.onEnter != nil && n > 0:
		p.onEnter(sel)
	default:
		return p.TableWidget.HandleFocusKey(ev)
	}
	return true
}
