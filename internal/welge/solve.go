package welge

import (
	"fmt"

	"buckley-lev/internal/fluid"
	"buckley-lev/internal/schema"
)

// Result 是一次完整求解的全部产物：输入、本构模型、Welge 切点、
// 激波运动学量与稀疏波剖面。
type Result struct {
	// Case 是求解所用的输入（已应用默认值）。
	Case *schema.Case
	// Model 是本构模型。
	Model *fluid.Model
	// Tangent 是 Welge 切点。
	Tangent *Tangent
	// Shock 是激波运动学量。
	Shock Shock
	// Profile 是稀疏波 + 前缘剖面。
	Profile *Profile
	// Monotonic 是 f 单调性检查结论。
	Monotonic fluid.MonotonicResult
}

// Solve 完成一次完整求解：校验输入 → 构造本构 → 求 Welge 切点 →
// 构造激波与剖面。fractions 为 nil 时使用默认 11 点采样。
//
// 任何一步失败都返回带上下文的错误，调用方负责把错误写 stderr
// 并以非零退出码结束。
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

// SolveFile 读取并求解一个算例文件，是 CLI 的主入口。
func SolveFile(path string, fractions []float64) (*Result, error) {
	c, err := schema.ParseFile(path)
	if err != nil {
		return nil, err
	}
	return Solve(c, fractions)
}
