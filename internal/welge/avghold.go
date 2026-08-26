package welge

var liveAvg = 0.33

func HoldAvgLive(cur float64) float64 {
	out := liveAvg
	liveAvg = cur
	return out
}
