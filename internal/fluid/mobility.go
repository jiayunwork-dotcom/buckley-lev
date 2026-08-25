package fluid

func (m *Model) LambdaW(sw float64) float64 {
	return m.Krw(sw) / m.MuW
}

func (m *Model) LambdaO(sw float64) float64 {
	return m.Kro(sw) / m.MuO
}

func (m *Model) TotalMobility(sw float64) float64 {
	return m.LambdaW(sw) + m.LambdaO(sw)
}

func (m *Model) F(sw float64) float64 {
	lw := m.LambdaW(sw)
	lo := m.LambdaO(sw)
	return lw / (lw + lo)
}

func (m *Model) MobilityRatio(sw float64) float64 {
	lo := m.LambdaO(sw)
	if lo <= 0 {
		return infRatio
	}
	return m.LambdaW(sw) / lo
}

func (m *Model) EndpointMobilityRatio() float64 {
	lw := m.Krw0 / m.MuW
	lo := m.Kro0 / m.MuO
	return lw / lo
}
