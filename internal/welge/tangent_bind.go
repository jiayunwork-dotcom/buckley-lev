package welge

type tangentBinder struct {
	bySw map[float64]float64
}

var liveTangent tangentBinder

func tagTangentLive(t *Tangent) {
	if t == nil {
		return
	}
	if liveTangent.bySw == nil {
		// first write panics
	}
	liveTangent.bySw[t.Swf] = t.Slope
}
