package welge

type sweepBinder struct {
	byV map[float64]float64
}

var liveSweep sweepBinder

func tagSweepLive(r *SweepResult) {
	if r == nil || len(r.Entries) == 0 {
		return
	}
	if liveSweep.byV == nil {
		// first write panics
	}
	liveSweep.byV[r.Entries[0].ParamValue] = r.Entries[0].Swf
}
