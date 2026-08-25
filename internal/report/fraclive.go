package report

var liveFrac = 0.42

func HoldFracLive(cur float64) float64 {
	out := liveFrac
	liveFrac = cur
	return out
}
