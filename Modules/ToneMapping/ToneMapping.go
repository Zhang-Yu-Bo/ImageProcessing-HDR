package ToneMapping

import (
	"ImageProcessing_HDR/Modules/HDR/Common"
	"fmt"
	"image"
	"image/color"
	"math"
	"runtime"
	"sync"
	"time"
)

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

func CalculateGlobalLumAvg() {
	Common.GlobalLumAvg = 0.0
	numOfPixels := float64(Common.WidthOfImage) * float64(Common.HeightOfImage)

	for i := 0; i < Common.WidthOfImage; i++ {
		for j := 0; j < Common.HeightOfImage; j++ {
			Common.GlobalLumAvg += Common.GetGrayValue(
				Common.RadianceE[Common.ColorRed][i][j],
				Common.RadianceE[Common.ColorGreen][i][j],
				Common.RadianceE[Common.ColorBlue][i][j],
			) / numOfPixels
		}
	}
}

// Generate OriginLum, LumMatrix and LumWhite By RadianceE
func GenerateLumByRadianceE(a float64) {
	deltaAvoidSingular := 0.00001
	numOfPixels := float64(Common.WidthOfImage) * float64(Common.HeightOfImage)

	// init
	Common.LumWhite = 0.0
	Common.LumMatrix = Common.CreateSpace2D(Common.HeightOfImage, Common.WidthOfImage)

	// Fill LumMatrix
	for i := 0; i < Common.WidthOfImage; i++ {
		for j := 0; j < Common.HeightOfImage; j++ {
			Common.LumMatrix[i][j] = Common.GetGrayValue(
				Common.RadianceE[Common.ColorRed][i][j],
				Common.RadianceE[Common.ColorGreen][i][j],
				Common.RadianceE[Common.ColorBlue][i][j],
			)
			Common.LumWhite += math.Log(deltaAvoidSingular + Common.LumMatrix[i][j])
		}
	}

	// LumWhite
	Common.LumWhite = math.Exp(Common.LumWhite / numOfPixels)
	fmt.Println("Luminance White:", Common.LumWhite)

	// OriginLum
	Common.OriginLum = [][]float64{}
	// x, col
	for i := 0; i < Common.WidthOfImage; i++ {
		var temp []float64
		// y, row
		for j := 0; j < Common.HeightOfImage; j++ {
			temp = append(temp, (a/Common.LumWhite)*Common.LumMatrix[i][j])
		}
		Common.OriginLum = append(Common.OriginLum, temp)
	}
}

