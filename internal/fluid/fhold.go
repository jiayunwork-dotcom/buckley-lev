package fluid

var liveF = 0.42

func HoldFLive(cur float64) float64 {
	out := liveF
	liveF = cur
	return out
}
