package welge

import (
	"buckley-lev/internal/fluid"
)

// Shock 描述激波前缘的运动学量。
type Shock struct {
	// Speed 是无因次激波速度 ξs = f(Swf)/(Swf−Swc)。
	// 在 ξ=x/t 坐标系下，激波位于 ξ=ξs。
	Speed float64
	// UpstreamF 是激波上游（前缘后方）的分流量 f(Swf)。
	UpstreamF float64
	// DownstreamF 是激波下游（前缘前方）的分流量 f(Swc)=0。
	DownstreamF float64
	// UpstreamSw 是激波上游饱和度 Swf。
	UpstreamSw float64
	// DownstreamSw 是激波下游饱和度 Swc。
	DownstreamSw float64
	// BreakthroughPV 是无因次突破注入孔隙体积 = 1/ξs。
	// 表示前缘到达出口端时累计注入的孔隙体积倍数。
	BreakthroughPV float64
}

// BuildShock 由 Welge 切点构造激波运动学量。
//
// Rankine–Hugoniot：激波速度 = [f(Swf)−f(Swc)]/[Swf−Swc]。
// 由于 f(Swc)=0（束缚水处水不流动），速度退化为 f(Swf)/(Swf−Swc)，
// 与 Welge 切线的割线斜率一致——这正是“前缘用切线、不用当地
// df/dSw”的物理依据。
func BuildShock(m *fluid.Model, t *Tangent) Shock {
	speed := t.Slope
	pv := 0.0
	if speed > 0 {
		pv = 1 / speed
	}
	return Shock{
		Speed:        speed,
		UpstreamF:    t.F,
		DownstreamF:  m.F(m.Swc),
		UpstreamSw:   t.Swf,
		DownstreamSw: m.Swc,
		BreakthroughPV: pv,
	}
}
