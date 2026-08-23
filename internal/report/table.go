// Package report 把求解结果渲染成命令行可读的文本：对齐表格、
// 剖面摘要、ASCII 饱和度图与交叉规则探针输出。
package report

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Table 是一个简单的对齐表格，用于把数值结果排成等宽列。
// 宽度按字符（rune）计算，CJK 文本与 ASCII 数字混排时也能对齐。
type Table struct {
	headers []string
	rows    [][]string
	widths  []int
}

// NewTable 创建带表头的表格。
func NewTable(headers ...string) *Table {
	t := &Table{headers: headers, widths: make([]int, len(headers))}
	for i, h := range headers {
		t.widths[i] = utf8.RuneCountInString(h)
	}
	return t
}

// AddRow 追加一行，列数必须与表头一致。
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

// String 输出对齐后的文本，每列左对齐、列间两个空格。
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

// padRunes 按字符数左对齐补齐空格。
func padRunes(s string, width int) string {
	n := utf8.RuneCountInString(s)
	if n >= width {
		return s
	}
	return s + strings.Repeat(" ", width-n)
}

// Num 把浮点数格式化为给定小数位数的定宽文本。
func Num(v float64, prec int) string {
	return fmt.Sprintf("%.*f", prec, v)
}

// Frac 把 [0,1] 内的分数格式化为百分比文本。
func Frac(v float64) string {
	return fmt.Sprintf("%.1f%%", v*100)
}
