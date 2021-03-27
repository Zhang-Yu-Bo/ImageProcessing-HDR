package DebevecMalik

import (
	"ImageProcessing_HDR/Modules/HDR/Common"
	"fmt"
	"gonum.org/v1/gonum/mat"
	"math"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

var (
	samplePixels 	[]Common.Vec2
	// functionGz: 0 -> R, 1 -> G, 2 -> B
	// functionGz[]: g(Z0) g(Z1) ... g(Z255)
	functionGz		[][]float64
)

func weightValue(value float64) float64 {
	if value <= 128 {
		return value
	}
	return 255 - value
}

func PixelSampling(N int) {
	fmt.Println("DebevecMalik Sampling Begin")
	nowTime := time.Now()
	// 隨機取樣
	//rand.Seed(time.Now().UnixNano())
	//samplePixels = []Vec2{}
	//for i := 0; i < NumOfSamples; i++ {
	//	randVec2 := Vec2{
	//		X: rand.Int() % WidthOfImage,
	//		Y: rand.Int() % HeightOfImage,
	//	}
	//	fmt.Println(randVec2)
	//	samplePixels = append(samplePixels, randVec2)
	//}

	// 間隔(平均)取樣
	samplePixels = []Common.Vec2{}
	gridWidth := Common.WidthOfImage / int(math.Sqrt(float64(N)))
	gridHeight := Common.HeightOfImage / int(math.Sqrt(float64(N)))
	samplesTemp := 0
	for i := gridWidth; i < Common.WidthOfImage; i += gridWidth {
		for j := gridHeight; j < Common.HeightOfImage; j += gridHeight {
			sample := Common.Vec2{X: i, Y: j}
			samplePixels = append(samplePixels, sample)
			samplesTemp++
		}
	}
	// 不足的用隨機取點的方式補
	rand.Seed(time.Now().UnixNano())
	for ; samplesTemp < N; samplesTemp++ {
		randVec2 := Common.Vec2{
			X: rand.Int() % Common.WidthOfImage,
			Y: rand.Int() % Common.HeightOfImage,
		}
		fmt.Println(randVec2)
		samplePixels = append(samplePixels, randVec2)
	}
	fmt.Println("DebevecMalik Sampling End:", time.Now().Sub(nowTime))
}

func GenerateFunctionGz() error {
	fmt.Println("DebevecMalik Generate Function g(Zij) Begin")
	nowTime := time.Now()
	functionGz = [][]float64{}

	// 建立三個色彩通道的matrix以恢復各別的 g(Zij) 曲線
	for c := 0; c < 3; c++ {
		n := 256
		matrixA := mat.NewDense(Common.NumOfSamplePixels*Common.NumOfImages+n+1, n+Common.NumOfSamplePixels, nil)
		matrixB := mat.NewDense(Common.NumOfSamplePixels*Common.NumOfImages+n+1, 1, nil)

		k := 0
		for i := 0; i < Common.NumOfSamplePixels; i++ {
			for j := 0; j < Common.NumOfImages; j++ {
				var ans uint32
				switch c {
				case Common.ColorRed:
					r, _, _, _ := Common.DataOfImages[j].At(samplePixels[i].X, samplePixels[i].Y).RGBA()
					ans = r
					break
				case Common.ColorGreen:
					_, g, _, _ := Common.DataOfImages[j].At(samplePixels[i].X, samplePixels[i].Y).RGBA()
					ans = g
					break
				case Common.ColorBlue:
					_, _, b, _ := Common.DataOfImages[j].At(samplePixels[i].X, samplePixels[i].Y).RGBA()
					ans = b
					break
				}
				ans = ans >> 8
				wValue := weightValue(float64(ans) + 1)
				//wValue := weightValue(float64(ans))
				matrixA.Set(k, int(ans), wValue)
				matrixA.Set(k, n+i, -wValue)
				matrixB.Set(k, 0, wValue*math.Log(Common.ExposureTimes[j]))
				k++
			}
		}
		matrixA.Set(k, 128, 1)
		k++
		for i := 0; i < n-2; i++ {
			wValue := weightValue(float64(i) + 1)
			//wValue := weightValue(float64(i))
			matrixA.Set(k, i, wValue)
			matrixA.Set(k, i+1, -2*wValue)
			matrixA.Set(k, i+2, wValue)
			k++
		}

		// Ax = b
		// At * A * x = At * b
		// x = (At * A)^-1 * At * b
		var matrixAtA mat.Dense
		matrixAtA.Mul(matrixA.T(), matrixA)
		var inverseMatrix mat.Dense
		err := inverseMatrix.Inverse(&matrixAtA)
		if err != nil {
			return err
		}
		var matrixInverseMulAt mat.Dense
		matrixInverseMulAt.Mul(&inverseMatrix, matrixA.T())
		var X mat.Dense
		X.Mul(&matrixInverseMulAt, matrixB)

		var tempFunctionGz []float64
		for i := 0; i < 256; i++ {
			tempFunctionGz = append(tempFunctionGz, X.At(i, 0))
			//fmt.Println(X.At(i, 0))
		}
		functionGz = append(functionGz, tempFunctionGz)

		// 速度很慢
		//var svd mat.SVD
		//svd.Factorize(matrixA, mat.SVDFull)
		//const rcond = 1e-15
		//rank := svd.Rank(rcond)
		//var x mat.Dense
		//svd.SolveTo(&x, matrixB, rank)
		//var tempFunctionGz []float64
		//for i := 0; i < 256; i++ {
		//	tempFunctionGz = append(tempFunctionGz, x.At(i, 0))
		//}
		//functionGz = append(functionGz, tempFunctionGz)
	}
	fmt.Println("DebevecMalik Generate Function g(Zij) End:", time.Now().Sub(nowTime))
	return nil
}

func CalculateRadianceE() {
	fmt.Println("DebevecMalik Generate RadianceE Begin")
	nowTime := time.Now()
	Common.RadianceE = [][][]float64{}

	// initialization [Common.RadianceE]
	for c:=0; c < 3; c++ {
		var tempRadianceSlice [][]float64
		for i := 0; i < Common.WidthOfImage; i++ {
			var tempE []float64
			for j := 0; j < Common.HeightOfImage; j++ {
				tempE = append(tempE, 0)
			}
			tempRadianceSlice = append(tempRadianceSlice, tempE)
		}
		Common.RadianceE = append(Common.RadianceE, tempRadianceSlice)
	}

	// set max process to work
	runtime.GOMAXPROCS(runtime.NumCPU())
	// check calculate is done
	wg := sync.WaitGroup{}
	// 透過各別的 g(Zij) 去計算出真實場景的能量分布
	for c := 0; c < 3; c++ {
		wg.Add(1)
		go func(channelIndex int) {
			defer wg.Done()
			minValue := math.MaxFloat64
			maxValue := 0.0
			for i := 0; i < Common.WidthOfImage; i++ {
				for j := 0; j < Common.HeightOfImage; j++ {
					var sumOfRadiance float64
					var sumOfWeight float64
					for p := 0; p < Common.NumOfImages; p++ {
						var ans uint32
						switch channelIndex {
						case Common.ColorRed:
							r, _, _, _ := Common.DataOfImages[p].At(i, j).RGBA()
							ans = r
							break
						case Common.ColorGreen:
							_, g, _, _ := Common.DataOfImages[p].At(i, j).RGBA()
							ans = g
							break
						case Common.ColorBlue:
							_, _, b, _ := Common.DataOfImages[p].At(i, j).RGBA()
							ans = b
							break
						}
						ans = ans >> 8
						wValue := weightValue(float64(ans))
						// sigma( (g(Zij) - ln(Tj)) * weight ) for j = 0 to NumOfImages
						sumOfRadiance += wValue * (functionGz[channelIndex][ans] - math.Log(Common.ExposureTimes[p]))
						sumOfWeight += wValue
					}
					if sumOfWeight == 0 {
						sumOfWeight = 1
					}
					tempRadianceE := math.Pow(math.E, sumOfRadiance/sumOfWeight)
					Common.RadianceE[channelIndex][i][j] = tempRadianceE
					if tempRadianceE < minValue {
						minValue = tempRadianceE
					}
					if tempRadianceE > maxValue {
						maxValue = tempRadianceE
					}
				}
			}
			fmt.Println("[", maxValue, minValue, "]")
		}(c)
	}
	// if calculate is not over, then wait
	wg.Wait()
	fmt.Println("DebevecMalik Generate RadianceE End:", time.Now().Sub(nowTime))
}
