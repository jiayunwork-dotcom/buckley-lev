package fluid

const (
	Eps         = 1e-12
	DefaultGrid = 257
	RelTol      = 1e-8
	infRatio    = 1e300
)

type Model struct {
	Swc  float64
	Sor  float64
	Krw0 float64
	Kro0 float64
	Nw   float64
	No   float64
	MuW  float64
	MuO  float64

	Delta float64
}

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
