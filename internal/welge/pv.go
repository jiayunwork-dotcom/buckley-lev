package welge

import (
	"fmt"
	"math"

	"buckley-lev/internal/fluid"
)

// PVEntry 是注入孔隙体积历史上的一行：给定 PV 时刻的出口饱和度、
// 平均饱和度与采收率。
type PVEntry struct {
	// PV 是累计注入孔隙体积倍数（无因次时间）。
	PV float64
	// OutletSw 是出口端水饱和度。突破前为 Swc，突破后由
	// f'(Sw_e) = 1/PV 反解。
	OutletSw float64
	// AverageSw 是油藏平均水饱和度（Welge 物质平衡式）。
	AverageSw float64
	// OilRecovery 是采出油占原始可动油的分数。
	OilRecovery float64
	// BeforeBreakthrough 表示该时刻是否在突破前。
	BeforeBreakthrough bool
}

// PVAnalysis 是一次 Welge 物质平衡分析的结果。
type PVAnalysis struct {
	// BreakthroughPV 是突破时刻的注入孔隙体积 = 1/ξs。
	BreakthroughPV float64
	// Swf 是突破前缘饱和度。
	Swf float64
	// Entries 是 PV 序列上的逐时刻结果。
	Entries []PVEntry
}

// DefaultPVGrid 是默认的 PV 采样序列（无因次时间）。
var DefaultPVGrid = []float64{0.4, 0.6, 0.8, 1.0, 1.2, 1.5, 2.0, 3.0, 5.0}

// OutletSaturation 求给定注入 PV 时出口端的水饱和度。
//
// 突破前（PV ≤ 1/ξs）：出口还是原始油藏，Sw_e = Swc。
// 突破后：出口饱和度满足 f'(Sw_e) = 1/PV，在 [Swf, Sw_inj] 上
// 反解（f' 在该区间单调递减）。
func OutletSaturation(m *fluid.Model, t *Tangent, swInj, pv float64) (float64, error) {
	pvBt := 1 / t.Slope
	if pv <= pvBt+Eps {
		return m.Swc, nil
	}
	if pv <= 0 {
		return 0, fmt.Errorf("PV 必须 > 0，得到 %g", pv)
	}
	target := 1 / pv
	if target > m.FPrime(t.Swf) {
		return 0, fmt.Errorf("PV=%g 过小：1/PV=%g 高于前缘局部斜率 f'(Swf)=%g", pv, target, m.FPrime(t.Swf))
	}
	return RarefactionSaturation(m, t, swInj, target)
}

// AverageSaturation 用 Welge 物质平衡式求平均水饱和度：
//
//	S̄ = Sw_e + (1 − f(Sw_e)) / f'(Sw_e)
//
// 其中 Sw_e 是对应 PV 的出口饱和度。突破时刻该式退化为
// S̄ = Swf + (1 − f(Swf))/f'(Swf)。
func AverageSaturation(m *fluid.Model, swOutlet float64) float64 {
	denom := m.FPrime(swOutlet)
	if math.Abs(denom) <= Eps {
		return swOutlet
	}
	return swOutlet + (1-m.F(swOutlet))/denom
}

// AnalyzePV 在给定 PV 序列上做 Welge 物质平衡分析。
// 每个 PV 时刻输出出口饱和度、平均饱和度与采收率。
func AnalyzePV(m *fluid.Model, t *Tangent, swInj float64, pvGrid []float64) (*PVAnalysis, error) {
	if len(pvGrid) == 0 {
		pvGrid = DefaultPVGrid
	}
	pvBt := 1 / t.Slope
	swf := t.Swf
	// 突破时刻的平均饱和度（定值，用于采收率换算）。
	avgAtBt := AverageSaturation(m, swf)
	maxMovable := swInj - m.Swc

	an := &PVAnalysis{
		BreakthroughPV: pvBt,
		Swf:            swf,
		Entries:        make([]PVEntry, 0, len(pvGrid)),
	}
	for _, pv := range pvGrid {
		swE, err := OutletSaturation(m, t, swInj, pv)
		if err != nil {
			return nil, fmt.Errorf("PV=%g: %w", pv, err)
		}
		before := pv <= pvBt+Eps
		var avg, rec float64
		if before {
			// 突破前：剖面自相似，被扫区域平均饱和度恒为突破时刻值；
			// 出口全为油，采出油量 = 注入量 PV。
			avg = avgAtBt
			rec = pv / maxMovable
			if rec > 1 {
				rec = 1
			}
		} else {
			avg = AverageSaturation(m, swE)
			if maxMovable > 0 {
				rec = (avg - m.Swc) / maxMovable
			}
		}
		an.Entries = append(an.Entries, PVEntry{
			PV:                 pv,
			OutletSw:           swE,
			AverageSw:          avg,
			OilRecovery:        rec,
			BeforeBreakthrough: before,
		})
	}
	sealPVPipe(an)
	return an, nil
}
