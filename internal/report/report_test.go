package report

import (
	"strings"
	"testing"
	"unicode/utf8"

	"buckley-lev/internal/fluid"
	"buckley-lev/internal/schema"
	"buckley-lev/internal/welge"
)

func exampleResult(t *testing.T) *welge.Result {
	t.Helper()
	p := 0.25
	c := &schema.Case{
		Rock:      schema.RockSpec{Swc: 0.2, Sor: 0.2, Porosity: &p},
		RelPerm:   schema.RelPermSpec{Krw0: 0.4, Kro0: 0.9, Nw: 2, No: 2},
		Fluid:     schema.FluidSpec{MuW: 1, MuO: 2},
		Injection: schema.InjectionSpec{SwInj: 0.8},
	}
	res, err := welge.Solve(c, nil)
	if err != nil {
		t.Fatalf("基准算例求解失败: %v", err)
	}
	return res
}

func TestRenderProfileContainsKeyNumbers(t *testing.T) {
	res := exampleResult(t)
	out := RenderHeader(res) + "\n" + RenderWelge(res) + "\n" + RenderProfile(res)
	for _, want := range []string{
		"0.636564",  // Swf
		"1.978640",  // 激波速度 ξs
		"0.5054",    // 突破 PV
		"激波前缘",
		"稀疏波",
		"原始油藏",
		"0.636564",  // 表内重复出现也覆盖
	} {
		if !strings.Contains(out, want) {
			t.Errorf("渲染输出缺少 %q", want)
		}
	}
}

func TestRenderProfileTableAligned(t *testing.T) {
	res := exampleResult(t)
	out := RenderProfile(res)
	lines := strings.Split(out, "\n")
	headerIdx := -1
	for i, ln := range lines {
		if strings.HasPrefix(ln, "ξ") && strings.Contains(ln, "Sw") {
			headerIdx = i
			break
		}
	}
	if headerIdx < 0 {
		t.Fatal("找不到剖面表头")
	}
	sep := lines[headerIdx+1]
	if !strings.Contains(sep, "----") {
		t.Errorf("表头下方应有分隔线，得到 %q", sep)
	}
	if utf8.RuneCountInString(lines[headerIdx]) != utf8.RuneCountInString(sep) {
		t.Errorf("表头字符数 %d 与分隔线字符数 %d 不一致",
			utf8.RuneCountInString(lines[headerIdx]), utf8.RuneCountInString(sep))
	}
}

func TestSaturationChartHasShockColumn(t *testing.T) {
	res := exampleResult(t)
	out := SaturationChart(res, 60, 16)
	for _, ch := range []string{"#", "|", "@", "."} {
		if !strings.Contains(out, ch) {
			t.Errorf("ASCII 剖面图缺少字符 %q", ch)
		}
	}
	if !strings.Contains(out, "ξ=ξs") {
		t.Errorf("ASCII 剖面图缺少 ξ=ξs 标注")
	}
	if strings.Contains(out, "(图宽/高过小") {
		t.Errorf("默认尺寸不应触发过小提示")
	}
}

func TestRenderFractionalShowsEndpointOne(t *testing.T) {
	m, err := fluid.NewModel(0.2, 0.2, 0.4, 0.9, 2, 2, 1, 2)
	if err != nil {
		t.Fatalf("NewModel 失败: %v", err)
	}
	out := RenderFractional(m, 10)
	if !strings.Contains(out, "f(1−Sor)=1.000000") {
		t.Errorf("分流量表应显示端点 f(1−Sor)=1，得到:\n%s", out)
	}
	if !strings.Contains(out, "单调性: OK") {
		t.Errorf("分流量表应显示单调性 OK，得到:\n%s", out)
	}
}

func TestRenderCheckIssuesListed(t *testing.T) {
	c := &schema.Case{
		Rock:      schema.RockSpec{Swc: 0.8, Sor: 0.5}, // Swc+Sor ≥ 1
		RelPerm:   schema.RelPermSpec{Krw0: 0.4, Kro0: 0.9, Nw: 2, No: 2},
		Fluid:     schema.FluidSpec{MuW: 1, MuO: 2},
		Injection: schema.InjectionSpec{SwInj: 0.9},
	}
	out := RenderCheck(c)
	if !strings.Contains(out, "输入校验未通过") {
		t.Errorf("check 输出应说明校验未通过，得到:\n%s", out)
	}
	if !strings.Contains(out, "无可动油带") {
		t.Errorf("check 输出应列出具体问题，得到:\n%s", out)
	}
}

func TestRenderCheckPasses(t *testing.T) {
	p := 0.25
	c := &schema.Case{
		Rock:      schema.RockSpec{Swc: 0.2, Sor: 0.2, Porosity: &p},
		RelPerm:   schema.RelPermSpec{Krw0: 0.4, Kro0: 0.9, Nw: 2, No: 2},
		Fluid:     schema.FluidSpec{MuW: 1, MuO: 2},
		Injection: schema.InjectionSpec{SwInj: 0.8},
	}
	out := RenderCheck(c)
	if !strings.Contains(out, "OK: 输入校验通过") {
		t.Errorf("合法算例 check 应通过，得到:\n%s", out)
	}
}

func TestRenderPVAnalysis(t *testing.T) {
	res := exampleResult(t)
	an, err := welge.AnalyzePV(res.Model, res.Tangent, res.Case.Injection.SwInj, []float64{0.4, 1.0})
	if err != nil {
		t.Fatalf("AnalyzePV 失败: %v", err)
	}
	out := RenderPVAnalysis(an)
	for _, want := range []string{"PV_bt", "0.5054", "突破前", "突破后"} {
		if !strings.Contains(out, want) {
			t.Errorf("物质平衡渲染缺少 %q", want)
		}
	}
	if !strings.Contains(out, "66.7%") {
		t.Errorf("物质平衡渲染应显示突破前采出率，得到:\n%s", out)
	}
}

func TestRenderSweepTrend(t *testing.T) {
	res := exampleResult(t)
	sr, err := welge.Sweep(res.Case, welge.ParamMuW, 0.25, 1.0, 2)
	if err != nil {
		t.Fatalf("Sweep 失败: %v", err)
	}
	out := RenderSweep(sr)
	if !strings.Contains(out, "趋势") {
		t.Errorf("扫描渲染应含趋势摘要，得到:\n%s", out)
	}
	if !strings.Contains(out, "mu_w") {
		t.Errorf("扫描渲染应含参数名，得到:\n%s", out)
	}
}
