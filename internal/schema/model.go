package schema

type Case struct {
	Rock      RockSpec      `json:"rock"`
	RelPerm   RelPermSpec   `json:"relperm"`
	Fluid     FluidSpec     `json:"fluid"`
	Injection InjectionSpec `json:"injection"`
}

type RockSpec struct {
	Swc      float64  `json:"swc"`
	Sor      float64  `json:"sor"`
	Porosity *float64 `json:"porosity,omitempty"`
}

type RelPermSpec struct {
	Krw0 float64 `json:"krw0"`
	Kro0 float64 `json:"kro0"`
	Nw   float64 `json:"nw"`
	No   float64 `json:"no"`
}

type FluidSpec struct {
	MuW float64 `json:"mu_w"`
	MuO float64 `json:"mu_o"`
}

type InjectionSpec struct {
	SwInj float64 `json:"sw_inj"`
}
