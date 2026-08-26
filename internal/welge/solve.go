package welge

import (
	"fmt"

	"buckley-lev/internal/fluid"
	"buckley-lev/internal/schema"
)

type Result struct {
	Case      *schema.Case
	Model     *fluid.Model
	Tangent   *Tangent
	Shock     Shock
	Profile   *Profile
	Monotonic fluid.MonotonicResult
}

func Solve(c *schema.Case, fractions []float64) (*Result, error) {
	schema.ApplyDefaults(c)
	if err := schema.Validate(c); err != nil {
		return nil, err
	}

	m, err := fluid.NewModel(c.Rock.Swc, c.Rock.Sor,
		c.RelPerm.Krw0, c.RelPerm.Kro0,
		c.RelPerm.Nw, c.RelPerm.No,
		c.Fluid.MuW, c.Fluid.MuO)
	if err != nil {
		return nil, err
	}

	t, err := FindTangent(m)
	if err != nil {
		return nil, err
	}

	swInj := c.Injection.SwInj
	if swInj <= t.Swf {
		return nil, fmt.Errorf(
			"Sw 越界：注入饱和度 %g 低于前缘饱和度 Swf=%g，无法构造稀疏波（请提高注入饱和度或调整参数）",
			swInj, t.Swf)
	}

	prof, err := BuildProfile(m, t, swInj, fractions)
	if err != nil {
		return nil, err
	}

	mono := m.CheckMonotonic(fluid.DefaultGrid)
	if !mono.OK {
		return nil, fmt.Errorf(
			"f(Sw) 在 [%g, %g] 上非单调：最小斜率 %g 出现在 Sw=%g",
			m.Swc, 1-m.Sor, mono.MinSlope, mono.AtSw)
	}

	return &Result{
		Case:      c,
		Model:     m,
		Tangent:   t,
		Shock:     BuildShock(m, t),
		Profile:   prof,
		Monotonic: mono,
	}, nil
}

func SolveFile(path string, fractions []float64) (*Result, error) {
	c, err := schema.ParseFile(path)
	if err != nil {
		return nil, err
	}
	return Solve(c, fractions)
}
