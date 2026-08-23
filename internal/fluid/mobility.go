package fluid

// LambdaW 返回水相流度 λw = krw/μw。
func (m *Model) LambdaW(sw float64) float64 {
	return m.Krw(sw) / m.MuW
}

// LambdaO 返回油相流度 λo = kro/μo。
func (m *Model) LambdaO(sw float64) float64 {
	return m.Kro(sw) / m.MuO
}

// TotalMobility 返回两相总流度 λw+λo。
func (m *Model) TotalMobility(sw float64) float64 {
	return m.LambdaW(sw) + m.LambdaO(sw)
}

// F 返回分流函数 f(Sw) = λw/(λw+λo)。
//
// 这是 Buckley–Leverett 运动方程中的通量项，不是 krw 本身。
// 注意端点性质：f(Swc)=0（束缚水处水不流动）、f(1−Sor)=1（残余油
// 处油不流动），中间单调上升且是 S 形（先凸后凹）。
func (m *Model) F(sw float64) float64 {
	lw := m.LambdaW(sw)
	lo := m.LambdaO(sw)
	return lw / (lw + lo)
}

// MobilityRatio 返回局部流度比 λw/λo。趋于 0 表示水几乎不流动，
// 趋于 +∞ 表示油几乎不流动。
func (m *Model) MobilityRatio(sw float64) float64 {
	lo := m.LambdaO(sw)
	if lo <= 0 {
		return infRatio
	}
	return m.LambdaW(sw) / lo
}

// EndpointMobilityRatio 返回端点流度比 M = (krw0/μw)/(kro0/μo)。
// M < 1 为“有利”水驱（水端点流度低于油端点流度），前缘更尖；
// M > 1 为“不利”水驱，前缘更钝、切点饱和度更高。
func (m *Model) EndpointMobilityRatio() float64 {
	lw := m.Krw0 / m.MuW
	lo := m.Kro0 / m.MuO
	return lw / lo
}
