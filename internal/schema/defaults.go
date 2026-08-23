package schema

// DefaultPorosity 是未显式给出孔隙度时采用的默认值，与常规砂岩水驱
// 一致。孔隙度只用于把无因次速度换算成注入孔隙体积倍数。
const DefaultPorosity = 0.25

// ApplyDefaults 为省略的可选字段填入默认值。当前唯一可选字段是
// rock.porosity：只有未显式给出（nil）时才填默认值；显式给出 0 会
// 保留并触发校验错误，绝不静默“修正”用户输入。
func ApplyDefaults(c *Case) {
	if c.Rock.Porosity == nil {
		d := DefaultPorosity
		c.Rock.Porosity = &d
	}
}
