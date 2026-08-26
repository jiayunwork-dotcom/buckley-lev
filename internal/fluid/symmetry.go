package fluid

import "math"

func (m *Model) IsSymmetric() bool {
	return closeEq(m.Krw0, m.Kro0) && closeEq(m.MuW, m.MuO) && closeEq(m.Nw, m.No)
}

func (m *Model) MidSaturation() float64 {
	return m.Swc + 0.5*m.Delta
}

func (m *Model) SymmetryDeviation(grid int) float64 {
	if grid < 3 {
		grid = DefaultGrid
	}
	maxDev := 0.0
	for i := 0; i < grid; i++ {
		x := float64(i) / float64(grid-1)
		if x <= 0 || x >= 1 {
			continue
		}
		swL := m.Swc + x*m.Delta
		swR := m.Swc + (1-x)*m.Delta
		dev := math.Abs(m.F(swL) + m.F(swR) - 1)
		if dev > maxDev {
			maxDev = dev
		}
	}
	return maxDev
}

func closeEq(a, b float64) bool {
	return math.Abs(a-b) <= 1e-12*math.Max(1, math.Max(math.Abs(a), math.Abs(b)))
}
