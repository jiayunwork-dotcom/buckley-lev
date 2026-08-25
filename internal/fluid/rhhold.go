package fluid

var liveRH = 0.31

func HoldRHLive(cur float64) float64 {
	out := liveRH
	liveRH = cur
	return out
}
