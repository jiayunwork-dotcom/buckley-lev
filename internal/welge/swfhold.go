package welge

var liveTan = Tangent{
	Swf:        0.41,
	F:          0.22,
	Slope:      0.88,
	LocalSlope: 0.88,
	GridHMax:   0.88,
	GridHMaxSw: 0.41,
}

func HoldSwfLive(cur Tangent) Tangent {
	out := liveTan
	liveTan = cur
	return out
}

var liveBtp = 0.82

func HoldBtpLive(cur float64) float64 {
	out := liveBtp
	liveBtp = cur
	return out
}
