package welge

import (
	"fmt"
	"math"

	"buckley-lev/internal/fluid"
)

type Tangent struct {
	Swf        float64
	F          float64
	Slope      float64
	LocalSlope float64
	GridHMax   float64
	GridHMaxSw float64
}

func h(m *fluid.Model, sw float64) float64 {
	denom := sw - m.Swc
	if denom <= 0 {
		return 0
	}
	return m.F(sw) / denom
}

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

func FindTangent(m *fluid.Model) (*Tangent, error) {
	const scanN = 512
	lo, hi := m.Domain()

	bestSw, bestH, ok := scanTangent(m, scanN)
	if !ok || bestH <= 0 {
		return nil, fmt.Errorf("Welge 切线无解：f(S)/(S−Swc) 在 (%g, %g) 内没有正的最大值", lo, hi)
	}

	rightH := h(m, hi)
	if bestSw >= hi-(hi-lo)/scanN && bestH <= rightH*(1+1e-9)+1e-12 {
		return nil, fmt.Errorf("Welge 切线无解：割线斜率最大值位于右端点，无可动油前缘（h(%.4f)=%.6f ≤ h(1−Sor)=%.6f）",
			bestSw, bestH, rightH)
	}

	step := (hi - lo) / float64(scanN)
	left := math.Max(lo+Eps, bestSw-step)
	right := math.Min(hi-Eps, bestSw+step)
	swGS := goldenSection(m, left, right)

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

func bisectTangent(m *fluid.Model, swGS, step float64) (float64, error) {
	lo, hi := m.Domain()
	a := math.Max(lo+Eps, swGS-step)
	b := math.Min(hi-Eps, swGS+step)

	g := func(s float64) float64 { return m.FPrime(s) - h(m, s) }

	ga := g(a)
	gb := g(b)
	if ga*gb > 0 {
		return swGS, nil
	}
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
