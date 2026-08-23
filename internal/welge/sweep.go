package welge

import (
	"fmt"

	"buckley-lev/internal/fluid"
	"buckley-lev/internal/schema"
)

// SweepParam 标识可扫描的物理参数。
type SweepParam string

// 支持的扫描参数。参数名与算例 JSON 字段一致，便于用户对照。
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

// SweepParams 返回全部可扫描参数名。
func SweepParams() []SweepParam {
	return []SweepParam{ParamMuW, ParamMuO, ParamSor, ParamSwc, ParamKrw0, ParamKro0, ParamNw, ParamNo}
}

// SweepEntry 是扫描中的一行：参数取某值时的切点结果。
type SweepEntry struct {
	// ParamValue 是本次参数取值。
	ParamValue float64
	// Swf 是重解后的前缘饱和度。
	Swf float64
	// ShockSpeed 是无因次激波速度。
	ShockSpeed float64
	// MobilityRatio 是端点流度比。
	MobilityRatio float64
	// TerminalSw 是注入端饱和度（Sor 变化时为 1−Sor，否则不变）。
	TerminalSw float64
}

// SweepResult 是一次参数扫描的完整结果。
type SweepResult struct {
	// Param 是被扫描的参数。
	Param SweepParam
	// From、To、Steps 是扫描区间与步数。
	From, To float64
	Steps    int
	// Entries 是按参数值升序排列的扫描行。
	Entries []SweepEntry
}

// validParamValue 校验参数取值是否在合法区间，避免扫描中途撞上
// 非法本构（例如 swc+sor≥1）。
func validParamValue(p SweepParam, base *schema.Case, v float64) bool {
	c := cloneCase(base)
	switch p {
	case ParamMuW:
		c.Fluid.MuW = v
	case ParamMuO:
		c.Fluid.MuO = v
	case ParamSor:
		c.Rock.Sor = v
		c.Injection.SwInj = 1 - v // 注入端同步取端点，保证剖面完备
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

// Sweep 对单个参数在 [from, to] 上均匀取 Steps+1 个点，每个点
// 重新求解 Welge 切点，输出 Swf 与激波速度随参数的变化。
//
// 典型用途：
//   - sweep --param mu_w --from 0.1 --to 2  观察降低水粘度时
//     Swf 单调下降、激波速度上升；
//   - sweep --param sor  --from 0.2 --to 0.5 观察末端饱和度左移。
//
// 扫描区间内任何一步构造/求解失败都会返回带步进序号的错误。
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
	// 区间端点先做合法性预检，给出清晰错误而不是在中途爆出。
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
			v = to // 避免浮点累加误差。
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
			ParamValue:   v,
			Swf:          t.Swf,
			ShockSpeed:   t.Slope,
			MobilityRatio: m.EndpointMobilityRatio(),
			TerminalSw:   1 - cc.Rock.Sor,
		})
	}
	return res, nil
}

// cloneCase 深度复制一个算例，避免扫描过程互相污染。
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
