package ToneMapping

func ToonMappingACES(color float64, lum float64) float64 {
	const A = 2.51
	const B = 0.03
	const C = 2.43
	const D = 0.59
	const E = 0.14
	color *= lum
	color = (color * (A*color + B)) / (color*(C*color+D) + E)
	// clipping
	if color > 1 {
		color = 1
	} else if color < 0 {
		color = 0
	}
	return color
}
