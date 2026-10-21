// Package app is the duckview TUI: a data pane on the left, the catalog on the
// right, and commands that sort, filter, query, chart and export whatever the
// database holds.
package app

import (
	"github.com/yairgd/duckview/internal/db"
	"github.com/yairgd/termforge"
	"github.com/yairgd/termforge/commands"
	"github.com/yairgd/termforge/platform"
)

type App struct {
	*termforge.App
	db          *db.DB
	commandReg  *commands.CommandRegistry
	keyBindings *commands.KeyBindingRegistry

	tab       *termforge.TabWidget
	cmdWidget *termforge.CmdWidget
	ctx       platform.AppContext

	cat     *dataPane // catalog, kept across openSource so its selection survives
	pressed bool      // primary button was down on the previous mouse event
	tables  []string  // "schema.name" for ':open' completion
}

func New(d *db.DB) (*App, error) {
	a := &App{App: termforge.NewApp(), db: d, commandReg: commands.NewCommandRegistry()}
	a.App.Api = a
	if err := a.init(); err != nil {
		a.Close()
		return nil, err
	}
	return a, nil
}

func (a *App) Close() {
	if a.App != nil {
		a.App.Close()
	}
}

func (a *App) init() error {
	a.ctx = platform.NewAppContext()
	a.cmdWidget = termforge.NewCmdWidget(a.commandReg)
	a.cmdWidget.Ctx = a.ctx
	a.cmdWidget.SetPostInterrupt(a.PostInterrupt)
	a.cmdWidget.SetClipboard(a.ClipboardIO())
	a.SetCmdline(a.cmdWidget)

	base, err := a.db.Catalog()
	if err != nil {
		return err
	}
	cat := a.newDataPane(db.Source{Name: "catalog", Base: base})
	a.cat = cat
	cat.onEnter = func(row int) {
		a.openTable(cat.Table().CellText(row, 0), cat.Table().CellText(row, 1))
	}
	for i := 0; i < cat.Table().NumRows(); i++ {
		a.tables = append(a.tables, cat.Table().CellText(i, 0)+"."+cat.Table().CellText(i, 1))
	}

	empty := termforge.NewTableWidget(a.ctx)
	empty.PaneName = "data: pick a table in the catalog (Enter)"
	a.tab = termforge.NewTabWidget("duckview", a.newTree(empty))
	a.AddWidget(a.tab)

	if a.ctx.Bus != nil {
		platform.Subscribe(a.ctx.Bus, func(msg termforge.SubmitMsg) {
			if msg.CmdID == termforge.CmdExitMode {
				a.leaveCommandMode()
			}
		})
	}
	a.initKeyBindings()
	a.initCommands()
	a.RegisterModeHandler(platform.ModeNormal, a.handleNormalKey)
	a.RegisterModeHandler(platform.ModeCommand, a.handleCommandKey)
	return nil
}

// newTree is a split tree with left beside the catalog (focused), and the
// shared command line pinned below.
func (a *App) newTree(left termforge.NodeWidget) *termforge.WidgetTree {
	tree := termforge.NewWidgetTree(left)
	tree.SetStatusClipboard(a.ClipboardIO())
	tree.SetOnResize(a.RequestFrame)
	if a.cat != nil {
		tree.Split(termforge.Vertical, a.cat)
		tree.Root().Ratio = 0.7
		tree.FocusWidget(a.cat)
	}
	tree.PinBottom(a.cmdWidget, 1)
	return tree
}

// Layout is the split tree.
func (a *App) Layout() *termforge.WidgetTree {
	tree, _ := a.tab.Layout().(*termforge.WidgetTree)
	return tree
}

func (a *App) focusedWidget() termforge.Widget {
	if lay := a.Layout(); lay != nil {
		return lay.FocusedWidget()
	}
	return nil
}

// openTable shows schema.name in the data tab, creating the tab on first use.
func (a *App) openTable(schema, name string) {
	a.openSource(db.Source{Name: schema + "." + name, Base: "select * from " + db.Quote(schema) + "." + db.Quote(name)})
}

// openSource shows src in the left pane, dropping any splits made there.
func (a *App) openSource(src db.Source) {
	a.tab.SetLayout(a.newTree(a.newDataPane(src)))
	a.RequestFrame()
}
