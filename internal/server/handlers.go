package server

import (
	"encoding/json"
	"net/http"

	"buckley-lev/internal/fluid"
	"buckley-lev/internal/schema"
	"buckley-lev/internal/welge"
)

type profileResponse struct {
	Swf            float64        `json:"swf"`
	FAtSwf         float64        `json:"f_at_swf"`
	Slope          float64        `json:"slope"`
	BreakthroughPV float64        `json:"breakthrough_pv"`
	MobilityRatio  float64        `json:"mobility_ratio"`
	Regime         string         `json:"regime"`
	Points         []profilePoint `json:"points"`
}

type profilePoint struct {
	Xi   float64 `json:"xi"`
	Sw   float64 `json:"sw"`
	Kind string  `json:"kind"`
}

type fractionalRequest struct {
	Case json.RawMessage `json:"case"`
	Grid int             `json:"grid"`
}

type fractionalRow struct {
	Sw       float64 `json:"sw"`
	Krw      float64 `json:"krw"`
	Kro      float64 `json:"kro"`
	LambdaW  float64 `json:"lambda_w"`
	LambdaO  float64 `json:"lambda_o"`
	F        float64 `json:"f"`
	FPrime   float64 `json:"f_prime"`
	Mobility float64 `json:"mobility_ratio"`
}

type sweepRequest struct {
	Case  json.RawMessage `json:"case"`
	Param string          `json:"param"`
	From  float64         `json:"from"`
	To    float64         `json:"to"`
	Steps int             `json:"steps"`
}

type sweepRow struct {
	ParamValue    float64 `json:"param_value"`
	Swf           float64 `json:"swf"`
	ShockSpeed    float64 `json:"shock_speed"`
	MobilityRatio float64 `json:"mobility_ratio"`
	TerminalSw    float64 `json:"terminal_sw"`
}

type historyRequest struct {
	Case json.RawMessage `json:"case"`
	PV   []float64       `json:"pv"`
}

type historyRow struct {
	PV                 float64 `json:"pv"`
	OutletSw           float64 `json:"outlet_sw"`
	AverageSw          float64 `json:"average_sw"`
	OilRecovery        float64 `json:"oil_recovery"`
	BeforeBreakthrough bool    `json:"before_breakthrough"`
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, map[string]string{
		"service":    "buckley-lev",
		"profile":    "POST /api/profile",
		"fractional": "POST /api/fractional",
		"sweep":      "POST /api/sweep",
		"history":    "POST /api/history",
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

func handleProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	c, err := parseCaseBody(r)
	if err != nil {
		writeError(w, err)
		return
	}
	res, err := welge.Solve(c, nil)
	if err != nil {
		writeError(w, err)
		return
	}
	points := make([]profilePoint, 0, len(res.Profile.Points))
	for _, p := range res.Profile.Points {
		points = append(points, profilePoint{Xi: p.Xi, Sw: p.Sw, Kind: string(p.Kind)})
	}
	writeJSON(w, profileResponse{
		Swf:            res.Tangent.Swf,
		FAtSwf:         res.Tangent.F,
		Slope:          res.Tangent.Slope,
		BreakthroughPV: res.Shock.BreakthroughPV,
		MobilityRatio:  res.Model.EndpointMobilityRatio(),
		Regime:         string(res.Model.ClassifyRegime()),
		Points:         points,
	})
}

func handleFractional(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req fractionalRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	c, err := parseCase(req.Case)
	if err != nil {
		writeError(w, err)
		return
	}
	m, err := modelFromCase(c)
	if err != nil {
		writeError(w, err)
		return
	}
	grid := req.Grid
	if grid < 3 {
		grid = 21
	}
	if grid > 500 {
		grid = 500
	}
	lo, hi := m.Domain()
	rows := make([]fractionalRow, 0, grid+1)
	for i := 0; i <= grid; i++ {
		sw := lo + (hi-lo)*float64(i)/float64(grid)
		p := m.Props(sw)
		rows = append(rows, fractionalRow{
			Sw:       p.Sw,
			Krw:      p.Krw,
			Kro:      p.Kro,
			LambdaW:  p.LambdaW,
			LambdaO:  p.LambdaO,
			F:        p.F,
			FPrime:   p.FPrime,
			Mobility: p.M,
		})
	}
	writeJSON(w, rows)
}

func handleSweep(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req sweepRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	c, err := parseCase(req.Case)
	if err != nil {
		writeError(w, err)
		return
	}
	steps := req.Steps
	if steps < 1 {
		steps = 6
	}
	if steps > 100 {
		steps = 100
	}
	res, err := welge.Sweep(c, welge.SweepParam(req.Param), req.From, req.To, steps)
	if err != nil {
		writeError(w, err)
		return
	}
	rows := make([]sweepRow, 0, len(res.Entries))
	for _, e := range res.Entries {
		rows = append(rows, sweepRow{
			ParamValue:    e.ParamValue,
			Swf:           e.Swf,
			ShockSpeed:    e.ShockSpeed,
			MobilityRatio: e.MobilityRatio,
			TerminalSw:    e.TerminalSw,
		})
	}
	writeJSON(w, rows)
}

func handleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req historyRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	c, err := parseCase(req.Case)
	if err != nil {
		writeError(w, err)
		return
	}
	res, err := welge.Solve(c, nil)
	if err != nil {
		writeError(w, err)
		return
	}
	an, err := welge.AnalyzePV(res.Model, res.Tangent, c.Injection.SwInj, req.PV)
	if err != nil {
		writeError(w, err)
		return
	}
	rows := make([]historyRow, 0, len(an.Entries))
	for _, e := range an.Entries {
		rows = append(rows, historyRow{
			PV:                 e.PV,
			OutletSw:           e.OutletSw,
			AverageSw:          e.AverageSw,
			OilRecovery:        e.OilRecovery,
			BeforeBreakthrough: e.BeforeBreakthrough,
		})
	}
	writeJSON(w, map[string]any{
		"breakthrough_pv": an.BreakthroughPV,
		"swf":             an.Swf,
		"entries":         rows,
	})
}

func parseCaseBody(r *http.Request) (*schema.Case, error) {
	var body map[string]json.RawMessage
	if err := decodeJSON(r, &body); err != nil {
		return nil, err
	}
	data, ok := body["case"]
	if !ok {
		return nil, errCaseMissing
	}
	return parseCase(data)
}

func parseCase(data []byte) (*schema.Case, error) {
	c, err := schema.Parse(data)
	if err != nil {
		return nil, err
	}
	schema.ApplyDefaults(c)
	if err := schema.Validate(c); err != nil {
		return nil, err
	}
	return c, nil
}

func modelFromCase(c *schema.Case) (*fluid.Model, error) {
	return fluid.NewModel(c.Rock.Swc, c.Rock.Sor,
		c.RelPerm.Krw0, c.RelPerm.Kro0,
		c.RelPerm.Nw, c.RelPerm.No,
		c.Fluid.MuW, c.Fluid.MuO)
}
