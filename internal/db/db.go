// Package db is the thin DuckDB layer duckview needs: schema-agnostic paging,
// counting, binning and stats over any SQL relation.
package db

import (
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	_ "github.com/marcboeker/go-duckdb"
)

type DB struct{ *sql.DB }

// Open opens a DuckDB file, read-only unless write is set.
func Open(path string, write bool) (*DB, error) {
	dsn := path
	if !write {
		dsn += "?access_mode=read_only"
	}
	d, err := sql.Open("duckdb", dsn)
	if err != nil {
		return nil, err
	}
	if err := d.Ping(); err != nil {
		return nil, err
	}
	return &DB{d}, nil
}

// Source is a relation plus the user's filter and sort. Base is any SELECT.
type Source struct {
	Name, Base, Where, Order string
}

// rel is the filtered relation, without the sort.
func (s Source) rel() string {
	r := "(select * from (" + s.Base + ") _t"
	if s.Where != "" {
		r += " where " + s.Where
	}
	return r + ")"
}

func (s Source) query() string {
	q := "select * from " + s.rel() + " _f"
	if s.Order != "" {
		q += " order by " + s.Order
	}
	return q
}

// Quote quotes an identifier.
func Quote(id string) string { return `"` + strings.ReplaceAll(id, `"`, `""`) + `"` }

func (d *DB) Count(s Source) (n int64, err error) {
	err = d.QueryRow("select count(*) from " + s.rel() + " _c").Scan(&n)
	return
}

// Page returns one window of rows as display strings, plus column names and
// DuckDB type names.
func (d *DB) Page(s Source, limit, offset int) (cols, types []string, rows [][]string, err error) {
	rs, err := d.Query(fmt.Sprintf("%s limit %d offset %d", s.query(), limit, offset))
	if err != nil {
		return nil, nil, nil, err
	}
	defer rs.Close()
	cols, err = rs.Columns()
	if err != nil {
		return
	}
	cts, _ := rs.ColumnTypes()
	for _, ct := range cts {
		types = append(types, ct.DatabaseTypeName())
	}
	vals := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}
	for rs.Next() {
		if err = rs.Scan(ptrs...); err != nil {
			return
		}
		row := make([]string, len(cols))
		for i, v := range vals {
			row[i] = Format(v)
		}
		rows = append(rows, row)
	}
	err = rs.Err()
	return
}

// Format renders one cell value for display.
func Format(v any) string {
	switch x := v.(type) {
	case nil:
		return "·"
	case time.Time:
		if x.Hour() == 0 && x.Minute() == 0 && x.Second() == 0 && x.Nanosecond() == 0 {
			return x.Format("2006-01-02")
		}
		return x.Format("2006-01-02 15:04:05")
	case float64:
		return strconv.FormatFloat(x, 'g', 6, 64)
	case float32:
		return strconv.FormatFloat(float64(x), 'g', 6, 32)
	case []byte:
		return string(x)
	case string:
		return strings.NewReplacer("\n", " ", "\t", " ").Replace(x)
	}
	return fmt.Sprint(v)
}

// Hist bins the numeric expr over the filtered relation into n equal-width
// bins. labels carry the range ends on the first and last bin.
func (d *DB) Hist(s Source, expr string, n int) (labels []string, counts []float64, err error) {
	var lo, hi sql.NullFloat64
	num := "(" + expr + ")::double"
	src := s.rel() + " _h"
	if err = d.QueryRow("select min("+num+"), max("+num+") from "+src).Scan(&lo, &hi); err != nil {
		return
	}
	if !lo.Valid {
		return nil, nil, fmt.Errorf("no values")
	}
	counts = make([]float64, n)
	rs, err := d.Query(fmt.Sprintf(
		"select least(floor((%[1]s - %[2]g) / %[3]g * %[4]d), %[4]d - 1)::int b, count(*) from %[5]s where %[1]s is not null group by 1",
		num, lo.Float64, math.Max(hi.Float64-lo.Float64, 1e-12), n, src))
	if err != nil {
		return
	}
	defer rs.Close()
	for rs.Next() {
		var b int
		var c float64
		if err = rs.Scan(&b, &c); err != nil {
			return
		}
		counts[max(min(b, n-1), 0)] = c
	}
	labels = make([]string, n)
	labels[0] = strconv.FormatFloat(lo.Float64, 'g', 5, 64)
	labels[n-1] = strconv.FormatFloat(hi.Float64, 'g', 5, 64)
	return labels, counts, rs.Err()
}

