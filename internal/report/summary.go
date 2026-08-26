package report

import (
	"fmt"
	"strings"

	"buckley-lev/internal/schema"
	"buckley-lev/internal/welge"
)

func RenderProbe(res *welge.ProbeResult) string {
	var b strings.Builder

	b.WriteString("交叉规则探针\n")
	fmt.Fprintf(&b, "基准前缘饱和度 Swf = %s\n\n", Num(res.BaseSwf, 6))

	b.WriteString("实验 1：只降低水粘度（μo 不变）——水相更易流动，前缘饱和度下降、前缘变尖\n")
	tbl := NewTable("μw", "端点流度比 M", "Swf", "激波速度 ξs")
	for _, c := range res.ViscosityScan {
		tbl.AddRow(Num(c.MuW, 4), Num(c.MobilityRatio, 4), Num(c.Swf, 6), Num(c.ShockSpeed, 6))
	}
	b.WriteString(tbl.String())

	b.WriteString("\n实验 2：只增大 Sor——可动油变少，末端饱和度左移\n")
	tbl2 := NewTable("Sor", "1−Sor", "Swf", "激波速度 ξs")
	for _, c := range res.ResidualScan {
		tbl2.AddRow(Num(c.Sor, 4), Num(c.TerminalSw, 4), Num(c.Swf, 6), Num(c.ShockSpeed, 6))
	}
	b.WriteString(tbl2.String())

	b.WriteString("\n实验 3：对称本构（krw0=kro0，μw=μo，nw=no）\n")
	fmt.Fprintf(&b, "  f(可动油带中点) = %s（应为 0.5）\n", Num(res.Symmetry.FAtMid, 6))
	fmt.Fprintf(&b, "  对称偏差 max|f(x)+f(1−x)−1| = %.4g（应接近 0）\n", res.Symmetry.Deviation)
	fmt.Fprintf(&b, "  切点饱和度 Swf = %s\n", Num(res.Symmetry.Swf, 6))
	return b.String()
}

func RenderCheck(c *schema.Case) string {
	schema.ApplyDefaults(c)
	err := schema.Validate(c)
	if err == nil {
		return "OK: 输入校验通过（未计算）\n"
	}
	var b strings.Builder
	b.WriteString("输入校验未通过：\n")
	if ve, ok := err.(*schema.ValidationError); ok {
		for _, issue := range ve.Issues {
			b.WriteString("  - " + issue + "\n")
		}
	} else {
		b.WriteString("  - " + err.Error() + "\n")
	}
	return b.String()
}
