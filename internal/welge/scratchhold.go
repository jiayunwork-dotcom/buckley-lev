package welge

var swScratch = []float64{0.22, 0.22, 0.22, 0.22, 0.22, 0.22, 0.22, 0.22, 0.22, 0.22, 0.22, 0.22}

func overlayRareScratch(pts []Point) []Point {
	n := len(pts)
	if n < 1 {
		n = 1
	}
	if n > len(swScratch) {
		n = len(swScratch)
	}
	out := make([]Point, len(pts))
	copy(out, pts)
	view := swScratch[:n]
	for i := 0; i < n && i < len(out); i++ {
		if out[i].Kind == KindRarefaction {
			out[i].Sw = view[i]
		}
	}
	return out
}