// Series returns y against x (ordered by x; row order when x is empty),
// thinned to at most max points.
func (d *DB) Series(s Source, y, x string, max int) (labels []string, vals []float64, err error) {
	if x == "" {
		x = "row_number() over ()"
	}
	inner := fmt.Sprintf("select xs, y, row_number() over (order by xv) rn from (select (%s) xv, (%s)::varchar xs, (%s)::double y from %s _s) where y is not null",
		x, x, y, s.rel())
	var n int
	if err = d.QueryRow("select count(*) from (" + inner + ")").Scan(&n); err != nil {
		return
	}
	if n == 0 {
		return nil, nil, fmt.Errorf("no values")
	}
	step := (n + max - 1) / max
	rs, err := d.Query(fmt.Sprintf("select xs, y from (%s) where rn %% %d = 0 or rn = %d order by rn", inner, step, n))
	if err != nil {
		return
	}
	defer rs.Close()
	for rs.Next() {
		var l string
		var v float64
		if err = rs.Scan(&l, &v); err != nil {
			return
		}
		labels, vals = append(labels, l), append(vals, v)
	}
	return labels, vals, rs.Err()
}

// Stats is one row per column of the filtered relation: name, type, nulls,
// distinct (approximate), min, max — computed in a single scan.
func (d *DB) Stats(s Source, cols, types []string) ([][]string, error) {
	var sel []string
	sel = append(sel, "count(*)")
	for _, c := range cols {
		q := Quote(c)
		sel = append(sel, "count("+q+")", "approx_count_distinct("+q+")", "min("+q+")::varchar", "max("+q+")::varchar")
	}
	vals := make([]any, len(sel))
	ptrs := make([]any, len(sel))
	for i := range vals {
		ptrs[i] = &vals[i]
	}
	if err := d.QueryRow("select " + strings.Join(sel, ",") + " from " + s.rel() + " _s").Scan(ptrs...); err != nil {
		return nil, err
	}
	total, _ := vals[0].(int64)
	rows := make([][]string, len(cols))
	for i, c := range cols {
		v := vals[1+i*4:]
		nn, _ := v[0].(int64)
		rows[i] = []string{c, types[i], strconv.FormatInt(total-nn, 10), Format(v[1]), Format(v[2]), Format(v[3])}
	}
	return rows, nil
}

// Export writes the filtered, sorted relation to path (csv, parquet or json).
func (d *DB) Export(s Source, format, path string) error {
	opt := map[string]string{"csv": "(format csv, header)", "parquet": "(format parquet)", "json": "(format json, array true)"}[format]
	if opt == "" {
		return fmt.Errorf("format must be csv, parquet or json")
	}
	_, err := d.Exec(fmt.Sprintf("copy (%s) to '%s' %s", s.query(), strings.ReplaceAll(path, "'", "''"), opt))
	return err
}

// Catalog is a SELECT listing every table and view with its row and column count.
func (d *DB) Catalog() (string, error) {
	rs, err := d.Query(`select t.table_schema, t.table_name, t.table_type,
		(select count(*) from information_schema.columns c where c.table_schema = t.table_schema and c.table_name = t.table_name)
		from information_schema.tables t order by 1, 2`)
	if err != nil {
		return "", err
	}
	defer rs.Close()
	var parts []string
	for rs.Next() {
		var sc, name, typ string
		var nc int
		if err := rs.Scan(&sc, &name, &typ, &nc); err != nil {
			return "", err
		}
		parts = append(parts, fmt.Sprintf("select '%s' as \"schema\", '%s' as name, '%s' as type, (select count(*) from %s.%s) as \"rows\", %d as cols",
			sc, strings.ReplaceAll(name, "'", "''"), typ, Quote(sc), Quote(name), nc))
	}
	if len(parts) == 0 {
		return "select 'main' as \"schema\", '' as name, '' as type, 0 as \"rows\", 0 as cols where false", rs.Err()
	}
	return strings.Join(parts, " union all "), rs.Err()
}
