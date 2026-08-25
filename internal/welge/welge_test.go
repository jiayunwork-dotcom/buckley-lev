package welge

import (
	"math"
	"strings"
	"testing"

	"buckley-lev/internal/fluid"
	"buckley-lev/internal/schema"
)

func exampleCase() *schema.Case {
	p := 0.25
	return &schema.Case{
		Rock:      schema.RockSpec{Swc: 0.2, Sor: 0.2, Porosity: &p},
		RelPerm:   schema.RelPermSpec{Krw0: 0.4, Kro0: 0.9, Nw: 2, No: 2},
		Fluid:     schema.FluidSpec{MuW: 1, MuO: 2},
		Injection: schema.InjectionSpec{SwInj: 0.8},
	}
}

func solveExample(t *testing.T) *Result {
	t.Helper()
	res, err := Solve(exampleCase(), nil)
	if err != nil {
		t.Fatalf("基准算例求解失败: %v", err)
	}
	return res
}

func TestWelgeTangentSlope(t *testing.T) {
	cases := []struct {
		name string
		mut  func(*schema.Case)
	}{
		{"基准", func(c *schema.Case) {}},
		{"对称本构", func(c *schema.Case) {
			c.RelPerm.Krw0, c.RelPerm.Kro0 = 1, 1
			c.Fluid.MuW, c.Fluid.MuO = 1, 1
		}},
		{"强粘差", func(c *schema.Case) { c.Fluid.MuW = 0.2 }},
		{"高指数", func(c *schema.Case) { c.RelPerm.Nw, c.RelPerm.No = 3.5, 2.5 }},
	}
	for _, c := range cases {
		cc := exampleCase()
		c.mut(cc)
		res, err := Solve(cc, nil)
		if err != nil {
			t.Errorf("%s: 求解失败: %v", c.name, err)
			continue
		}
		tk := res.Tangent
		scale := math.Max(1, math.Abs(tk.Slope))
		if d := math.Abs(tk.LocalSlope - tk.Slope); d > fluid.RelTol*scale {
			t.Errorf("%s: 切点斜率 f'(Swf)=%g 与割线 f(Swf)/(Swf−Swc)=%g 偏差 %g 超容差",
				c.name, tk.LocalSlope, tk.Slope, d)
		}
		if tk.Swf <= res.Model.Swc || tk.Swf >= 1-res.Model.Sor {
			t.Errorf("%s: Swf=%g 不在 (Swc, 1−Sor)=(%g, %g) 内", c.name, tk.Swf, res.Model.Swc, 1-res.Model.Sor)
		}
		if tk.Slope <= 0 {
			t.Errorf("%s: 割线斜率应 > 0，得到 %g", c.name, tk.Slope)
		}
		if fpp := res.Model.FDoublePrime(tk.Swf); fpp >= 0 {
			t.Errorf("%s: 切点应位于凹段，f''(Swf)=%g 应 < 0", c.name, fpp)
		}
	}
}

func TestShockRankineHugoniot(t *testing.T) {
	res := solveExample(t)
	m := res.Model
	s := res.Shock
	expected := (m.F(s.UpstreamSw) - m.F(s.DownstreamSw)) / (s.UpstreamSw - s.DownstreamSw)
	if math.Abs(s.Speed-expected) > 1e-12 {
		t.Errorf("激波速度 %g 不满足 Rankine–Hugoniot %g", s.Speed, expected)
	}
	if s.DownstreamSw != m.Swc {
		t.Errorf("激波下游饱和度应为 Swc=%g，得到 %g", m.Swc, s.DownstreamSw)
	}
	if math.Abs(s.DownstreamF) > 1e-12 {
		t.Errorf("激波下游分流量 f(Swc) 应为 0，得到 %g", s.DownstreamF)
	}
}

func TestRarefactionCurve(t *testing.T) {
	res := solveExample(t)
	m := res.Model
	swInj := res.Case.Injection.SwInj
	prof := res.Profile

	if len(prof.Points) < 4 {
		t.Fatalf("剖面采样点过少: %d", len(prof.Points))
	}

	interior := 0
	maxDev := 0.0
	for _, p := range prof.Points {
		if p.Kind != KindRarefaction {
			continue
		}
		interior++
		slope := m.FPrime(p.Sw)
		scale := math.Max(1, math.Abs(p.Xi))
		if d := math.Abs(slope - p.Xi); d > 1e-6*scale {
			t.Errorf("稀疏波点 (ξ=%g, Sw=%g): f'(Sw)=%g 与 ξ 偏差 %g", p.Xi, p.Sw, slope, d)
		}
		lin := swInj + (prof.SwFront-swInj)*(p.Xi/prof.XiShock)
		if d := math.Abs(p.Sw - lin); d > maxDev {
			maxDev = d
		}
	}
	if interior < 3 {
		t.Fatalf("稀疏波点不足: %d", interior)
	}
	if maxDev < 5e-3 {
		t.Errorf("稀疏波与直线插值的最大偏差 %g 过小——曲线几乎就是一条斜直线", maxDev)
	}
}

