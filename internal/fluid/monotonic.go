package fluid

import "math"

type MonotonicResult struct {
	OK       bool
	MinSlope float64
	AtSw     float64
	MaxSlope float64
	AtMaxSw  float64
	Grid     int
}

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
