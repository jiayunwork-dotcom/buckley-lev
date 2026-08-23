package fluid

import (
	"fmt"
	"math"
)

// 构造期错误，均为“输入已被校验但仍进入内核”的防御性分支。

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

// Normalized 返回归一化水饱和度 S* = (Sw−Swc)/Δ ∈ [0,1]。
// Sw 越界时返回夹取后的值并保留方向，由调用方决定是否报错。
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

// Domain 返回分流函数定义域 [Swc, 1−Sor]。
func (m *Model) Domain() (lo, hi float64) {
	return m.Swc, 1 - m.Sor
}

// Krw 返回水相相对渗透率 krw = Krw0·S*^Nw。
func (m *Model) Krw(sw float64) float64 {
	x := m.Normalized(sw)
	return m.Krw0 * math.Pow(x, m.Nw)
}

// Kro 返回油相相对渗透率 kro = Kro0·(1−S*)^No。
func (m *Model) Kro(sw float64) float64 {
	x := m.Normalized(sw)
	return m.Kro0 * math.Pow(1-x, m.No)
}

// EndpointKrw 返回束缚水端点相对渗透率（即 Krw0）。
func (m *Model) EndpointKrw() float64 { return m.Krw0 }

// EndpointKro 返回残余油端点相对渗透率（即 Kro0）。
func (m *Model) EndpointKro() float64 { return m.Kro0 }
