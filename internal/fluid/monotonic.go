package fluid

import "math"

// MonotonicResult 是 f(Sw) 单调性检查的结论。
type MonotonicResult struct {
	// OK 表示在整个定义域 [Swc, 1−Sor] 上 f' ≥ −Tol。
	OK bool
	// MinSlope 是采样网格上观察到的最小 f'。
	MinSlope float64
	// AtSw 是 MinSlope 出现的饱和度。
	AtSw float64
	// MaxSlope 是采样网格上观察到的最大 f'。
	MaxSlope float64
	// AtMaxSw 是 MaxSlope 出现的饱和度。
	AtMaxSw float64
	// Grid 是采样点数。
	Grid int
}

// CheckMonotonic 在 [Swc, 1−Sor] 上采样 Grid 个内部点，
// 检查 f(Sw) 是否单调不减（允许端点附近 -1e-9 的浮点误差）。
// 对 Corey 幂次本构，f 恒为 S 形单调函数；检查的意义在于把这条
// 不变量做成显式可测的行为。
func (m *Model) CheckMonotonic(grid int) MonotonicResult {
	if grid < 3 {
		grid = DefaultGrid
	}
	lo, hi := m.Domain()
	res := MonotonicResult{
		MinSlope: math.Inf(1),
		MaxSlope: math.Inf(-1),
		Grid:     grid,
	}
	step := (hi - lo) / float64(grid-1)
	for i := 0; i < grid; i++ {
		sw := lo + step*float64(i)
		slope := m.FPrime(sw)
		if slope < res.MinSlope {
			res.MinSlope = slope
			res.AtSw = sw
		}
		if slope > res.MaxSlope {
			res.MaxSlope = slope
			res.AtMaxSw = sw
		}
	}
	res.OK = res.MinSlope >= -1e-9
	return res
}
