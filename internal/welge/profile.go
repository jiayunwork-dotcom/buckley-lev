package welge

import (
	"fmt"

	"buckley-lev/internal/fluid"
)

// PointKind 描述剖面采样点的语义类别。
type PointKind string

// 剖面点类别。
const (
	KindInjection PointKind = "注入端"
	KindRarefaction PointKind = "稀疏波"
	KindShock     PointKind = "激波前缘"
	KindConnate   PointKind = "原始油藏"
)

// Point 是剖面上的一个采样点：无因次位置 ξ 与对应饱和度 Sw。
type Point struct {
	Xi   float64
	Sw   float64
	Kind PointKind
}

// Profile 是激波后方稀疏波 + 前缘的无因次剖面采样。
type Profile struct {
	// Points 按 ξ 升序排列（注入端在前，激波前缘在后），
	// 最后额外附带一个“前缘前方原始油藏”点（ξ>ξs, Sw=Swc）。
	Points []Point
	// XiShock 是激波所在的无因次位置 ξs。
	XiShock float64
	// SwFront 是激波前缘饱和度 Swf。
	SwFront float64
}

// DefaultFractions 是默认的 ξ/ξs 采样分数：0、0.1、…、1.0，
// 共 11 个稀疏波/前缘点。
func DefaultFractions() []float64 {
	out := make([]float64, 11)
	for i := range out {
		out[i] = float64(i) / 10
	}
	return out
}

// ParseFractions 把 "0,0.3,1" 形式的用户输入解析为 [0,1] 内的
// 分数序列（相对激波速度 ξs 的比例）。越界或非数字返回错误。
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

// BuildProfile 按 fractions（ξ/ξs 的比例）采样剖面。
//
// 剖面构成：
//   - ξ=0：注入端，Sw=Sw_inj；
//   - 0<ξ<ξs：稀疏波，由 f'(S)=ξ 反解（RarefactionSaturation）；
//   - ξ=ξs：激波前缘，Sw=Swf；
//   - ξ>ξs：原始油藏，Sw=Swc。
//
// 激波位置由 Welge 切线的割线斜率 ξs 决定，而不是前缘处的局部
// 斜率 df/dSw——后者会把激波放到错误的无因次位置。
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
		XiShock:  t.Slope,
		SwFront:  t.Swf,
		Points:   make([]Point, 0, len(fractions)+1),
	}
	seen := map[int64]bool{}
	for _, frac := range fractions {
		xi := frac * t.Slope
		// 去重：同一 ξ 只保留一个点。
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

	// 附加前缘前方的原始油藏点（ξ 略大于 ξs，饱和度回到 Swc）。
	prof.Points = append(prof.Points, Point{
		Xi:   t.Slope * 1.25,
		Sw:   m.Swc,
		Kind: KindConnate,
	})
	return prof, nil
}

// roundKey 把 ξ 值映射到用于去重的整数键（精度 1e-10）。
func roundKey(v float64) int64 {
	return int64(v*1e10 + 0.5)
}
