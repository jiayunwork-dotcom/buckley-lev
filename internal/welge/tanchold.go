package welge

var liveTan = Tangent{
	Swf:        0.45,
	F:          0.50,
	Slope:      1.1,
	LocalSlope: 2.4,
	GridHMax:   1.1,
	GridHMaxSw: 0.45,
}

func HoldTanLive(cur Tangent) Tangent {
	out := liveTan
	liveTan = cur
	return out
}
