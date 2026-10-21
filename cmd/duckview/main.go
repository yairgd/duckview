package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yairgd/duckview/internal/app"
	"github.com/yairgd/duckview/internal/db"
)

func main() {
	write := flag.Bool("w", false, "open the database read-write")
	flag.Usage = func() { fmt.Fprintln(os.Stderr, "usage: duckview [-w] <file.duckdb>") }
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	d, err := db.Open(flag.Arg(0), *write)
	if err != nil {
		fmt.Fprintf(os.Stderr, "duckview: %v\n", err)
		os.Exit(1)
	}
	defer d.Close()
	a, err := app.New(d)
	if err != nil {
		fmt.Fprintf(os.Stderr, "duckview: %v\n", err)
		os.Exit(1)
	}
	defer a.Close()
	a.Run()
}
