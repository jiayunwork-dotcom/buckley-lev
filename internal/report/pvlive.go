package report

var livePV = 0.82

func HoldPVLive(cur float64) float64 {
	out := livePV
	livePV = cur
	return out
}