func TestProfileShockFrontUsesWelgeTangent(t *testing.T) {
	res := solveExample(t)
	m := res.Model
	tk := res.Tangent

	var front *Point
	for i := range res.Profile.Points {
		if res.Profile.Points[i].Kind == KindShock {
			front = &res.Profile.Points[i]
			break
		}
	}
	if front == nil {
		t.Fatal("剖面中找不到激波前缘点")
	}
	if math.Abs(front.Xi-res.Shock.Speed) > 1e-9 {
		t.Errorf("激波前缘 ξ=%g 不等于 Welge 切线速度 ξs=%g", front.Xi, res.Shock.Speed)
	}
	if math.Abs(front.Sw-tk.Swf) > 1e-9 {
		t.Errorf("激波前缘饱和度 %g 不等于 Swf=%g", front.Sw, tk.Swf)
	}

	endSlope := m.FPrime(res.Case.Injection.SwInj)
	if endSlope > res.Shock.Speed-1e-9 {
		t.Errorf("注入端局部斜率 f'(Sw_inj)=%g 不应 ≥ 激波速度 %g", endSlope, res.Shock.Speed)
	}
	if res.Shock.Speed < 1 {
		t.Errorf("激波速度 %g 明显偏小，疑似误用局部斜率而非 Welge 切线", res.Shock.Speed)
	}
	maxSlope := 0.0
	for i := 0; i < 513; i++ {
		sw := m.Swc + (1-m.Sor-m.Swc)*float64(i)/512
		if s := m.FPrime(sw); s > maxSlope {
			maxSlope = s
		}
	}
	if math.Abs(maxSlope-res.Shock.Speed) < 0.1 {
		t.Errorf("最大局部斜率 %g 与切线速度 %g 过于接近，无法区分——测试不具判别力", maxSlope, res.Shock.Speed)
	}
}

func TestCrossRuleLowerWaterViscosity(t *testing.T) {
	pr, err := ProbeCrossRules(exampleCase())
	if err != nil {
		t.Fatalf("探针失败: %v", err)
	}
	if len(pr.ViscosityScan) < 3 {
		t.Fatalf("粘度扫描点数不足: %d", len(pr.ViscosityScan))
	}
	if math.Abs(pr.ViscosityScan[0].Swf-pr.BaseSwf) > 1e-12 {
		t.Errorf("扫描首点应等于基准 Swf，%g != %g", pr.ViscosityScan[0].Swf, pr.BaseSwf)
	}
	for i := 1; i < len(pr.ViscosityScan); i++ {
		if pr.ViscosityScan[i].Swf >= pr.ViscosityScan[i-1].Swf {
			t.Errorf("μw %g→%g: Swf %g→%g 应下降",
				pr.ViscosityScan[i-1].MuW, pr.ViscosityScan[i].MuW,
				pr.ViscosityScan[i-1].Swf, pr.ViscosityScan[i].Swf)
		}
		if pr.ViscosityScan[i].ShockSpeed <= pr.ViscosityScan[i-1].ShockSpeed {
			t.Errorf("μw %g→%g: 激波速度 %g→%g 应上升",
				pr.ViscosityScan[i-1].MuW, pr.ViscosityScan[i].MuW,
				pr.ViscosityScan[i-1].ShockSpeed, pr.ViscosityScan[i].ShockSpeed)
		}
	}
	for i := 1; i < len(pr.ViscosityScan); i++ {
		if pr.ViscosityScan[i].MobilityRatio <= pr.ViscosityScan[i-1].MobilityRatio {
			t.Errorf("μw %g→%g: 端点流度比应上升，%g→%g",
				pr.ViscosityScan[i-1].MuW, pr.ViscosityScan[i].MuW,
				pr.ViscosityScan[i-1].MobilityRatio, pr.ViscosityScan[i].MobilityRatio)
		}
	}
}

