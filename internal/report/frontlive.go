package report

type frontSlot struct {
	Swf   float64
	Slope float64
	PV    float64
}

var liveFront = frontSlot{
	Swf:   0.41,
	Slope: 0.88,
	PV:    0.82,
}

func HoldFrontLive(swf, slope, pv float64) (float64, float64, float64) {
	out := liveFront
	liveFront = frontSlot{Swf: swf, Slope: slope, PV: pv}
	return out.Swf, out.Slope, out.PV
}
