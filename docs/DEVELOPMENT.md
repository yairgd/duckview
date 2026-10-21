# Development

```bash
task build     # go build + bin/duckview
task test      # go test ./...
task check     # build, vet, test, gofmt
```

`go.mod` has `replace github.com/yairgd/termforge => ../termforge` because `HistogramWidget` is not
in a tagged termforge yet. After tagging termforge, remove the `replace` and bump the version.

## Tests

`internal/db/db_test.go` runs Page/Hist/Series/Stats against an in-memory DuckDB. The UI has no
unit tests; it was verified by driving the real binary in tmux.

## Driving the app in tmux

```bash
tmux new-session -d -s dv -x 150 -y 30 "bin/duckview file.duckdb 2>err.log; read"
tmux send-keys -t dv jjj Enter          # keys
tmux send-keys -t dv ':filter x=1' Enter
tmux send-keys -t dv -l $'\e[<0;120;8M' ; tmux send-keys -t dv -l $'\e[<0;120;8m'   # SGR click at col 120,row 8
tmux capture-pane -t dv -p              # add -e to see colors
```

## Gotchas learned

- Taskfile `desc:` containing `word: word` breaks YAML — avoid colons or quote it.
- LeafRest arguments can arrive as separate words; `arg()` joins them. Rest-arg commands need the whole text.
- A window function cannot sit inside another window's `order by`; `Series` computes the x
  expression in a subselect first.
- `Split` keeps focus on the original leaf and puts the new widget second; call `Rebalance()` or the
  data pane takes almost all the space.
- termforge has no tab closing, which is why `:open` replaces the left pane instead of adding tabs.
- DuckDB file format must match go-duckdb's engine: v1.8.5 reads the `sf19` file (storage v64).
- Queries run on the UI thread; heavy scans (e.g. `:schema` on 22M rows) block briefly.

## Ideas not built

- Async queries via `PostInterrupt`.
- Per-database config: saved queries, default joins (e.g. tickers ↔ prices).
- `Enter` on a `:schema` row → `:hist` for that column (one `onEnter` assignment).
- Column hiding/reordering, right-aligned numbers, search (`/`).

## History

Designed and built in one session against `/home/yair/value_inveting_2026/sf19` (1.9 GB, 17 tables,
`prices_daily` 21.9M rows). Order: probe schema with a Go snippet → db layer → app → fix plot SQL and
arg parsing → two-pane layout → click-to-open → selected-row color (same as the tradedeck strategies list).