func TestCrossRuleResidualOilLeftShift(t *testing.T) {
	pr, err := ProbeCrossRules(exampleCase())
	if err != nil {
		t.Fatalf("探针失败: %v", err)
	}
	if len(pr.ResidualScan) < 3 {
		t.Fatalf("残余油扫描点数不足: %d", len(pr.ResidualScan))
	}
	for _, c := range pr.ResidualScan {
		if math.Abs(c.TerminalSw-(1-c.Sor)) > 1e-12 {
			t.Errorf("末端饱和度应为 1−Sor=%g，得到 %g", 1-c.Sor, c.TerminalSw)
		}
	}
	for i := 1; i < len(pr.ResidualScan); i++ {
		if pr.ResidualScan[i].TerminalSw >= pr.ResidualScan[i-1].TerminalSw {
			t.Errorf("Sor %g→%g: 末端饱和度 %g→%g 应左移（变小）",
				pr.ResidualScan[i-1].Sor, pr.ResidualScan[i].Sor,
				pr.ResidualScan[i-1].TerminalSw, pr.ResidualScan[i].TerminalSw)
		}
		if pr.ResidualScan[i].Swf >= pr.ResidualScan[i-1].Swf {
			t.Errorf("Sor %g→%g: Swf %g→%g 应下降",
				pr.ResidualScan[i-1].Sor, pr.ResidualScan[i].Sor,
				pr.ResidualScan[i-1].Swf, pr.ResidualScan[i].Swf)
		}
	}
}

func TestCrossRuleSymmetryProbe(t *testing.T) {
	pr, err := ProbeCrossRules(exampleCase())
	if err != nil {
		t.Fatalf("探针失败: %v", err)
	}
	if math.Abs(pr.Symmetry.FAtMid-0.5) > 1e-12 {
		t.Errorf("对称本构 f(中点)=%g 应为 0.5", pr.Symmetry.FAtMid)
	}
	if pr.Symmetry.Deviation > 1e-12 {
		t.Errorf("对称本构对称偏差 %g 应接近 0", pr.Symmetry.Deviation)
	}
}

func TestBreakthroughPoreVolume(t *testing.T) {
	res := solveExample(t)
	want := 0.5053975515463918
	if math.Abs(res.Shock.BreakthroughPV-want) > 1e-6 {
		t.Errorf("突破注入孔隙体积 %g 应为 %g", res.Shock.BreakthroughPV, want)
	}
	if math.Abs(res.Shock.BreakthroughPV*res.Shock.Speed-1) > 1e-12 {
		t.Errorf("突破 PV × 激波速度应 = 1，得到 %g", res.Shock.BreakthroughPV*res.Shock.Speed)
	}
}

func TestProfileRejectsInjectionBelowFront(t *testing.T) {
	c := exampleCase()
	c.Injection.SwInj = 0.5
	_, err := Solve(c, nil)
	if err == nil {
		t.Fatal("注入饱和度低于前缘饱和度时应报错（Sw 越界）")
	}
	if !strings.Contains(err.Error(), "Sw 越界") {
		t.Errorf("错误应说明 Sw 越界，得到: %v", err)
	}
}

func TestSolveRejectsInvalidCase(t *testing.T) {
	c := exampleCase()
	c.Rock.Swc = 0.7
	c.Rock.Sor = 0.4
	if _, err := Solve(c, nil); err == nil {
		t.Fatal("非法算例应求解失败")
	}
}

func TestParseFractions(t *testing.T) {
	ok := []struct {
		in   string
		want []float64
	}{
		{"0,0.3,1", []float64{0, 0.3, 1}},
		{"1", []float64{1}},
		{"0.5,0.5", []float64{0.5, 0.5}},
	}
	for _, c := range ok {
		got, err := ParseFractions(c.in)
		if err != nil {
			t.Errorf("ParseFractions(%q) 不应报错: %v", c.in, err)
			continue
		}
		if len(got) != len(c.want) {
			t.Errorf("ParseFractions(%q) 长度 %d != %d", c.in, len(got), len(c.want))
			continue
		}
		for i := range got {
			if math.Abs(got[i]-c.want[i]) > 1e-12 {
				t.Errorf("ParseFractions(%q)[%d]=%g != %g", c.in, i, got[i], c.want[i])
			}
		}
	}
	bad := []string{"", "0.5,1.5", "a", "0,,1", "0,0.3,", "2"}
	for _, in := range bad {
		if _, err := ParseFractions(in); err == nil {
			t.Errorf("ParseFractions(%q) 应报错", in)
		}
	}
}

