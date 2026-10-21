# duckview

A fast, generic terminal browser for [DuckDB](https://duckdb.org) files, built on
[termforge](https://github.com/yairgd/termforge).

```
duckview [-w] file.duckdb
```

The database opens read-only unless you pass `-w`. Nothing in duckview knows about a
particular schema: the catalog, columns and types all come from `information_schema`.

```
┌ data (left, 70%) ──────────────────────┬ catalog (right) ───────────────┐
│ stock_id date       open  high  low …  │ schema name         type  rows │
│ 500      2025-08-15 9.98  10    9.97 … │ main   mrt          TABLE 732K │
│ …                                      │ main   prices_daily TABLE 21M  │
│ main.prices_daily rows 1-1000 of 235   │ main   tickers      TABLE 17K  │
│ hist / plot / schema panes stack here  │                                │
├────────────────────────────────────────┴────────────────────────────────┤
│ :command line                                                           │
└─────────────────────────────────────────────────────────────────────────┘
```

## Quick start

```bash
task build                       # bin/duckview (cgo: DuckDB is compiled in)
task run -- /path/to/file.duckdb
```

Move in the catalog with `j`/`k` or the arrow keys, press Enter (or click) to open a table in
the left pane, then use `:` commands.

| Command | Does |
|---|---|
| `:open [schema.]table` | show a table or view |
| `:sql <select …>` | show a query result |
| `:sort <expr> [desc]` / `:filter <expr>` | ORDER BY / WHERE, run in DuckDB; no argument clears |
| `:schema` | type, nulls, distinct, min, max for every column |
| `:hist <expr> [bins]` | histogram of a numeric expression |
| `:plot <y> [<x>]` | line graph; x defaults to the first date column |
| `:export csv\|parquet\|json <path>` | write the filtered, sorted rows |
| `:reload` `:close` `:only` `:help` `:quit` | housekeeping |

## Docs

- [Usage](docs/USAGE.md) — keys, mouse, every command with examples
- [Architecture](docs/ARCHITECTURE.md) — packages, the `Source` idea, panes, data flow
- [Development](docs/DEVELOPMENT.md) — build, tests, driving the app in tmux, gotchas, history

## License

BSD 3-Clause, see [LICENSE](LICENSE).
