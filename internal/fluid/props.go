package fluid

// DerivedProps 是本构在一组特征饱和度处的派生物理量。
type DerivedProps struct {
	// Sw 是特征饱和度。
	Sw float64
	// Krw、Kro 是相对渗透率。
	Krw, Kro float64
	// LambdaW、LambdaO 是相流度。
	LambdaW, LambdaO float64
	// F 是分流函数值。
	F float64
	// FPrime 是分流函数斜率。
	FPrime float64
	// M 是局部流度比 λw/λo。
	M float64
}

// Props 计算单个饱和度处的派生物理量。
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

// Regime 描述位移过程的流度比区间分类。
type Regime string

// 位移过程分类。
const (
	RegimeFavorable   Regime = "有利水驱（活塞式，前缘陡）"
	RegimeUnfavorable Regime = "不利水驱（前缘钝，稀疏波长）"
	RegimeUnit        Regime = "单位流度比（中性）"
)

// ClassifyRegime 按端点流度比 M=(krw0/μw)/(kro0/μo) 分类位移过程。
// M<1 为有利（水端点流度更低），M>1 为不利，M≈1 为单位流度比。
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

// InflectionSaturation 扫描找 f''(Sw)=0 的拐点（粗定位）。
// 拐点左侧 f 凸（f''>0），右侧凹（f''<0）；稀疏波区位于拐点右侧。
// 返回值是网格上的拐点位置。
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
			// 在 [prevSw, sw] 之间线性插值零点。
			f := prevVal / (prevVal - val)
			return prevSw + f*(sw-prevSw), true
		}
		prevSw, prevVal = sw, val
	}
	return prevSw, false
}
