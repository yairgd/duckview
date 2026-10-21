package app

import (
	tcell "github.com/gdamore/tcell/v2"
	"strconv"
	"strings"

	"github.com/yairgd/duckview/internal/db"
	"github.com/yairgd/termforge"
	"github.com/yairgd/termforge/commands"
)

func (a *App) initCommands() {
	a.commandReg.Root.
		Group("window",
			commands.Cmd("left", func(...any) { a.focus(a.Layout().FocusLeft) }),
			commands.Cmd("right", func(...any) { a.focus(a.Layout().FocusRight) }),
			commands.Cmd("up", func(...any) { a.focus(a.Layout().FocusUp) }),
			commands.Cmd("down", func(...any) { a.focus(a.Layout().FocusDown) }),
		).
		LeafRestComplete("open", a.cmdOpen, a.completeTables).
		LeafRestComplete("sort", a.cmdSort, a.completeColumns).
		LeafRestComplete("filter", a.cmdFilter, a.completeColumns).
		LeafRestComplete("hist", a.cmdHist, a.completeColumns).
		LeafRestComplete("plot", a.cmdPlot, a.completeColumns).
		LeafRest("sql", a.cmdSQL).
		LeafRest("export", a.cmdExport).
		Leaf("schema", a.cmdSchema).
		Leaf("reload", func(...any) { a.withData(func(p *dataPane) { p.reload() }) }).
		Leaf("help", a.cmdHelp).
		Leaf("close", func(...any) { a.Layout().DeleteFocus(); a.RequestFrame() }).
		Leaf("only", func(...any) { a.Layout().OnlyFocus(); a.RequestFrame() }).
		Leaf("quit", func(...any) { a.Exit() })
}

func (a *App) focus(move func()) { move(); a.RequestFrame() }

// arg is the command's whole argument text (rest args may arrive as words).
func arg(args []any) string {
	var words []string
	for _, a := range args {
		if w, ok := a.(string); ok {
			words = append(words, w)
		}
	}
	return strings.TrimSpace(strings.Join(words, " "))
}

// dataPane is the pane commands act on: the focused data pane, else the left
// one, else the catalog itself.
func (a *App) dataPane() *dataPane {
	lay := a.Layout()
	if p, ok := lay.FocusedWidget().(*dataPane); ok && p != a.cat {
		return p
	}
	n := lay.FindLeaf(func(w termforge.Widget) bool { p, ok := w.(*dataPane); return ok && p != a.cat })
	if n != nil {
		return n.Widget.(*dataPane)
	}
	return a.cat
}

func (a *App) withData(fn func(*dataPane)) {
	if p := a.dataPane(); p != nil {
		fn(p)
		a.RequestFrame()
	}
}

func (a *App) completeTables(prefix string, _ bool) []string {
	return filterPrefix(a.tables, prefix)
}

// completeColumns completes the first word against the data pane's columns.
func (a *App) completeColumns(prefix string, _ bool) []string {
	if p := a.dataPane(); p != nil && !strings.ContainsAny(prefix, " \t") {
		return filterPrefix(p.cols, prefix)
	}
	return nil
}

func filterPrefix(names []string, prefix string) []string {
	var out []string
	for _, n := range names {
		if strings.HasPrefix(strings.ToLower(n), strings.ToLower(prefix)) {
			out = append(out, n)
		}
	}
	return out
}

func (a *App) cmdOpen(args ...any) {
	name := arg(args)
	sc, tb, ok := strings.Cut(name, ".")
	if !ok {
		sc, tb = "main", name
	}
	a.openTable(sc, tb)
}

func (a *App) cmdSort(args ...any) {
	a.withData(func(p *dataPane) { n := p.src; n.Order = arg(args); p.apply(n) })
}

func (a *App) cmdFilter(args ...any) {
	a.withData(func(p *dataPane) { n := p.src; n.Where = arg(args); p.apply(n) })
}

func (a *App) cmdSQL(args ...any) {
	if q := strings.TrimRight(arg(args), "; "); q != "" {
		a.openSource(db.Source{Name: "sql", Base: q})
	}
}

// split stacks w under the left (data) pane, which it focuses first.
func (a *App) split(w termforge.NodeWidget) {
	if p := a.dataPane(); p != a.cat {
		a.Layout().FocusWidget(p)
	}
	a.Layout().Split(termforge.Horizontal, w)
	a.Layout().Rebalance()
	a.RequestFrame()
}

