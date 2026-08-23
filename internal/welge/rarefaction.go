package welge

import (
	"fmt"

	"buckley-lev/internal/fluid"
)

// RarefactionSaturation 在激波后方求取无因次位置 ξ 处的水饱和度。
//
// Buckley–Leverett 特征线：固定饱和度以速度 dx/dt ∝ f'(S) 传播，
// 因此无因次速度 ξ=x/t 与饱和度一一对应：ξ = f'(S)。
// 在 [Swf, Sw_inj] 上 f' 单调递减（f 为凹段），故对给定 ξ 可唯一
// 反解 S。这不是用直线连接激波与注入端的近似。
//
// 规则：
//   - ξ ≥ ξs        → Sw = Swf（激波前缘）
//   - ξ ≤ f'(Sw_inj) → Sw = Sw_inj（注入端平台，注入水未开始稀薄）
//   - 其余           → 在 [Swf, Sw_inj] 上二分 f'(S) = ξ
//
// 若反解区间非单调（f''(Swf) ≥ 0），说明参数组合下稀疏波区域不
// 成立，返回错误（对应“Sw 越界”）。
func RarefactionSaturation(m *fluid.Model, t *Tangent, swInj, xi float64) (float64, error) {
	if swInj <= m.Swc {
		return 0, fmt.Errorf("稀疏波区域非法：注入饱和度 %g 必须高于 Swc=%g", swInj, m.Swc)
	}
	if swInj <= t.Swf {
		return 0, fmt.Errorf("Sw 越界：注入饱和度 %g 低于前缘饱和度 Swf=%g，稀疏波区域为空", swInj, t.Swf)
	}

	// 端点处理。
	if xi >= t.Slope-Eps {
		return t.Swf, nil
	}
	endSlope := m.FPrime(swInj)
	if xi <= endSlope+Eps {
		return swInj, nil
	}

	// 反解区间的单调性前提：f'(Swf)=ξs ≥ ξ ≥ f'(Sw_inj)。
	if m.FPrime(t.Swf) < xi-Eps {
		return 0, fmt.Errorf("稀疏波无解：ξ=%g 高于前缘局部斜率 f'(Swf)=%g", xi, m.FPrime(t.Swf))
	}
	if endSlope > xi+Eps {
		return 0, fmt.Errorf("稀疏波无解：ξ=%g 低于注入端局部斜率 f'(Sw_inj)=%g", xi, endSlope)
	}
	if m.FDoublePrime(t.Swf) >= 0 {
		return 0, fmt.Errorf("稀疏波区域非凹：f''(Swf)=%g ≥ 0，[Swf, Sw_inj] 上 f' 不单调", m.FDoublePrime(t.Swf))
	}

	// 二分求 f'(S) = ξ，S ∈ [Swf, Sw_inj]。
	a, b := t.Swf, swInj
	for i := 0; i < MaxIter; i++ {
		mid := 0.5 * (a + b)
		if b-a < Eps {
			return mid, nil
		}
		slope := m.FPrime(mid)
		if slope > xi {
			a = mid
		} else {
			b = mid
		}
	}
	return 0.5 * (a + b), nil
}
