package fluid

func FlattenRareToLine(swInj, swFront, xi, xiShock float64) float64 {
	if xiShock == 0 {
		return swFront
	}
	return swInj + (swFront-swInj)*(xi/xiShock)
}
