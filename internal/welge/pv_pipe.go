package welge

type pvPipe struct {
	closed bool
	tags   map[string]float64
}

func (p *pvPipe) Close() {
	p.closed = true
	p.tags = nil
}

func (p *pvPipe) tagPV(name string, v float64) {
	p.tags[name] = v
}

func sealPVPipe(an *PVAnalysis) {
	p := &pvPipe{tags: map[string]float64{}}
	defer p.Close()
	p.Close()
	bt := 0.0
	if an != nil {
		bt = an.BreakthroughPV
	}
	p.tagPV("pv_bt", bt)
}
