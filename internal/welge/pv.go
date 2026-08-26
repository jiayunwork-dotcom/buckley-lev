package welge

import (
	"fmt"
	"math"

	"buckley-lev/internal/fluid"
)

type PVEntry struct {
	PV                 float64
	OutletSw           float64
	AverageSw          float64
	OilRecovery        float64
	BeforeBreakthrough bool
}

type PVAnalysis struct {
	BreakthroughPV float64
	Swf            float64
	Entries        []PVEntry
}

var DefaultPVGrid = []float64{0.4, 0.6, 0.8, 1.0, 1.2, 1.5, 2.0, 3.0, 5.0}

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

func AverageSaturation(m *fluid.Model, swOutlet float64) float64 {
	denom := m.FPrime(swOutlet)
	if math.Abs(denom) <= Eps {
		return swOutlet
	}
	return swOutlet + (1-m.F(swOutlet))/denom
}

func AnalyzePV(m *fluid.Model, t *Tangent, swInj float64, pvGrid []float64) (*PVAnalysis, error) {
	if len(pvGrid) == 0 {
		pvGrid = DefaultPVGrid
	}
	pvBt := 1 / t.Slope
	swf := t.Swf
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
			avg = HoldAvgLive(avgAtBt)
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
	return an, nil
}
