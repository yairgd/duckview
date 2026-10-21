package db

import "testing"

func TestPageHistSeriesStats(t *testing.T) {
	d, err := Open(":memory:", true)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	s := Source{Name: "t", Base: "select i, i::double * 2 as v, date '2020-01-01' + i::int as d from range(100) r(i)"}

	if n, err := d.Count(s); err != nil || n != 100 {
		t.Fatalf("count %d %v", n, err)
	}
	s.Where, s.Order = "i >= 10", "i desc"
	cols, types, rows, err := d.Page(s, 5, 0)
	if err != nil || len(cols) != 3 || types[2] != "DATE" || len(rows) != 5 || rows[0][0] != "99" {
		t.Fatalf("page %v %v %v %v", cols, types, rows, err)
	}
	labels, counts, err := d.Hist(s, "v", 4)
	if err != nil || len(counts) != 4 || labels[0] != "20" || labels[3] != "198" {
		t.Fatalf("hist %v %v %v", labels, counts, err)
	}
	if sum := counts[0] + counts[1] + counts[2] + counts[3]; sum != 90 {
		t.Fatalf("hist total %v", sum)
	}
	lx, vals, err := d.Series(s, "v", "d", 10)
	if err != nil || len(vals) > 11 || len(vals) < 9 || lx[len(lx)-1] != "2020-04-09" {
		t.Fatalf("series %v %v %v", lx, vals, err)
	}
	st, err := d.Stats(s, cols, types)
	if err != nil || st[0][4] != "10" || st[0][5] != "99" {
		t.Fatalf("stats %v %v", st, err)
	}
}
