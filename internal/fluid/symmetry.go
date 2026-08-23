package fluid

import "math"

// IsSymmetric 报告本构是否关于可动油带中点对称。
// 当且仅当 krw0=kro0、μw=μo 且 nw=no 时，f(x) 满足
// f(S*) + f(1−S*) = 1，其中点处 f(0.5)=0.5。
func (m *Model) IsSymmetric() bool {
	return closeEq(m.Krw0, m.Kro0) && closeEq(m.MuW, m.MuO) && closeEq(m.Nw, m.No)
}

// MidSaturation 返回可动油带中点 Swc + Δ/2。
func (m *Model) MidSaturation() float64 {
	return m.Swc + 0.5*m.Delta
}

// SymmetryDeviation 采样检查对称性 |f(S*) + f(1−S*) − 1| 的最大偏差。
// 对称本构下偏差应为 0；非对称本构返回正偏差供探针展示。
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

// closeEq 用于判断本构“对称”这类宽松相等，容差取 1e-12。
func closeEq(a, b float64) bool {
	return math.Abs(a-b) <= 1e-12*math.Max(1, math.Max(math.Abs(a), math.Abs(b)))
}
