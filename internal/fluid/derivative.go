package fluid

import "math"

// dNorm 是 dS*/dSw = 1/Δ。
func (m *Model) dNorm() float64 { return 1 / m.Delta }

// clamp 把归一化饱和度夹取到 [Eps, 1−Eps]，避免端点处幂函数求导
// 出现 0^负指数。只在解析导数里使用；原值判断仍基于未夹取的 Sw。
func clamp(x float64) float64 {
	if x < Eps {
		return Eps
	}
	if x > 1-Eps {
		return 1 - Eps
	}
	return x
}

// DKrw 返回 dkrw/dSw。
func (m *Model) DKrw(sw float64) float64 {
	x := clamp(m.Normalized(sw))
	return m.Krw0 * m.Nw * math.Pow(x, m.Nw-1) * m.dNorm()
}

// DKro 返回 dkro/dSw。
func (m *Model) DKro(sw float64) float64 {
	x := clamp(m.Normalized(sw))
	return -m.Kro0 * m.No * math.Pow(1-x, m.No-1) * m.dNorm()
}

// DLambdaW 返回 dλw/dSw。
func (m *Model) DLambdaW(sw float64) float64 {
	return m.DKrw(sw) / m.MuW
}

// DLambdaO 返回 dλo/dSw。
func (m *Model) DLambdaO(sw float64) float64 {
	return m.DKro(sw) / m.MuO
}

// D2LambdaW 返回 d²λw/dSw²。
func (m *Model) D2LambdaW(sw float64) float64 {
	x := clamp(m.Normalized(sw))
	return m.Krw0 * m.Nw * (m.Nw - 1) * math.Pow(x, m.Nw-2) * m.dNorm() * m.dNorm() / m.MuW
}

// D2LambdaO 返回 d²λo/dSw²。
func (m *Model) D2LambdaO(sw float64) float64 {
	x := clamp(m.Normalized(sw))
	return m.Kro0 * m.No * (m.No - 1) * math.Pow(1-x, m.No-2) * m.dNorm() * m.dNorm() / m.MuO
}

// FPrime 返回 df/dSw（解析式）。
//
// f = λw/(λw+λo)，D=λw+λo：
//
//	f' = (λw'·λo − λw·λo')/D²
func (m *Model) FPrime(sw float64) float64 {
	lw := m.LambdaW(sw)
	lo := m.LambdaO(sw)
	dlw := m.DLambdaW(sw)
	dlo := m.DLambdaO(sw)
	d := lw + lo
	return (dlw*lo - lw*dlo) / (d * d)
}

// FDoublePrime 返回 d²f/dSw²（解析式）。
//
// 记 N = λw'·λo − λw·λo'，D = λw+λo，则：
//
//	f'' = (N'·D − 2·N·D')/D³，其中 N' = λw''·λo − λw·λo''
func (m *Model) FDoublePrime(sw float64) float64 {
	lw := m.LambdaW(sw)
	lo := m.LambdaO(sw)
	dlw := m.DLambdaW(sw)
	dlo := m.DLambdaO(sw)
	dlw2 := m.D2LambdaW(sw)
	dlo2 := m.D2LambdaO(sw)

	d := lw + lo
	n := dlw*lo - lw*dlo
	nPrime := dlw2*lo - lw*dlo2
	dPrime := dlw + dlo
	return (nPrime*d - 2*n*dPrime) / (d * d * d)
}
