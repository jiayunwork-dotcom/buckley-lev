package schema

const DefaultPorosity = 0.25

func ApplyDefaults(c *Case) {
	if c.Rock.Porosity == nil {
		d := DefaultPorosity
		c.Rock.Porosity = &d
	}
}
