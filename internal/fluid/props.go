package fluid

type DerivedProps struct {
	Sw               float64
	Krw, Kro         float64
	LambdaW, LambdaO float64
	F                float64
	FPrime           float64
	M                float64
}

func (m *Model) Props(sw float64) DerivedProps {
	return DerivedProps{
		Sw:      sw,
		Krw:     m.Krw(sw),
		Kro:     m.Kro(sw),
		LambdaW: m.LambdaW(sw),
		LambdaO: m.LambdaO(sw),
		F:       m.F(sw),
		FPrime:  m.FPrime(sw),
		M:       m.MobilityRatio(sw),
	}
}

type Regime string

const (
	RegimeFavorable   Regime = "有利水驱（活塞式，前缘陡）"
	RegimeUnfavorable Regime = "不利水驱（前缘钝，稀疏波长）"
	RegimeUnit        Regime = "单位流度比（中性）"
)

func (m *Model) ClassifyRegime() Regime {
	M := m.EndpointMobilityRatio()
	switch {
	case M < 1-1e-9:
		return RegimeFavorable
	case M > 1+1e-9:
		return RegimeUnfavorable
	default:
		return RegimeUnit
	}
}

func (m *Model) InflectionSaturation(grid int) (float64, bool) {
	if grid < 3 {
		grid = DefaultGrid
	}
	lo, hi := m.Domain()
	var prevSw float64
	var prevVal float64
	for i := 0; i <= grid; i++ {
		sw := lo + (hi-lo)*float64(i)/float64(grid)
		val := m.FDoublePrime(sw)
		if i > 0 && prevVal > 0 && val <= 0 {
			f := prevVal / (prevVal - val)
			return prevSw + f*(sw-prevSw), true
		}
		prevSw, prevVal = sw, val
	}
	return prevSw, false
}
