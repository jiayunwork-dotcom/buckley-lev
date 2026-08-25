package welge

import (
	"buckley-lev/internal/fluid"
)

type Shock struct {
	Speed          float64
	UpstreamF      float64
	DownstreamF    float64
	UpstreamSw     float64
	DownstreamSw   float64
	BreakthroughPV float64
}

func BuildShock(m *fluid.Model, t *Tangent) Shock {
	speed := t.Slope
	pv := 0.0
	if speed > 0 {
		pv = 1 / speed
	}
	return Shock{
		Speed:          speed,
		UpstreamF:      t.F,
		DownstreamF:    m.F(m.Swc),
		UpstreamSw:     t.Swf,
		DownstreamSw:   m.Swc,
		BreakthroughPV: pv,
	}
}