func (a *App) cmdHist(args ...any) {
	fields := strings.Fields(arg(args))
	if len(fields) == 0 {
		return
	}
	bins := 40
	if n, err := strconv.Atoi(fields[len(fields)-1]); err == nil && len(fields) > 1 {
		bins, fields = max(n, 2), fields[:len(fields)-1]
	}
	expr := strings.Join(fields, " ")
	a.withData(func(p *dataPane) {
		w := termforge.NewHistogramWidget(a.ctx)
		w.PaneName = "hist"
		labels, counts, err := a.db.Hist(p.src, expr, bins)
		if err != nil {
			w.Graph().SetTitle("hist " + expr + ": " + err.Error())
		} else {
			w.Graph().SetTitle("hist " + expr)
			w.Graph().SetPoints(labels, counts)
		}
		a.split(w)
	})
}

// cmdPlot: ":plot <y> [<x>]" — x defaults to the first date column.
func (a *App) cmdPlot(args ...any) {
	fields := strings.Fields(arg(args))
	if len(fields) == 0 {
		return
	}
	a.withData(func(p *dataPane) {
		y, x := fields[0], ""
		if len(fields) > 1 {
			x = strings.Join(fields[1:], " ")
		}
		for i, t := range p.types {
			if x == "" && (t == "DATE" || strings.HasPrefix(t, "TIMESTAMP")) {
				x = db.Quote(p.cols[i])
			}
		}
		w := termforge.NewGraphWidget(a.ctx)
		w.PaneName = "plot"
		w.InitGraphKeyBindings()
		labels, vals, err := a.db.Series(p.src, y, x, 4000)
		if err != nil {
			w.Graph().SetTitle("plot " + y + ": " + err.Error())
		} else {
			w.Graph().SetTitle("plot " + y)
			w.Graph().SetPoints(labels, vals)
		}
		a.split(w)
	})
}

func (a *App) cmdExport(args ...any) {
	format, path, _ := strings.Cut(arg(args), " ")
	a.withData(func(p *dataPane) {
		if err := a.db.Export(p.src, format, strings.TrimSpace(path)); err != nil {
			p.fail(err)
			return
		}
		p.err = ""
		p.status(p.Table().NumRows())
		p.PaneName += "  exported " + strings.TrimSpace(path)
	})
}

// staticTable shows fixed rows in a new pane.
func (a *App) staticTable(title string, cols []string, rows [][]string) {
	w := termforge.NewTableWidget(a.ctx)
	w.PaneName = title
	w.Table().SetShowHeader(true)
	w.SetRowStyleFunc(func(row int) tcell.Style { return selStyle(row == w.SelectedRow(), w.Focused()) })
	for _, c := range cols {
		w.Table().AddColumn(c)
	}
	for _, r := range rows {
		w.Table().AddRow(r...)
	}
	w.InitSelectionKeyBindings()
	w.SetClipboard(a.ClipboardIO())
	a.split(w)
}

func (a *App) cmdSchema(...any) {
	a.withData(func(p *dataPane) {
		rows, err := a.db.Stats(p.src, p.cols, p.types)
		if err != nil {
			p.fail(err)
			return
		}
		a.staticTable("schema "+p.src.Name, []string{"column", "type", "nulls", "distinct~", "min", "max"}, rows)
	})
}

func (a *App) cmdHelp(...any) {
	a.staticTable("help", []string{"command", "does"}, [][]string{
		{":open [schema.]table", "show a table or view in the left pane (Enter in the catalog)"},
		{":sql <select ...>", "show a query result in the left pane"},
		{":sort <expr> [desc]", "ORDER BY, pushed to SQL; no argument clears"},
		{":filter <expr>", "WHERE, pushed to SQL; no argument clears"},
		{":schema", "column types, nulls, distinct, min/max in a new pane"},
		{":hist <expr> [bins]", "histogram of a numeric expression"},
		{":plot <y> [<x>]", "line graph; x defaults to the first date column"},
		{":export csv|parquet|json <path>", "write the filtered, sorted rows"},
		{":reload  :close  :only  :quit", "refresh, close pane, keep pane, exit"},
		{"j k PgUp PgDn g G h l", "move; crossing a 1000-row window edge loads the next"},
		{"+ - h l 0 (graph)", "zoom and pan the line graph"},
		{"C-w hjkl  C-d", "pane focus, quit"},
	})
}
