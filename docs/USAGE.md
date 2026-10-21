# Usage

## Layout

One screen, two panes side by side plus the `:` command line pinned at the bottom.

- **Left: data pane.** The table, view or query you opened. Status band:
  `main.prices_daily  rows 1-1000 of 235  where stock_id=500  order volume desc`.
- **Right: catalog.** Every table and view with row and column counts. Focus starts here.

`:hist`, `:plot` and `:schema` stack extra panes under the left pane; `:close` removes the
focused one, `:only` keeps just the focused one.

## Keys (normal mode)

| Key | Action |
|---|---|
| `j` `k` `↓` `↑` | move the row marker |
| `PgDn` `PgUp` `C-f` `C-b` | page |
| `g` `Home` / `G` `End` | first / last window of the table |
| `h` `l` `←` `→` | pan columns (tables can have 100+) |
| `Enter` | in the catalog: open the selected table |
| `:` | command line |
| `C-w h/j/k/l` | move focus between panes |
| `Esc` | leave command line |
| `C-d` | quit |
| `+` `-` `h` `l` `0` | in a `:plot` pane: zoom, pan, reset |

Tables are loaded in windows of 1000 rows. Moving down from the last row (or up from the
first) loads the neighbouring window, so a 22M-row table scrolls without loading it all.

## Mouse

- Click a catalog row: select it **and open it** (same as Enter).
- Click a pane: focus it. Drag a separator: resize. Wheel: move the marker.
- Click the command line to type a command.

The selected row is bold on slate blue while its pane has focus, gray otherwise.

## Commands

Expressions are raw DuckDB SQL, so quote odd column names: `:sort "table" desc`.
Tab completes table names (`:open`) and the first word of column arguments.

- `:open main.mrt` — table or view; `main.` may be omitted.
- `:sql select ticker, name from tickers where ticker like 'AA%'` — result replaces the left pane.
- `:filter stock_id = 500 and date >= DATE '2025-01-01'` — a bad expression is rolled back and
  the error is shown in the status band. `:filter` alone clears.
- `:sort volume desc` — `:sort` alone clears.
- `:schema` — one scan over the filtered rows, so it respects `:filter`.
- `:hist volume 30` — 40 bins by default. The range ends are shown under the first and last bar.
- `:plot close` / `:plot close date` — thinned to at most 4000 points.
- `:export csv /tmp/out.csv` — exports with the current filter and sort.

`:hist`, `:plot`, `:schema`, `:filter`, `:sort`, `:export` act on the left data pane even
when the catalog has focus. If no table is open yet they act on the catalog.

## Example session (sf19, a fundamentals + prices database)

```
Enter on prices_daily
:filter stock_id=500
:plot close
:hist volume 30
:schema
:export csv /tmp/s500.csv
```
