package report

import (
	"fmt"
	"strings"

	"buckley-lev/internal/welge"
)

// RenderPVAnalysis 输出 Welge 物质平衡（注入孔隙体积历史）。
func RenderPVAnalysis(an *welge.PVAnalysis) string {
	var b strings.Builder
	b.WriteString("Welge 物质平衡（注入孔隙体积历史）\n")
	fmt.Fprintf(&b, "  突破时刻 PV_bt = %s（激波速度 ξs = %s）\n",
		Num(an.BreakthroughPV, 4), Num(1/an.BreakthroughPV, 6))
	tbl := NewTable("PV", "出口 Sw", "平均 Sw", "采出率", "状态")
	for _, e := range an.Entries {
		state := "突破后"
		if e.BeforeBreakthrough {
			state = "突破前"
		}
		tbl.AddRow(
			Num(e.PV, 2), Num(e.OutletSw, 4), Num(e.AverageSw, 4),
			fmt.Sprintf("%.1f%%", e.OilRecovery*100), state,
		)
	}
	b.WriteString(tbl.String())
	return b.String()
}

// RenderSweep 输出参数扫描结果。
func RenderSweep(res *welge.SweepResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "参数扫描：%s ∈ [%s, %s]（%d 步）\n",
		res.Param, Num(res.From, 4), Num(res.To, 4), res.Steps)
	tbl := NewTable(string(res.Param), "Swf", "激波速度 ξs", "流度比 M", "末端 1−Sor")
	for _, e := range res.Entries {
		tbl.AddRow(
			Num(e.ParamValue, 4), Num(e.Swf, 6), Num(e.ShockSpeed, 6),
			Num(e.MobilityRatio, 4), Num(e.TerminalSw, 4),
		)
	}
	b.WriteString(tbl.String())

	// 趋势摘要：首尾对比，便于快速阅读交叉规则方向。
	if len(res.Entries) >= 2 {
		first, last := res.Entries[0], res.Entries[len(res.Entries)-1]
		var dirSwf, dirSpeed string
		switch {
		case last.Swf > first.Swf+1e-9:
			dirSwf = "上升"
		case last.Swf < first.Swf-1e-9:
			dirSwf = "下降"
		default:
			dirSwf = "基本不变"
		}
		switch {
		case last.ShockSpeed > first.ShockSpeed+1e-9:
			dirSpeed = "上升"
		case last.ShockSpeed < first.ShockSpeed-1e-9:
			dirSpeed = "下降"
		default:
			dirSpeed = "基本不变"
		}
		fmt.Fprintf(&b, "趋势：%s 从 %s 到 %s → Swf %s，激波速度 %s\n",
			res.Param, Num(first.ParamValue, 4), Num(last.ParamValue, 4), dirSwf, dirSpeed)
	}
	return b.String()
}