func TestAnalyzePVBreakthrough(t *testing.T) {
	res := solveExample(t)
	an, err := AnalyzePV(res.Model, res.Tangent, res.Case.Injection.SwInj, []float64{0.4, 0.6, 1.0})
	if err != nil {
		t.Fatalf("AnalyzePV 失败: %v", err)
	}
	if math.Abs(an.BreakthroughPV-0.5053975515) > 1e-6 {
		t.Errorf("突破 PV 应为 %g，得到 %g", 0.5053975515, an.BreakthroughPV)
	}
	if an.Entries[0].BeforeBreakthrough != true || math.Abs(an.Entries[0].OutletSw-res.Model.Swc) > 1e-12 {
		t.Errorf("PV=0.4 应处于突破前且出口为 Swc，得到 %+v", an.Entries[0])
	}
	for _, e := range an.Entries[1:] {
		if e.BeforeBreakthrough {
			continue
		}
		if d := math.Abs(res.Model.FPrime(e.OutletSw)*e.PV - 1); d > 1e-6 {
			t.Errorf("PV=%g: f'(Sw_e)·PV=%g 应 = 1", e.PV, res.Model.FPrime(e.OutletSw)*e.PV)
		}
		if e.AverageSw < e.OutletSw-Eps {
			t.Errorf("PV=%g: 平均饱和度 %g 应 ≥ 出口饱和度 %g", e.PV, e.AverageSw, e.OutletSw)
		}
	}
	for i := 1; i < len(an.Entries); i++ {
		if an.Entries[i].OilRecovery < an.Entries[i-1].OilRecovery-Eps {
			t.Errorf("采出率应单调不减：PV=%g 采出率 %g < PV=%g 的 %g",
				an.Entries[i].PV, an.Entries[i].OilRecovery, an.Entries[i-1].PV, an.Entries[i-1].OilRecovery)
		}
		if an.Entries[i].OilRecovery > 1+1e-9 || an.Entries[i].OilRecovery < 0 {
			t.Errorf("采出率越界: %g", an.Entries[i].OilRecovery)
		}
	}
}

func TestAnalyzePVAverageAtBreakthrough(t *testing.T) {
	res := solveExample(t)
	an, err := AnalyzePV(res.Model, res.Tangent, res.Case.Injection.SwInj, []float64{0.4, 0.5})
	if err != nil {
		t.Fatalf("AnalyzePV 失败: %v", err)
	}
	tk := res.Tangent
	want := tk.Swf + (1-tk.F)/tk.LocalSlope
	for _, e := range an.Entries {
		if !e.BeforeBreakthrough {
			continue
		}
		if math.Abs(e.AverageSw-want) > 1e-9 {
			t.Errorf("突破前平均饱和度 %g 应为 Welge 切线外推值 %g", e.AverageSw, want)
		}
	}
}

func TestSweepLowerWaterViscosityDirection(t *testing.T) {
	pr, err := Sweep(exampleCase(), ParamMuW, 0.25, 1.0, 3)
	if err != nil {
		t.Fatalf("Sweep 失败: %v", err)
	}
	if len(pr.Entries) != 4 {
		t.Fatalf("扫描应产生 4 行，得到 %d", len(pr.Entries))
	}
	for i := 1; i < len(pr.Entries); i++ {
		if pr.Entries[i].Swf <= pr.Entries[i-1].Swf {
			t.Errorf("μw %g→%g: Swf %g→%g 应上升",
				pr.Entries[i-1].ParamValue, pr.Entries[i].ParamValue,
				pr.Entries[i-1].Swf, pr.Entries[i].Swf)
		}
		if pr.Entries[i].ShockSpeed >= pr.Entries[i-1].ShockSpeed {
			t.Errorf("μw %g→%g: 激波速度 %g→%g 应下降",
				pr.Entries[i-1].ParamValue, pr.Entries[i].ParamValue,
				pr.Entries[i-1].ShockSpeed, pr.Entries[i].ShockSpeed)
		}
	}
}

func TestSweepResidualOilShrinksTerminal(t *testing.T) {
	pr, err := Sweep(exampleCase(), ParamSor, 0.2, 0.4, 2)
	if err != nil {
		t.Fatalf("Sweep 失败: %v", err)
	}
	for _, e := range pr.Entries {
		if math.Abs(e.TerminalSw-(1-e.ParamValue)) > 1e-12 {
			t.Errorf("末端饱和度应为 1−Sor=%g，得到 %g", 1-e.ParamValue, e.TerminalSw)
		}
	}
	for i := 1; i < len(pr.Entries); i++ {
		if pr.Entries[i].TerminalSw >= pr.Entries[i-1].TerminalSw {
			t.Errorf("Sor %g→%g: 末端饱和度应左移",
				pr.Entries[i-1].ParamValue, pr.Entries[i].ParamValue)
		}
	}
}

func TestSweepRejectsBadArguments(t *testing.T) {
	c := exampleCase()
	if _, err := Sweep(c, ParamMuW, 1.0, 0.5, 2); err == nil {
		t.Error("to<from 的扫描区间应报错")
	}
	if _, err := Sweep(c, "bogus", 0, 1, 2); err == nil {
		t.Error("未知扫描参数应报错")
	}
	if _, err := Sweep(c, ParamMuW, 0, 1, 0); err == nil {
		t.Error("步数 0 应报错")
	}
	if _, err := Sweep(c, ParamSor, 0.2, 0.9, 2); err == nil {
		t.Error("区间端点导致本构非法的扫描应报错")
	}
}
