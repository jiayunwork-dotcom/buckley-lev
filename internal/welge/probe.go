package welge

import (
	"fmt"

	"buckley-lev/internal/fluid"
	"buckley-lev/internal/schema"
)

// ViscosityCase 是一次粘度扫描的结果：给定水粘度下重解 Welge 切点。
type ViscosityCase struct {
	// MuW 是本次扫描使用的水粘度。
	MuW float64
	// MobilityRatio 是对应的端点流度比 (krw0/μw)/(kro0/μo)。
	MobilityRatio float64
	// Swf 是重解后的前缘饱和度。
	Swf float64
	// ShockSpeed 是重解后的无因次激波速度。
	ShockSpeed float64
}

// ResidualOilCase 是一次残余油扫描的结果。
type ResidualOilCase struct {
	// Sor 是本次扫描使用的残余油饱和度。
	Sor float64
	// TerminalSw 是注入端/末端饱和度 1−Sor。
	TerminalSw float64
	// Swf 是重解后的前缘饱和度。
	Swf float64
	// ShockSpeed 是重解后的无因次激波速度。
	ShockSpeed float64
}

// SymmetryCase 是对称本构的验证结果。
type SymmetryCase struct {
	// FAtMid 是 f(可动油带中点) 的数值，应为 0.5。
	FAtMid float64
	// Deviation 是 max|f(S*)+f(1−S*)−1|，应接近 0。
	Deviation float64
	// Mid 是切点饱和度（对称本构下 = 中点 + 切点偏移）。
	Swf float64
}

// ProbeResult 汇总交叉规则探针的输出。
type ProbeResult struct {
	// BaseSwf 是原始算例的前缘饱和度。
	BaseSwf float64
	// ViscosityScan 是“只降低水粘度”实验序列（μo 不变）。
	ViscosityScan []ViscosityCase
	// ResidualScan 是“只增大 Sor”实验序列（其余不变）。
	ResidualScan []ResidualOilCase
	// Symmetry 是对称本构实验。
	Symmetry SymmetryCase
}

// viscosityMultipliers 是粘度扫描使用的水粘度倍率（由大往小）。
var viscosityMultipliers = []float64{1.0, 0.5, 0.25, 0.125}

// residualMultipliers 是残余油扫描使用的 Sor 倍率。
var residualMultipliers = []float64{1.0, 1.5, 2.0}

// ProbeCrossRules 按第 1 轮提问的交叉规则做三组实验：
//
//  1. 只把水粘度相对油粘度降低（μo 不变）→ 观察 Swf 与激波速度；
//  2. 只把 Sor 加大 → 可动油变少、末端饱和度左移；
//  3. μw=μo、端点与指数相同的对称本构 → f 在中点对称、f=0.5。
//
// 每组实验都重新求解 Welge 切点，结果可用于校验交叉规则的可复现性。
func ProbeCrossRules(c *schema.Case) (*ProbeResult, error) {
	m, err := fluid.NewModel(c.Rock.Swc, c.Rock.Sor,
		c.RelPerm.Krw0, c.RelPerm.Kro0,
		c.RelPerm.Nw, c.RelPerm.No,
		c.Fluid.MuW, c.Fluid.MuO)
	if err != nil {
		return nil, err
	}
	base, err := FindTangent(m)
	if err != nil {
		return nil, fmt.Errorf("基准算例无 Welge 切点: %w", err)
	}

	res := &ProbeResult{BaseSwf: base.Swf}

	// 实验 1：只降低水粘度。
	for _, mult := range viscosityMultipliers {
		muW := c.Fluid.MuW * mult
		mm, err := fluid.NewModel(c.Rock.Swc, c.Rock.Sor,
			c.RelPerm.Krw0, c.RelPerm.Kro0,
			c.RelPerm.Nw, c.RelPerm.No,
			muW, c.Fluid.MuO)
		if err != nil {
			return nil, fmt.Errorf("粘度实验 %g× 构造失败: %w", mult, err)
		}
		t, err := FindTangent(mm)
		if err != nil {
			return nil, fmt.Errorf("粘度实验 %g× 无切点: %w", mult, err)
		}
		res.ViscosityScan = append(res.ViscosityScan, ViscosityCase{
			MuW:           muW,
			MobilityRatio: mm.EndpointMobilityRatio(),
			Swf:           t.Swf,
			ShockSpeed:    t.Slope,
		})
	}

	// 实验 2：只增大 Sor（注入端饱和度同步取 1−Sor，保证注入端为端点）。
	for _, mult := range residualMultipliers {
		sor := c.Rock.Sor * mult
		if c.Rock.Swc+sor >= 1 {
			return nil, fmt.Errorf("残余油扫描 %g× 失败：Swc+Sor=%g ≥ 1", mult, c.Rock.Swc+sor)
		}
		mm, err := fluid.NewModel(c.Rock.Swc, sor,
			c.RelPerm.Krw0, c.RelPerm.Kro0,
			c.RelPerm.Nw, c.RelPerm.No,
			c.Fluid.MuW, c.Fluid.MuO)
		if err != nil {
			return nil, fmt.Errorf("残余油实验 %g× 构造失败: %w", mult, err)
		}
		t, err := FindTangent(mm)
		if err != nil {
			return nil, fmt.Errorf("残余油实验 %g× 无切点: %w", mult, err)
		}
		res.ResidualScan = append(res.ResidualScan, ResidualOilCase{
			Sor:         sor,
			TerminalSw:  1 - sor,
			Swf:         t.Swf,
			ShockSpeed:  t.Slope,
		})
	}

	// 实验 3：对称本构（krw0=kro0、μw=μo、nw=no）。
	sym, err := fluid.NewModel(c.Rock.Swc, c.Rock.Sor,
		1, 1, // krw0=kro0=1
		c.RelPerm.Nw, c.RelPerm.Nw, // 指数取相同值
		1, 1) // μw=μo=1
	if err != nil {
		return nil, fmt.Errorf("对称实验构造失败: %w", err)
	}
	st, err := FindTangent(sym)
	if err != nil {
		return nil, fmt.Errorf("对称实验无切点: %w", err)
	}
	res.Symmetry = SymmetryCase{
		FAtMid:    sym.F(sym.MidSaturation()),
		Deviation: sym.SymmetryDeviation(fluid.DefaultGrid),
		Swf:       st.Swf,
	}
	return res, nil
}
