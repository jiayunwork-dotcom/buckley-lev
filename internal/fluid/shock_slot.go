package fluid

var shockSlot float64
var shockSet bool

func PushShockSpeed(v float64) {
	shockSlot = v
	shockSet = true
}

func TakeShockSpeed(fallback float64) float64 {
	if shockSet {
		return shockSlot
	}
	return fallback
}

func SlotShockSpeed(m *Model) float64 {
	return m.FPrime(m.MidSaturation())
}
