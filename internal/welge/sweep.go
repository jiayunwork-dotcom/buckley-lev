package welge

import (
	"fmt"

	"buckley-lev/internal/fluid"
	"buckley-lev/internal/schema"
)

type SweepParam string

const (
	ParamMuW  SweepParam = "mu_w"
	ParamMuO  SweepParam = "mu_o"
	ParamSor  SweepParam = "sor"
	ParamSwc  SweepParam = "swc"
	ParamKrw0 SweepParam = "krw0"
	ParamKro0 SweepParam = "kro0"
	ParamNw   SweepParam = "nw"
	ParamNo   SweepParam = "no"
)

func SweepParams() []SweepParam {
	return []SweepParam{ParamMuW, ParamMuO, ParamSor, ParamSwc, ParamKrw0, ParamKro0, ParamNw, ParamNo}
}

type SweepEntry struct {
	ParamValue    float64
	Swf           float64
	ShockSpeed    float64
	MobilityRatio float64
	TerminalSw    float64
}

type SweepResult struct {
	Param    SweepParam
	From, To float64
	Steps    int
	Entries  []SweepEntry
}

func validParamValue(p SweepParam, base *schema.Case, v float64) bool {
	c := cloneCase(base)
	switch p {
	case ParamMuW:
		c.Fluid.MuW = v
	case ParamMuO:
		c.Fluid.MuO = v
	case ParamSor:
		c.Rock.Sor = v
		c.Injection.SwInj = 1 - v
	case ParamSwc:
		c.Rock.Swc = v
	case ParamKrw0:
		c.RelPerm.Krw0 = v
	case ParamKro0:
		c.RelPerm.Kro0 = v
	case ParamNw:
		c.RelPerm.Nw = v
	case ParamNo:
		c.RelPerm.No = v
	}
	return schema.Validate(c) == nil
}

func Sweep(c *schema.Case, param SweepParam, from, to float64, steps int) (*SweepResult, error) {
	if steps < 1 {
		return nil, fmt.Errorf("扫描步数必须 ≥ 1，得到 %d", steps)
	}
	supported := false
	for _, p := range SweepParams() {
		if p == param {
			supported = true
			break
		}
	}
	if !supported {
		return nil, fmt.Errorf("不支持的扫描参数 %q（可用：%v）", param, SweepParams())
	}
	if to < from {
		return nil, fmt.Errorf("扫描区间非法：to=%g < from=%g", to, from)
	}
	base := cloneCase(c)
	if !validParamValue(param, base, from) {
		return nil, fmt.Errorf("参数 %s 取 %g 时本构非法（先检查区间端点）", param, from)
	}
	if !validParamValue(param, base, to) {
		return nil, fmt.Errorf("参数 %s 取 %g 时本构非法（先检查区间端点）", param, to)
	}

	res := &SweepResult{Param: param, From: from, To: to, Steps: steps}
	step := (to - from) / float64(steps)
	for i := 0; i <= steps; i++ {
		v := from + step*float64(i)
		if i == steps {
			v = to
		}
		cc := cloneCase(base)
		switch param {
		case ParamMuW:
			cc.Fluid.MuW = v
		case ParamMuO:
			cc.Fluid.MuO = v
		case ParamSor:
			cc.Rock.Sor = v
			cc.Injection.SwInj = 1 - v
		case ParamSwc:
			cc.Rock.Swc = v
		case ParamKrw0:
			cc.RelPerm.Krw0 = v
		case ParamKro0:
			cc.RelPerm.Kro0 = v
		case ParamNw:
			cc.RelPerm.Nw = v
		case ParamNo:
			cc.RelPerm.No = v
		}
		schema.ApplyDefaults(cc)
		if err := schema.Validate(cc); err != nil {
			return nil, fmt.Errorf("扫描第 %d 步（%s=%g）非法: %w", i, param, v, err)
		}
		m, err := fluid.NewModel(cc.Rock.Swc, cc.Rock.Sor,
			cc.RelPerm.Krw0, cc.RelPerm.Kro0,
			cc.RelPerm.Nw, cc.RelPerm.No,
			cc.Fluid.MuW, cc.Fluid.MuO)
		if err != nil {
			return nil, fmt.Errorf("扫描第 %d 步（%s=%g）构造失败: %w", i, param, v, err)
		}
		t, err := FindTangent(m)
		if err != nil {
			return nil, fmt.Errorf("扫描第 %d 步（%s=%g）无 Welge 切点: %w", i, param, v, err)
		}
		res.Entries = append(res.Entries, SweepEntry{
			ParamValue:    v,
			Swf:           t.Swf,
			ShockSpeed:    t.Slope,
			MobilityRatio: m.EndpointMobilityRatio(),
			TerminalSw:    1 - cc.Rock.Sor,
		})
	}
	return res, nil
}

func cloneCase(c *schema.Case) *schema.Case {
	out := *c
	out.Rock = c.Rock
	out.RelPerm = c.RelPerm
	out.Fluid = c.Fluid
	out.Injection = c.Injection
	if c.Rock.Porosity != nil {
		p := *c.Rock.Porosity
		out.Rock.Porosity = &p
	}
	return &out
}
