// Package schema 定义一维两相水驱（Buckley–Leverett）核算的输入模型、
// JSON 解析、默认值填充与物理合法性校验。
package schema

// Case 是一个完整的水驱算例：岩石、相对渗透率、流体粘度与注入条件。
// 所有字段都有明确的物理含义；缺字段由 JSON 零值落入，随后被
// Validate 用明确错误拒绝（不静默补零）。
type Case struct {
	Rock      RockSpec      `json:"rock"`
	RelPerm   RelPermSpec   `json:"relperm"`
	Fluid     FluidSpec     `json:"fluid"`
	Injection InjectionSpec `json:"injection"`
}

// RockSpec 描述油藏的岩石与流体基础物性。
type RockSpec struct {
	// Swc 是束缚水饱和度，范围 [0,1)。
	Swc float64 `json:"swc"`
	// Sor 是残余油饱和度，范围 [0,1)。
	Sor float64 `json:"sor"`
	// Porosity 是孔隙度（可选，默认 0.25），只用于换算突破孔隙体积。
	// 指针字段用于区分“未给出”（nil → 默认）与“显式给出 0”（非法）。
	Porosity *float64 `json:"porosity,omitempty"`
}

// RelPermSpec 是 Corey 幂次形式的相对渗透率端点与指数。
// krw = Krw0·S*^Nw，kro = Kro0·(1−S*)^No，其中
// S* = (Sw−Swc)/(1−Swc−Sor) 为归一化水饱和度。
type RelPermSpec struct {
	// Krw0 是束缚水饱和度处的水端点相对渗透率，必须 > 0。
	Krw0 float64 `json:"krw0"`
	// Kro0 是残余油饱和度处的油端点相对渗透率，必须 > 0。
	Kro0 float64 `json:"kro0"`
	// Nw 是水相 Corey 指数，必须 > 1（线性曲线会使 Welge 切线退化）。
	Nw float64 `json:"nw"`
	// No 是油相 Corey 指数，必须 > 1。
	No float64 `json:"no"`
}

// FluidSpec 描述两相粘度。运动分量用分流量公式中的“粘度/相对渗透率”
// 构成流度，粘度必须为正。
type FluidSpec struct {
	// MuW 是水粘度（mPa·s 或任意一致单位），必须 > 0。
	MuW float64 `json:"mu_w"`
	// MuO 是油粘度（与 MuW 同单位），必须 > 0。
	MuO float64 `json:"mu_o"`
}

// InjectionSpec 描述注入端边界条件。
type InjectionSpec struct {
	// SwInj 是注入端水饱和度，必须落在 (Swc, 1−Sor]。
	// 水驱工程惯例为 SwInj = 1−Sor，此时该端分流量为 1。
	SwInj float64 `json:"sw_inj"`
}
