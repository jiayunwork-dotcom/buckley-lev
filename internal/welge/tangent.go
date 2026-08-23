// Package welge 实现 Buckley–Leverett 的解析解：Welge 切点求根、
// 激波（Rankine–Hugoniot）、激波后方稀疏波（由 df/dSw 反解）与
// 无因次剖面（ξ=x/t），以及交叉规则探针。
package welge

import (
	"fmt"
	"math"

	"buckley-lev/internal/fluid"
)

// Tangent 是 Welge 切点：从 (Swc,0) 向 f(Sw) 作切线的切点。
type Tangent struct {
	// Swf 是激波前缘饱和度（切点饱和度）。
	Swf float64
	// F 是 f(Swf)。
	F float64
	// Slope 是割线/激波无因次速度 f(Swf)/(Swf−Swc)。
	Slope float64
	// LocalSlope 是切点处局部斜率 f'(Swf)。切线成立时
	// LocalSlope 与 Slope 相等。
	LocalSlope float64
	// GridHMax 是粗网格扫描到的最大 h(S)=f/(S−Swc) 值，用于
	// 与最终切点交叉验证。
	GridHMax float64
	// GridHMaxSw 是粗网格最大 h 出现的饱和度。
	GridHMaxSw float64
}

// h 返回割线斜率候选值 f(S)/(S−Swc)。
func h(m *fluid.Model, sw float64) float64 {
	denom := sw - m.Swc
	if denom <= 0 {
		return 0
	}
	return m.F(sw) / denom
}

// scanTangent 在定义域内部网格上找到 h 的全局最大值（粗定位）。
func scanTangent(m *fluid.Model, n int) (bestSw, bestH float64, ok bool) {
	lo, hi := m.Domain()
	bestSw, bestH = 0, -1
	for i := 1; i <= n; i++ {
		sw := lo + (hi-lo)*float64(i)/float64(n+1)
		val := h(m, sw)
		if val > bestH {
			bestSw, bestH = sw, val
		}
	}
	if bestH > 0 {
		ok = true
	}
	return bestSw, bestH, ok
}

// FindTangent 求 Welge 切点。
//
// 流程：
//  1. 粗网格扫描 h(S)=f(S)/(S−Swc) 定位全局最大值；
//  2. 以该位置为括弧做黄金分割搜索，收敛到 h 的精确极大值；
//  3. 在黄金分割结果附近用二分法求切线方程 g(S)=f'(S)−h(S)=0
//     的根，作为最终 Swf；
//  4. 校验切线自洽：|f'(Swf) − h(Swf)| 必须小于容差，
//     且 Swf 严格落在 (Swc, 1−Sor) 内。
//
// 粗网格最大值落在端点附近（无内点切点）或切点越界时返回错误，
// 对应“切线无解或 Sw 越界”。
func FindTangent(m *fluid.Model) (*Tangent, error) {
	const scanN = 512
	lo, hi := m.Domain()

	bestSw, bestH, ok := scanTangent(m, scanN)
	if !ok || bestH <= 0 {
		return nil, fmt.Errorf("Welge 切线无解：f(S)/(S−Swc) 在 (%g, %g) 内没有正的最大值", lo, hi)
	}

	// 端点对照：切点必须严格在内部。若 h 的最大值贴着右端点，
	// 说明没有内点切点（例如水相流度极低、f 几乎处处为 0）。
	rightH := h(m, hi)
	if bestSw >= hi-(hi-lo)/scanN && bestH <= rightH*(1+1e-9)+1e-12 {
		return nil, fmt.Errorf("Welge 切线无解：割线斜率最大值位于右端点，无可动油前缘（h(%.4f)=%.6f ≤ h(1−Sor)=%.6f）",
			bestSw, bestH, rightH)
	}

	// 黄金分割：在粗定位点附近的小括弧内最大化 h。
	step := (hi - lo) / float64(scanN)
	left := math.Max(lo+Eps, bestSw-step)
	right := math.Min(hi-Eps, bestSw+step)
	swGS := goldenSection(m, left, right)

	// 二分法求 g(S)=0 精确化。
	swBis, err := bisectTangent(m, swGS, step)
	if err != nil {
		return nil, err
	}

	t := &Tangent{
		Swf:        swBis,
		F:          m.F(swBis),
		Slope:      h(m, swBis),
		LocalSlope: m.FPrime(swBis),
		GridHMax:   bestH,
		GridHMaxSw: bestSw,
	}

	// 自洽校验。
	if t.Swf <= lo || t.Swf >= hi {
		return nil, fmt.Errorf("Welge 切点越界：Swf=%g 不在 (%g, %g) 内", t.Swf, lo, hi)
	}
	scale := math.Max(1, math.Abs(t.Slope))
	if math.Abs(t.LocalSlope-t.Slope) > fluid.RelTol*scale {
		return nil, fmt.Errorf(
			"Welge 切线不自洽：f'(Swf)=%.9g 与 f(Swf)/(Swf−Swc)=%.9g 偏差超容差",
			t.LocalSlope, t.Slope)
	}
	return t, nil
}

// goldenSection 在 [a,b] 上最大化 h(S)。h 在切点附近单峰，
// 黄金分割保证线性收敛且无需导数信息。
func goldenSection(m *fluid.Model, a, b float64) float64 {
	const (
		golden = 0.6180339887498949
		iters  = 120
	)
	c := b - golden*(b-a)
	d := a + golden*(b-a)
	for i := 0; i < iters; i++ {
		if math.Abs(b-a) < Eps {
			break
		}
		if h(m, c) < h(m, d) {
			a = c
			c = d
			d = a + golden*(b-a)
		} else {
			b = d
			d = c
			c = b - golden*(b-a)
		}
	}
	return 0.5 * (a + b)
}

// bisectTangent 在 swGS 附近对 g(S)=f'(S)−h(S) 二分求根。
// g 在切点左侧为正、右侧为负（h 先升后降），保证符号变化存在。
func bisectTangent(m *fluid.Model, swGS, step float64) (float64, error) {
	lo, hi := m.Domain()
	a := math.Max(lo+Eps, swGS-step)
	b := math.Min(hi-Eps, swGS+step)

	g := func(s float64) float64 { return m.FPrime(s) - h(m, s) }

	ga := g(a)
	gb := g(b)
	if ga*gb > 0 {
		// 括弧内无符号变化：退化为黄金分割结果（自洽校验会兜底）。
		return swGS, nil
	}
	// 归一化方向：保证 a 端 g>0、b 端 g<0（h 在切点左侧上升）。
	if ga < 0 {
		a, b, ga, gb = b, a, gb, ga
	}
	for i := 0; i < 160; i++ {
		mid := 0.5 * (a + b)
		if math.Abs(b-a) < Eps {
			return mid, nil
		}
		gm := g(mid)
		if gm > 0 {
			a = mid
			ga = gm
		} else {
			b = mid
			gb = gm
		}
	}
	return 0.5 * (a + b), nil
}
