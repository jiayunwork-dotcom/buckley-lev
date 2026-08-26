package report

import (
	"fmt"
	"strings"

	"buckley-lev/internal/fluid"
	"buckley-lev/internal/welge"
)

func RenderHeader(res *welge.Result) string {
	c := res.Case
	var b strings.Builder
	fmt.Fprintf(&b, "算例参数\n")
	fmt.Fprintf(&b, "  Swc = %s   Sor = %s   可动油带 Δ = %s\n",
		Num(c.Rock.Swc, 4), Num(c.Rock.Sor, 4), Num(res.Model.Delta, 4))
	fmt.Fprintf(&b, "  Corey: krw = %.3g·S*^%.3g    kro = %.3g·(1−S*)^%.3g\n",
		c.RelPerm.Krw0, c.RelPerm.Nw, c.RelPerm.Kro0, c.RelPerm.No)
	fmt.Fprintf(&b, "  粘度: μw = %s   μo = %s   端点流度比 M = %.4f\n",
		Num(c.Fluid.MuW, 4), Num(c.Fluid.MuO, 4), res.Model.EndpointMobilityRatio())
	fmt.Fprintf(&b, "  注入端: Sw_inj = %s（1−Sor = %s）\n",
		Num(c.Injection.SwInj, 4), Num(1-c.Rock.Sor, 4))
	if c.Rock.Porosity != nil {
		fmt.Fprintf(&b, "  孔隙度: φ = %s\n", Num(*c.Rock.Porosity, 4))
	}
	return b.String()
}

func RenderWelge(res *welge.Result) string {
	t := res.Tangent
	s := res.Shock
	var b strings.Builder
	b.WriteString("Welge 切点与激波\n")
	fmt.Fprintf(&b, "  前缘饱和度 Swf        = %s\n", Num(t.Swf, 6))
	fmt.Fprintf(&b, "  切点局部斜率 f'(Swf)  = %s\n", Num(t.LocalSlope, 6))
	fmt.Fprintf(&b, "  割线斜率 f(Swf)/(Swf−Swc) = %s\n", Num(t.Slope, 6))
	fmt.Fprintf(&b, "  切线自洽偏差          = %.3g\n", t.LocalSlope-t.Slope)
	fmt.Fprintf(&b, "  无因次激波速度 ξs     = %s\n", Num(s.Speed, 6))
	fmt.Fprintf(&b, "  Rankine–Hugoniot      = (f(Swf)−f(Swc))/(Swf−Swc) = %s\n", Num(s.Speed, 6))
	fmt.Fprintf(&b, "  突破注入孔隙体积      = %s PV\n", Num(s.BreakthroughPV, 4))
	return b.String()
}

func RenderProfile(res *welge.Result) string {
	var b strings.Builder
	b.WriteString("无因次饱和度剖面（ξ=x/t）\n")
	tbl := NewTable("ξ", "Sw", "类别")
	for _, p := range res.Profile.Points {
		tbl.AddRow(Num(p.Xi, 6), Num(p.Sw, 6), string(p.Kind))
	}
	b.WriteString(tbl.String())
	return b.String()
}

func RenderFractional(m *fluid.Model, grid int) string {
	var b strings.Builder
	b.WriteString("分流函数 f(Sw) 采样\n")
	tbl := NewTable("Sw", "S*", "krw", "kro", "λw", "λo", "f", "f'")
	lo, hi := m.Domain()
	for i := 0; i <= grid; i++ {
		sw := lo + (hi-lo)*float64(i)/float64(grid)
		x := m.Normalized(sw)
		tbl.AddRow(
			Num(sw, 4), Num(x, 4),
			Num(m.Krw(sw), 4), Num(m.Kro(sw), 4),
			Num(m.LambdaW(sw), 4), Num(m.LambdaO(sw), 4),
			Num(m.F(sw), 4), Num(m.FPrime(sw), 4),
		)
	}
	b.WriteString(tbl.String())
	mono := m.CheckMonotonic(fluid.DefaultGrid)
	if mono.OK {
		fmt.Fprintf(&b, "单调性: OK（最小 f' = %s，在 Sw=%s）\n", Num(mono.MinSlope, 4), Num(mono.AtSw, 4))
	} else {
		fmt.Fprintf(&b, "单调性: 异常（最小 f' = %s，在 Sw=%s）\n", Num(mono.MinSlope, 4), Num(mono.AtSw, 4))
	}
	fmt.Fprintf(&b, "端点: f(Swc)=%s   f(1−Sor)=%s\n", Num(m.F(m.Swc), 6), Num(m.F(1-m.Sor), 6))
	return b.String()
}
