# Architecture

## Packages

| Path | Role |
|---|---|
| `cmd/duckview` | flag parsing, opens the DB, runs the app |
| `internal/db` | all SQL: `Source`, `Count`, `Page`, `Hist`, `Series`, `Stats`, `Export`, `Catalog` |
| `internal/app` | termforge wiring: `App`, `dataPane`, commands, keys, mouse |

Dependencies: duckview → termforge, `go-duckdb` (cgo, bundles DuckDB). termforge never imports
duckview.

## The `Source` idea

```go
type Source struct{ Name, Base, Where, Order string }
```

`Base` is any SELECT. A table is `select * from "s"."t"`, `:sql` is the user's query, the catalog is
a generated `union all` over `information_schema`. `Where` and `Order` are the user's `:filter` /
`:sort`. Everything else is derived:

- relation: `(select * from (Base) _t where Where)`
- page: relation + `order by Order limit 1000 offset N`
- hist / series / stats / export: the same relation, so they always match the table you see.

DuckDB does the sort, filter, binning, thinning and stats; Go only formats the page it displays.

## Panes

`dataPane` (`app/pane.go`) embeds `*termforge.TableWidget` and adds a `Source`, a 1000-row window
(`off`, `total`), error text and a hook:

```go
onEnter func(row int)   // nil for data; "open this table" for the catalog
```

- `Draw` is termforge's `TableWidget.Draw`; the row marker comes from `SetRowStyleFunc(p.rowStyle)`.
- `HandleFocusKey` intercepts only window-edge moves, `g`/`G`, and Enter; the rest goes to the widget.
- The status band text is `PaneName`, rebuilt by `status()` after every load.
- `apply()` swaps in a new `Source`, reloads, and rolls back if the SQL fails.

The catalog is a `dataPane` over the catalog query with `onEnter` set (`app/app.go`). A click or
Enter calls `openTable` → `openSource` → `newTree(newDataPane(src))`, which replaces the layout with
a fresh left pane beside the same catalog object (so its selection survives).

`:hist` uses termforge's `HistogramWidget`, `:plot` uses `GraphWidget`, `:schema`/`:help` use a plain
`TableWidget`. They are added with `Split` under the left pane and rebalanced.

## Layout

One `TabWidget` holding one `WidgetTree`: `left | catalog` (ratio 0.7), command line pinned at the
bottom (`PinBottom`). No tab bar. The wiring in `app.go`, `input.go` is the termforge demo's, trimmed:
mode handlers (normal, command), key bindings registry, `HandleMouse`, `HandleInterrupt`.

## Mouse click = Enter

`HandleMouse` lets termforge move the selection, then on the **press edge** (`a.pressed` tracks the
previous event) of a primary click on a catalog data row it calls `cat.onEnter(selected)`.

## termforge addition

`HistogramWidget` (`termforge/histogram_widget.go`): vertical bars from a `Graph` (value = height,
label = range ends), eighth-block resolution, bars widen to fill the width, extra bins merge by max.
Compact by design — no zoom/pan, no new architecture.
