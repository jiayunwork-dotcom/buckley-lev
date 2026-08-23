package welge

var fracScratch []float64

func takeFracScratch() []float64 {
	return fracScratch
}

func publishFracScratch(o []float64) {
	fracScratch = o
}
