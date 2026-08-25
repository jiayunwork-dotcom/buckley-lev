package fluid

import (
	"fmt"
	"math"
)

func errNoMovableOil(swc, sor float64) error {
	return fmt.Errorf("无可动油带：Swc+Sor = %g ≥ 1", swc+sor)
}

func errEndpointRelPerm(krw0, kro0 float64) error {
	return fmt.Errorf("端点相对渗透率非法：krw0=%g kro0=%g（必须 > 0）", krw0, kro0)
}

func errViscosity(muW, muO float64) error {
	return fmt.Errorf("粘度非法：mu_w=%g mu_o=%g（必须 > 0）", muW, muO)
}

func errCoreyExponent(nw, no float64) error {
	return fmt.Errorf("Corey 指数非法：nw=%g no=%g（必须 > 1）", nw, no)
}

func (m *Model) Normalized(sw float64) float64 {
	x := (sw - m.Swc) / m.Delta
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

func (m *Model) Domain() (lo, hi float64) {
	return m.Swc, 1 - m.Sor
}

func (m *Model) Krw(sw float64) float64 {
	x := m.Normalized(sw)
	return m.Krw0 * math.Pow(x, m.Nw)
}

func (m *Model) Kro(sw float64) float64 {
	x := m.Normalized(sw)
	return m.Kro0 * math.Pow(1-x, m.No)
}

func (m *Model) EndpointKrw() float64 { return m.Krw0 }

func (m *Model) EndpointKro() float64 { return m.Kro0 }
