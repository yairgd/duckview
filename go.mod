module github.com/yairgd/duckview

go 1.25.5

require (
	github.com/gdamore/tcell/v2 v2.13.8
	github.com/marcboeker/go-duckdb v1.8.5
	github.com/yairgd/termforge v0.11.0
)

require (
	github.com/apache/arrow-go/v18 v18.1.0 // indirect
	github.com/creack/pty v1.1.24 // indirect
	github.com/ebitengine/purego v0.10.1 // indirect
	github.com/gdamore/encoding v1.0.1 // indirect
	github.com/gitpod-io/xterm-go v0.0.0-20260828130427-e62e9648055e // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/google/flatbuffers v25.1.24+incompatible // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/cpuid/v2 v2.2.9 // indirect
	github.com/lucasb-eyer/go-colorful v1.3.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	golang.design/x/clipboard v0.9.0 // indirect
	golang.design/x/x11 v0.2.0 // indirect
	golang.org/x/exp v0.0.0-20250128182459-e0ece0dbea4c // indirect
	golang.org/x/exp/shiny v0.0.0-20250606033433-dcc06ee1d476 // indirect
	golang.org/x/image v0.28.0 // indirect
	golang.org/x/mobile v0.0.0-20250606033058-a2a15c67f36f // indirect
	golang.org/x/mod v0.29.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/telemetry v0.0.0-20251008203120-078029d740a8 // indirect
	golang.org/x/term v0.37.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	golang.org/x/tools v0.38.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
)

// Local termforge checkout (HistogramWidget is not released yet). Remove once tagged.
//replace github.com/yairgd/termforge => ../termforge
