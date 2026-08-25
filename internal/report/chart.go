package report

import (
	"fmt"
	"strings"

	"buckley-lev/internal/welge"
)

func SaturationChart(res *welge.Result, width, height int) string {
	if width < 20 || height < 8 {
		return "(图宽/高过小，无法绘制)"
	}
	m := res.Model
	swInj := res.Case.Injection.SwInj
	xiShock := res.Shock.Speed
	xiMax := xiShock * 1.25

	grid := make([][]byte, height)
	for r := range grid {
		grid[r] = bytesOf(width, ' ')
	}

	toX := func(xi float64) int {
		x := xi / xiMax
		if x < 0 {
			x = 0
		}
		if x > 1 {
			x = 1
		}
		return int(x*float64(width-1) + 0.5)
	}
	toY := func(sw float64) int {
		y := (sw - m.Swc) / (swInj - m.Swc)
		if y < 0 {
			y = 0
		}
		if y > 1 {
			y = 1
		}
		return int((1-y)*float64(height-1) + 0.5)
	}
	set := func(x, y int, ch byte) {
		if x >= 0 && x < width && y >= 0 && y < height {
			grid[y][x] = ch
		}
	}

	for xi := 0.0; xi <= xiShock; xi += xiShock / 200 {
		sw, err := welge.RarefactionSaturation(m, res.Tangent, swInj, xi)
		if err != nil {
			continue
		}
		set(toX(xi), toY(sw), '#')
	}

	xFront := toX(xiShock)
	yUp := toY(res.Tangent.Swf)
	yDown := toY(m.Swc)
	for y := yUp; y <= yDown; y++ {
		set(xFront, y, '|')
	}
	set(xFront, yUp, '+')

	set(toX(0), toY(swInj), '@')

	for x := xFront + 1; x < width; x++ {
		set(x, toY(m.Swc), '.')
	}

	const leftW = 16
	var b strings.Builder
	b.WriteString("饱和度剖面（纵轴 Sw，横轴 ξ=x/t）\n")
	for r := 0; r < height; r++ {
		frac := 1 - float64(r)/float64(height-1)
		sw := m.Swc + frac*(swInj-m.Swc)
		label := ""
		if r == 0 {
			label = fmt.Sprintf("Sw=%.3f", sw)
		} else if r == height-1 {
			label = fmt.Sprintf("Sw=%.3f", sw)
		}
		b.WriteString(padLeft(label, leftW) + "│" + string(grid[r]) + "\n")
	}
	b.WriteString(strings.Repeat(" ", leftW) + "└" + strings.Repeat("─", width) + "\n")
	b.WriteString(padLeft("ξ=0", leftW) + " " + "ξ=ξs 在激波位置\n")
	return b.String()
}

func padLeft(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}

func bytesOf(n int, ch byte) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = ch
	}
	return out
}