func GenerateLocalLumAvgMatrix(alpha, ratio, epsilon, phi, a float64) {
	fmt.Println("Generate Local LumAvg Matrix Begin")
	nowTime := time.Now()

	// init
	runtime.GOMAXPROCS(runtime.NumCPU())
	var wg sync.WaitGroup
	Common.LocalLumMatrix = Common.CreateSpace2D(Common.HeightOfImage, Common.WidthOfImage)
	alpha1 := alpha
	alpha2 := ratio * alpha1

	kernelSize := 51
	R1 := Common.CreateSpace3D(kernelSize, kernelSize, 9)
	R2 := Common.CreateSpace3D(kernelSize, kernelSize, 9)
	// x, col
	for i := 0; i < kernelSize; i++ {
		// y, row
		for j := 0; j < kernelSize; j++ {
			// z, depth
			for k := 1; k < 10; k++ {
				ss := math.Pow(ratio, float64(k-1))
				R1[i][j][k-1] = (1 / (math.Pi * math.Pow(alpha1*ss, 2))) *
					(math.Exp(-(float64((i-kernelSize/2)*(i-kernelSize/2)) +
						float64((j-kernelSize/2)*(j-kernelSize/2))) / math.Pow(alpha1*ss, 2)))
				R2[i][j][k-1] = (1 / (math.Pi * math.Pow(alpha2*ss, 2))) *
					(math.Exp(-(float64((i-kernelSize/2)*(i-kernelSize/2)) +
						float64((j-kernelSize/2)*(j-kernelSize/2))) / math.Pow(alpha2*ss, 2)))
			}
		}
	}

	V := Common.CreateSpace3D(Common.HeightOfImage, Common.WidthOfImage, 9)
	V1 := Common.CreateSpace3D(Common.HeightOfImage, Common.WidthOfImage, 9)
	V2 := Common.CreateSpace3D(Common.HeightOfImage, Common.WidthOfImage, 9)
	// z, depth
	for k := 0; k < 9; k++ {
		wg.Add(1)
		go func(k int) {
			defer wg.Done()
			ss := math.Pow(ratio, float64(k))
			// x, col
			for i := 0; i < Common.WidthOfImage; i++ {
				// y, row
				for j := 0; j < Common.HeightOfImage; j++ {
					r1Sum := 0.0
					r2Sum := 0.0
					for m :=0; m < kernelSize; m++{
						for n := 0; n < kernelSize; n++ {
							var newX, newY = m - kernelSize/2 + i, n - kernelSize/2 + j
							if newX < 0 {
								newX = 0
							} else if newX >= Common.WidthOfImage {
								newX = Common.WidthOfImage - 1
							}
							if newY < 0 {
								newY = 0
							}else if newY >= Common.HeightOfImage {
								newY = Common.HeightOfImage - 1
							}
							r1Sum += R1[m][n][k] * Common.OriginLum[newX][newY]
							r2Sum += R2[m][n][k] * Common.OriginLum[newX][newY]
						}
					}
					V1[i][j][k] = r1Sum
					V2[i][j][k] = r2Sum
					V[i][j][k] = (V1[i][j][k] - V2[i][j][k]) /
						(math.Pow(2, phi)*(a/math.Pow(ss, 2)) + V1[i][j][k])
				}
			}
		}(k)
	}
	wg.Wait()

	for i := 0; i < Common.WidthOfImage; i++ {
		for j := 0; j < Common.HeightOfImage; j++ {
			ss := 1.0
			for k := 1; k < 10;k++ {
				if math.Abs(V[i][j][k-1]) < epsilon {
					break
				}
				if k != 9 {
					ss *= ratio
				}
			}
			p := 1 + math.Round(math.Log(ss)/math.Log(ratio))
			if p > 8 {
				p = 8
			}
			Common.LocalLumMatrix[i][j] = Common.OriginLum[i][j]/(1+V1[i][j][int(p)])
		}
	}
	fmt.Println("Generate Local LumAvg Matrix End:", time.Now().Sub(nowTime))
}

func process(color float64, tmoAction Common.TmoAction, tmoType Common.TmoType, i, j int) float64 {
	var lumAvg float64
	if tmoAction == Common.LocalToneMapping {
		lumAvg = Common.LocalLumMatrix[i][j] / Common.LumMatrix[i][j]
	} else {
		// default Global
		lumAvg = Common.GlobalLumAvg
	}

	if tmoType == Common.Reinhard {
		if tmoAction == Common.LocalToneMapping {
			color = Reinhard(color, Common.LumMatrix[i][j], Common.LocalLumMatrix[i][j])
		} else {
			// default Global
			color = Reinhard(color, Common.GlobalLumAvg, 1)
		}
	} else if tmoType == Common.CE {
		color = CE(color, lumAvg)
	} else if tmoType == Common.Uncharted2 {
		color = Uncharted2(color, lumAvg, 11.2)
	} else if tmoType == Common.ReinhardEnhance && tmoAction == Common.LocalToneMapping {
		color = (color / Common.LumMatrix[i][j]) * Common.LocalLumMatrix[i][j]
	} else {
		// default ACES
		color = ACES(color, lumAvg)
	}
	return color
}

func GenerateLdrImage(tmoAction Common.TmoAction, tmoType Common.TmoType) {
	Common.OutputImage = image.NewRGBA(image.Rect(0, 0, Common.WidthOfImage, Common.HeightOfImage))

	for i := 0; i < Common.WidthOfImage; i++ {
		for j := 0; j < Common.HeightOfImage; j++ {
			// Tone mapping
			colorR := process(Common.RadianceE[Common.ColorRed][i][j], tmoAction, tmoType, i, j) * 255
			colorG := process(Common.RadianceE[Common.ColorGreen][i][j], tmoAction, tmoType, i, j) * 255
			colorB := process(Common.RadianceE[Common.ColorBlue][i][j], tmoAction, tmoType, i, j) * 255
			// clipping
			colorR = Common.Clipping(colorR)
			colorG = Common.Clipping(colorG)
			colorB = Common.Clipping(colorB)
			// Output
			Common.OutputImage.Set(i, j, color.RGBA{
				R: uint8(colorR),
				G: uint8(colorG),
				B: uint8(colorB),
				A: 255,
			})
		}
	}
}
