package ToneMapping

import "math"

func ACES(color float64, lumAvg float64) float64 {
	const A = 2.51
	const B = 0.03
	const C = 2.43
	const D = 0.59
	const E = 0.14
	color *= lumAvg
	color = (color * (A*color + B)) / (color*(C*color+D) + E)
	return color
}

// [middleGrey] 為自定義變數: 甚麼值是灰色，越大通常整體越亮
func Reinhard(color float64, lumAvg float64, middleGrey float64) float64 {
	color *= middleGrey / lumAvg
	return color / (1.0 + color)
}

func CE(color float64, lumAvg float64) float64 {
	return 1 - math.Exp(-lumAvg*color)
}

func uncharted2Function(x float64) float64 {
	const A = 0.22
	const B = 0.30
	const C = 0.10
	const D = 0.20
	const E = 0.01
	const F = 0.30
	return ((x*(A*x+C*B) + D*E) / (x*(A*x+B) + D*F)) - E/F
}

// [white] 為自定義變數: 甚麼值是白色，越大通常整體越暗
func Uncharted2(color float64, lumAvg float64, white float64) float64 {
	return uncharted2Function(1.6 * lumAvg * color) / uncharted2Function(white)
}
