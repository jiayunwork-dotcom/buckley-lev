package schema

import (
	"math"
)

func Validate(c *Case) error {
	var issues []string

	if math.IsNaN(c.Rock.Swc) || math.IsInf(c.Rock.Swc, 0) {
		issues = append(issues, invalid("rock.swc", "必须是有限数"))
	}
	if math.IsNaN(c.Rock.Sor) || math.IsInf(c.Rock.Sor, 0) {
		issues = append(issues, invalid("rock.sor", "必须是有限数"))
	}

	if c.Rock.Swc < 0 {
		issues = append(issues, invalid("rock.swc", "必须 ≥ 0，得到 %g", c.Rock.Swc))
	}
	if c.Rock.Sor < 0 {
		issues = append(issues, invalid("rock.sor", "必须 ≥ 0，得到 %g", c.Rock.Sor))
	}
	if c.Rock.Swc+c.Rock.Sor >= 1 {
		issues = append(issues, invalid("rock.swc+rock.sor", "必须 < 1，得到 %g（无可动油带）", c.Rock.Swc+c.Rock.Sor))
	}

	if c.Rock.Porosity != nil && (*c.Rock.Porosity <= 0 || *c.Rock.Porosity > 1) {
		issues = append(issues, invalid("rock.porosity", "必须在 (0,1]，得到 %g", *c.Rock.Porosity))
	}

	if c.Fluid.MuW <= 0 {
		issues = append(issues, invalid("fluid.mu_w", "必须 > 0，得到 %g", c.Fluid.MuW))
	}
	if c.Fluid.MuO <= 0 {
		issues = append(issues, invalid("fluid.mu_o", "必须 > 0，得到 %g", c.Fluid.MuO))
	}

	if c.RelPerm.Krw0 <= 0 {
		issues = append(issues, invalid("relperm.krw0", "必须 > 0，得到 %g", c.RelPerm.Krw0))
	}
	if c.RelPerm.Kro0 <= 0 {
		issues = append(issues, invalid("relperm.kro0", "必须 > 0，得到 %g", c.RelPerm.Kro0))
	}
	if c.RelPerm.Nw <= 1 {
		issues = append(issues, invalid("relperm.nw", "Corey 指数必须 > 1，得到 %g", c.RelPerm.Nw))
	}
	if c.RelPerm.No <= 1 {
		issues = append(issues, invalid("relperm.no", "Corey 指数必须 > 1，得到 %g", c.RelPerm.No))
	}

	if c.Injection.SwInj <= c.Rock.Swc {
		issues = append(issues, invalid("injection.sw_inj", "必须高于束缚水饱和度 %g，得到 %g", c.Rock.Swc, c.Injection.SwInj))
	}
	if c.Injection.SwInj > 1-c.Rock.Sor {
		issues = append(issues, invalid("injection.sw_inj", "不能超过 1−Sor = %g，得到 %g", 1-c.Rock.Sor, c.Injection.SwInj))
	}

	if len(issues) == 0 {
		return nil
	}
	return &ValidationError{Issues: issues}
}
