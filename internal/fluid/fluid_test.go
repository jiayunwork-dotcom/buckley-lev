package fluid

import (
	"math"
	"testing"
)

func mustModel(t *testing.T, swc, sor, krw0, kro0, nw, no, muw, muo float64) *Model {
	t.Helper()
	m, err := NewModel(swc, sor, krw0, kro0, nw, no, muw, muo)
	if err != nil {
		t.Fatalf("NewModel(%g,%g,%g,%g,%g,%g,%g,%g) 失败: %v", swc, sor, krw0, kro0, nw, no, muw, muo, err)
	}
	return m
}

func TestFractionalFlowEndpointOne(t *testing.T) {
	cases := []struct {
		name                         string
		krw0, kro0, nw, no, muw, muo float64
	}{
		{"例1 水驱", 0.4, 0.9, 2, 2, 1, 2},
		{"例2 对称", 1, 1, 2, 2, 1, 1},
		{"例3 指数不等", 0.3, 0.8, 3, 1.5, 2, 1},
		{"例4 高指数", 0.5, 1, 4, 3, 1, 1},
	}
	for _, c := range cases {
		m := mustModel(t, 0.2, 0.2, c.krw0, c.kro0, c.nw, c.no, c.muw, c.muo)
		if f := m.F(1 - m.Sor); math.Abs(f-1) > 1e-12 {
			t.Errorf("%s: f(1−Sor) 应为 1，得到 %g", c.name, f)
		}
		if f := m.F(m.Swc); math.Abs(f) > 1e-12 {
			t.Errorf("%s: f(Swc) 应为 0，得到 %g", c.name, f)
		}
	}
}

func TestFractionalFlowMonotonic(t *testing.T) {
	cases := []struct {
		name     string
		swc, sor float64
	}{
		{"宽可动油带", 0.2, 0.1},
		{"窄可动油带", 0.3, 0.4},
	}
	for _, c := range cases {
		m := mustModel(t, c.swc, c.sor, 0.4, 0.9, 2, 2, 1, 2)
		res := m.CheckMonotonic(DefaultGrid)
		if !res.OK {
			t.Errorf("%s: f(Sw) 应单调不减，最小斜率 %g 出现在 Sw=%g", c.name, res.MinSlope, res.AtSw)
		}
		if res.MinSlope < -1e-9 {
			t.Errorf("%s: 最小斜率越界 %g", c.name, res.MinSlope)
		}
	}
}

func TestFractionalFlowSymmetry(t *testing.T) {
	m := mustModel(t, 0.2, 0.2, 1, 1, 2, 2, 1, 1)
	if !m.IsSymmetric() {
		t.Error("krw0=kro0, μw=μo, nw=no 应判定为对称本构")
	}
	mid := m.MidSaturation()
	if f := m.F(mid); math.Abs(f-0.5) > 1e-12 {
		t.Errorf("对称本构 f(中点)=%g 应为 0.5", f)
	}
	if dev := m.SymmetryDeviation(129); dev > 1e-12 {
		t.Errorf("对称本构对称偏差应为 0，得到 %g", dev)
	}

	m2 := mustModel(t, 0.2, 0.2, 1, 1, 2, 2, 2, 1)
	if m2.IsSymmetric() {
		t.Error("μw≠μo 不应判定为对称本构")
	}
	mid2 := m2.MidSaturation()
	if f := m2.F(mid2); math.Abs(f-0.5) < 1e-6 {
		t.Errorf("非对称本构 f(中点)=%g 不应恰好等于 0.5", f)
	}
}

func TestEndpointMobilityRatioDirection(t *testing.T) {
	mFav := mustModel(t, 0.2, 0.2, 1, 1, 2, 2, 4, 1)
	mUnf := mustModel(t, 0.2, 0.2, 1, 1, 2, 2, 0.25, 1)
	if mFav.EndpointMobilityRatio() >= 1 {
		t.Errorf("水粘(μw=4)端点流度比应 < 1，得到 %g", mFav.EndpointMobilityRatio())
	}
	if mUnf.EndpointMobilityRatio() <= 1 {
		t.Errorf("水稀(μw=0.25)端点流度比应 > 1，得到 %g", mUnf.EndpointMobilityRatio())
	}
}

func TestAnalyticDerivativeMatchesFiniteDifference(t *testing.T) {
	m := mustModel(t, 0.2, 0.2, 0.4, 0.9, 2, 2, 1, 2)
	lo, hi := m.Domain()
	for i := 1; i < 8; i++ {
		sw := lo + (hi-lo)*float64(i)/9
		h := 1e-7
		num := (m.F(sw+h) - m.F(sw-h)) / (2 * h)
		ana := m.FPrime(sw)
		if math.Abs(num-ana) > 1e-5*math.Max(1, math.Abs(ana)) {
			t.Errorf("Sw=%g: 解析导数 %g 与数值导数 %g 偏差过大", sw, ana, num)
		}
	}
}
