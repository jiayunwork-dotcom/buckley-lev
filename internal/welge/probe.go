package welge

import (
	"fmt"

	"buckley-lev/internal/fluid"
	"buckley-lev/internal/schema"
)

type ViscosityCase struct {
	MuW           float64
	MobilityRatio float64
	Swf           float64
	ShockSpeed    float64
}

type ResidualOilCase struct {
	Sor        float64
	TerminalSw float64
	Swf        float64
	ShockSpeed float64
}

type SymmetryCase struct {
	FAtMid    float64
	Deviation float64
	Swf       float64
}

type ProbeResult struct {
	BaseSwf       float64
	ViscosityScan []ViscosityCase
	ResidualScan  []ResidualOilCase
	Symmetry      SymmetryCase
}

var viscosityMultipliers = []float64{1.0, 0.5, 0.25, 0.125}

var residualMultipliers = []float64{1.0, 1.5, 2.0}

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
			Sor:        sor,
			TerminalSw: 1 - sor,
			Swf:        t.Swf,
			ShockSpeed: t.Slope,
		})
	}

	sym, err := fluid.NewModel(c.Rock.Swc, c.Rock.Sor,
		1, 1,
		c.RelPerm.Nw, c.RelPerm.Nw,
		1, 1)
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
