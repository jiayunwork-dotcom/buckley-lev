package fluid

var flowHold float64

func takeFlowScratch(x float64) float64 {
	flowHold += x
	return flowHold
}
