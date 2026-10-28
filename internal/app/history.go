package app

import (
	"os"
	"path/filepath"
	"strings"
)

const historyMax = 500

// The cmdline history is kept in <user config dir>/duckview/history, one entry
// per line exactly as typed, so ':cmd' for commands and '/pat' for searches (oldest first).
func historyFile() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "duckview", "history")
}

func (a *App) loadHistory() {
	path := historyFile()
	if path == "" {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var cmds, search []string
	for _, line := range strings.Split(string(data), "\n") {
		if len(line) < 2 {
			continue
		}
		switch line[0] {
		case ':':
			cmds = append(cmds, line)
		case '/':
			search = append(search, line)
		}
	}
	a.cmdWidget.LoadCommandHistory(cmds)
	a.cmdWidget.LoadSearchHistory(search)
}

func (a *App) saveHistory() {
	path := historyFile()
	if path == "" || a.cmdWidget == nil {
		return
	}
	var b strings.Builder
	add := func(items []string) {
		if len(items) > historyMax {
			items = items[len(items)-historyMax:]
		}
		for _, it := range items {
			if it != "" && !strings.Contains(it, "\n") {
				b.WriteString(it + "\n")
			}
		}
	}
	add(a.cmdWidget.CommandHistoryItems())
	add(a.cmdWidget.SearchHistoryItems())
	_ = os.MkdirAll(filepath.Dir(path), 0o700)
	_ = os.WriteFile(path, []byte(b.String()), 0o600)
}
