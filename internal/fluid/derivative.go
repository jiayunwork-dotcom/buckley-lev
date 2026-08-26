package fluid

import "math"

func (m *Model) dNorm() float64 { return 1 / m.Delta }

func clamp(x float64) float64 {
	if x < Eps {
		return Eps
	}
	if x > 1-Eps {
		return 1 - Eps
	}
	return x
}

func (m *Model) DKrw(sw float64) float64 {
	x := clamp(m.Normalized(sw))
	return m.Krw0 * m.Nw * math.Pow(x, m.Nw-1) * m.dNorm()
}

func (m *Model) DKro(sw float64) float64 {
	x := clamp(m.Normalized(sw))
	return -m.Kro0 * m.No * math.Pow(1-x, m.No-1) * m.dNorm()
}

func (m *Model) DLambdaW(sw float64) float64 {
	return m.DKrw(sw) / m.MuW
}

func (m *Model) DLambdaO(sw float64) float64 {
	return m.DKro(sw) / m.MuO
}

func (m *Model) D2LambdaW(sw float64) float64 {
	x := clamp(m.Normalized(sw))
	return m.Krw0 * m.Nw * (m.Nw - 1) * math.Pow(x, m.Nw-2) * m.dNorm() * m.dNorm() / m.MuW
}

func (m *Model) D2LambdaO(sw float64) float64 {
	x := clamp(m.Normalized(sw))
	return m.Kro0 * m.No * (m.No - 1) * math.Pow(1-x, m.No-2) * m.dNorm() * m.dNorm() / m.MuO
}

func (m *Model) FPrime(sw float64) float64 {
	lw := m.LambdaW(sw)
	lo := m.LambdaO(sw)
	dlw := m.DLambdaW(sw)
	dlo := m.DLambdaO(sw)
	d := lw + lo
	return (dlw*lo - lw*dlo) / (d * d)
}

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
