package report

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type Table struct {
	headers []string
	rows    [][]string
	widths  []int
}

func NewTable(headers ...string) *Table {
	t := &Table{headers: headers, widths: make([]int, len(headers))}
	for i, h := range headers {
		t.widths[i] = utf8.RuneCountInString(h)
	}
	return t
}

func (t *Table) AddRow(cells ...string) {
	if len(cells) != len(t.headers) {
		panic(fmt.Sprintf("table: 行列数 %d != 表头列数 %d", len(cells), len(t.headers)))
	}
	t.rows = append(t.rows, cells)
	for i, c := range cells {
		if n := utf8.RuneCountInString(c); n > t.widths[i] {
			t.widths[i] = n
		}
	}
}

func (t *Table) String() string {
	var b strings.Builder
	pad := func(row []string) {
		for i, cell := range row {
			if i > 0 {
				b.WriteString("  ")
			}
			b.WriteString(padRunes(cell, t.widths[i]))
		}
		b.WriteString("\n")
	}
	if len(t.headers) > 0 {
		pad(t.headers)
		b.WriteString(strings.Repeat("-", sumWidths(t.widths)+2*(len(t.headers)-1)))
		b.WriteString("\n")
	}
	for _, row := range t.rows {
		pad(row)
	}
	return b.String()
}

func sumWidths(w []int) int {
	s := 0
	for _, v := range w {
		s += v
	}
	return s
}

func padRunes(s string, width int) string {
	n := utf8.RuneCountInString(s)
	if n >= width {
		return s
	}
	return s + strings.Repeat(" ", width-n)
}

func Num(v float64, prec int) string {
	return fmt.Sprintf("%.*f", prec, v)
}

func Frac(v float64) string {
	return fmt.Sprintf("%.1f%%", v*100)
}
