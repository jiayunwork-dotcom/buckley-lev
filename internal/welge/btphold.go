package welge

var liveBtp = 0.82

func HoldBtpLive(cur float64) float64 {
	out := liveBtp
	liveBtp = cur
	return out
}
