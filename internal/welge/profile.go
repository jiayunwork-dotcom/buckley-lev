package welge

import (
	"fmt"

	"buckley-lev/internal/fluid"
)

type PointKind string

const (
	KindInjection   PointKind = "注入端"
	KindRarefaction PointKind = "稀疏波"
	KindShock       PointKind = "激波前缘"
	KindConnate     PointKind = "原始油藏"
)

type Point struct {
	Xi   float64
	Sw   float64
	Kind PointKind
}

type Profile struct {
	Points  []Point
	XiShock float64
	SwFront float64
}

func DefaultFractions() []float64 {
	out := make([]float64, 11)
	for i := range out {
		out[i] = float64(i) / 10
	}
	return out
}

func ParseFractions(s string) ([]float64, error) {
	var out []float64
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			if start == i {
				return nil, fmt.Errorf("ξ 分数列表为空项")
			}
			var f float64
			if _, err := fmt.Sscanf(s[start:i], "%g", &f); err != nil {
				return nil, fmt.Errorf("ξ 分数 %q 不是数字", s[start:i])
			}
			if f < 0 || f > 1 {
				return nil, fmt.Errorf("ξ 分数 %g 必须在 [0,1] 内", f)
			}
			out = append(out, f)
			start = i + 1
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("ξ 分数列表为空")
	}
	return out, nil
}

func BuildProfile(m *fluid.Model, t *Tangent, swInj float64, fractions []float64) (*Profile, error) {
	if swInj <= m.Swc {
		return nil, fmt.Errorf("注入饱和度 %g 必须高于 Swc=%g", swInj, m.Swc)
	}
	if swInj <= t.Swf {
		return nil, fmt.Errorf("Sw 越界：注入饱和度 %g 低于前缘饱和度 Swf=%g，无法构造剖面", swInj, t.Swf)
	}
	if len(fractions) == 0 {
		fractions = DefaultFractions()
	}

	prof := &Profile{
		XiShock: t.Slope,
		SwFront: t.Swf,
		Points:  make([]Point, 0, len(fractions)+1),
	}
	seen := map[int64]bool{}
	for _, frac := range fractions {
		xi := frac * t.Slope
		key := roundKey(xi)
		if seen[key] {
			continue
		}
		seen[key] = true

		sw, err := RarefactionSaturation(m, t, swInj, xi)
		if err != nil {
			return nil, err
		}
		kind := KindRarefaction
		switch {
		case xi <= Eps:
			kind = KindInjection
		case xi >= t.Slope-Eps:
			kind = KindShock
		}
		prof.Points = append(prof.Points, Point{Xi: xi, Sw: sw, Kind: kind})
	}

	prof.Points = append(prof.Points, Point{
		Xi:   t.Slope * 1.25,
		Sw:   m.Swc,
		Kind: KindConnate,
	})
	prof.Points = overlayRareScratch(prof.Points)
	return prof, nil
}

func roundKey(v float64) int64 {
	return int64(v*1e10 + 0.5)
}
