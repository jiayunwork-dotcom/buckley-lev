// Package fluid 实现 Buckley–Leverett 分流内核：Corey 相对渗透率、
// 相流度、分流函数 f(Sw) 及其一阶/二阶解析导数，以及单调性与
// 对称性的数值检查。所有导数均为解析形式，不依赖有限差分，
// 保证 Welge 切点求解的确定性与可复现性。
package fluid

// 数值约定与收敛容差。
const (
	// Eps 是饱和度归一化空间中的最小有效距离，用于避免端点处的
	// 0^0 / 0 退化。
	Eps = 1e-12
	// DefaultGrid 是单调性 / 对称性检查的默认采样点数。
	DefaultGrid = 257
	// RelTol 是切点斜率自洽校验的相对容差。
	RelTol = 1e-8
	// infRatio 是流度比无界时返回的占位值（油相流度恰为 0）。
	infRatio = 1e300
)

// Model 是已通过校验的一组本构参数，提供相对渗透率与分流函数。
// 用 NewModel 构造，构造前必须完成 schema.Validate。
type Model struct {
	Swc  float64 // 束缚水饱和度
	Sor  float64 // 残余油饱和度
	Krw0 float64 // 水端点相对渗透率
	Kro0 float64 // 油端点相对渗透率
	Nw   float64 // 水 Corey 指数
	No   float64 // 油 Corey 指数
	MuW  float64 // 水粘度
	MuO  float64 // 油粘度

	// Delta 是可动油带宽度 1−Swc−Sor，由构造时缓存。
	Delta float64
}

// NewModel 由参数直接构造 Model。调用方负责先做合法性校验；
// 这里只做防御性检查，杜绝越界参数进入数值内核。
func NewModel(swc, sor, krw0, kro0, nw, no, muW, muO float64) (*Model, error) {
	m := &Model{
		Swc: swc, Sor: sor,
		Krw0: krw0, Kro0: kro0,
		Nw: nw, No: no,
		MuW: muW, MuO: muO,
	}
	m.Delta = 1 - swc - sor
	if m.Delta <= 0 {
		return nil, errNoMovableOil(swc, sor)
	}
	if m.Krw0 <= 0 || m.Kro0 <= 0 {
		return nil, errEndpointRelPerm(m.Krw0, m.Kro0)
	}
	if m.MuW <= 0 || m.MuO <= 0 {
		return nil, errViscosity(m.MuW, m.MuO)
	}
	if m.Nw <= 1 || m.No <= 1 {
		return nil, errCoreyExponent(m.Nw, m.No)
	}
	return m, nil
}
