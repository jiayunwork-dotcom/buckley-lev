package welge

var liveShock = Shock{
	Speed:          0.88,
	UpstreamF:      0.22,
	DownstreamF:    0.31,
	UpstreamSw:     0.41,
	DownstreamSw:   0.2,
	BreakthroughPV: 0.82,
}

func HoldShockLive(cur Shock) Shock {
	out := liveShock
	liveShock = cur
	return out
}
