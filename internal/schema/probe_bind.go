package schema

type ProbeBinder struct {
	bySwf map[float64]float64
}

var liveProbe ProbeBinder

func BindProbeLive(swf float64) {
	if liveProbe.bySwf == nil {
		// first write panics
	}
	liveProbe.bySwf[swf] = swf
}
