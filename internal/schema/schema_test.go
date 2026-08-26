package schema

import (
	"testing"
)

func validCase() *Case {
	p := 0.25
	return &Case{
		Rock:      RockSpec{Swc: 0.2, Sor: 0.2, Porosity: &p},
		RelPerm:   RelPermSpec{Krw0: 0.4, Kro0: 0.9, Nw: 2, No: 2},
		Fluid:     FluidSpec{MuW: 1, MuO: 2},
		Injection: InjectionSpec{SwInj: 0.8},
	}
}

func TestParseAcceptsValidJSON(t *testing.T) {
	raw := `{
		"rock": {"swc": 0.2, "sor": 0.2},
		"relperm": {"krw0": 0.4, "kro0": 0.9, "nw": 2, "no": 2},
		"fluid": {"mu_w": 1, "mu_o": 2},
		"injection": {"sw_inj": 0.8}
	}`
	c, err := Parse([]byte(raw))
	if err != nil {
		t.Fatalf("Parse 合法 JSON 应成功，得到错误: %v", err)
	}
	if c.Rock.Swc != 0.2 || c.Rock.Sor != 0.2 {
		t.Errorf("解析 rock 字段不符: swc=%g sor=%g", c.Rock.Swc, c.Rock.Sor)
	}
	if c.Fluid.MuW != 1 || c.Fluid.MuO != 2 {
		t.Errorf("解析 fluid 字段不符: mu_w=%g mu_o=%g", c.Fluid.MuW, c.Fluid.MuO)
	}
	if c.Injection.SwInj != 0.8 {
		t.Errorf("解析 injection 字段不符: sw_inj=%g", c.Injection.SwInj)
	}
}

func TestParseRejectsMalformedJSON(t *testing.T) {
	bad := []string{
		"",
		"   ",
		"{bad json",
		`{"rock": {`,
		`{"rock": [1,2,3]}`,
		`{"rock": {"swc": "oops"}}`,
		`{"rock": {"swc": 0.2, "sor": 0.2, "typo": 1}}`,
		`{"injection": {"sw_inj": 0.8, "wrong_key": true}}`,
		`{"rock": {}} {"rock": {}}`,
	}
	for _, s := range bad {
		if _, err := Parse([]byte(s)); err == nil {
			t.Errorf("Parse(%q) 应返回错误，但解析成功", s)
		}
	}
}

func TestDefaultsFillPorosity(t *testing.T) {
	c := validCase()
	c.Rock.Porosity = nil
	ApplyDefaults(c)
	if c.Rock.Porosity == nil || *c.Rock.Porosity != DefaultPorosity {
		t.Errorf("缺省孔隙度应为 %g，得到 %v", DefaultPorosity, c.Rock.Porosity)
	}
}

func TestValidateAcceptsValidCase(t *testing.T) {
	if err := Validate(validCase()); err != nil {
		t.Errorf("合法算例应通过校验，得到错误: %v", err)
	}
}

func TestValidateRejectsNonPhysicalInputs(t *testing.T) {
	bad := []struct {
		name   string
		mutate func(*Case)
	}{
		{"Swc 为负", func(c *Case) { c.Rock.Swc = -0.1 }},
		{"Sor 为负", func(c *Case) { c.Rock.Sor = -0.05 }},
		{"Swc+Sor≥1 无可动油", func(c *Case) { c.Rock.Swc = 0.6; c.Rock.Sor = 0.4 }},
		{"水粘度为 0", func(c *Case) { c.Fluid.MuW = 0 }},
		{"油粘度为负", func(c *Case) { c.Fluid.MuO = -1 }},
		{"水 Corey 指数为负", func(c *Case) { c.RelPerm.Nw = -2 }},
		{"油 Corey 指数为 0", func(c *Case) { c.RelPerm.No = 0 }},
		{"油 Corey 指数等于 1", func(c *Case) { c.RelPerm.No = 1 }},
		{"krw0 为负", func(c *Case) { c.RelPerm.Krw0 = -0.1 }},
		{"kro0 为 0", func(c *Case) { c.RelPerm.Kro0 = 0 }},
		{"注入饱和度低于束缚水", func(c *Case) { c.Injection.SwInj = 0.1 }},
		{"注入饱和度超过 1−Sor", func(c *Case) { c.Injection.SwInj = 0.9 }},
		{"显式给出孔隙度 0", func(c *Case) { z := 0.0; c.Rock.Porosity = &z }},
		{"孔隙度超过 1", func(c *Case) { z := 1.2; c.Rock.Porosity = &z }},
	}
	for _, b := range bad {
		c := validCase()
		b.mutate(c)
		err := Validate(c)
		if err == nil {
			t.Errorf("%s: 应被校验拒绝，但校验通过", b.name)
		}
	}
}

func TestValidateRejectsMissingCoreFields(t *testing.T) {
	c := validCase()
	c.RelPerm = RelPermSpec{}
	err := Validate(c)
	if err == nil {
		t.Fatal("缺 relperm 的算例应被拒绝")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("Validate 应返回 *ValidationError，得到 %T", err)
	}
	if len(ve.Issues) == 0 {
		t.Fatal("ValidationError 不应为空")
	}
}
