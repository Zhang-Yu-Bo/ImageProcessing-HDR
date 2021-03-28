package Common

import (
	"ImageProcessing_HDR/Modules/ToneMapping"
	"errors"
	"fmt"
	"gocv.io/x/gocv"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"runtime"
	"sync"
	"time"
)

type Vec2 struct {
	X int
	Y int
}

const (
	ColorRed			= 0
	ColorGreen			= 1
	ColorBlue			= 2
)

type TmpAction int
const LocalToneMapping 	TmpAction = 3
const GlobalToneMapping TmpAction = 4

type TmpType int
const Aces 				TmpType = 5
const Reinhard 			TmpType = 6
const CE 				TmpType = 7
const Uncharted2 		TmpType = 8
const ReinhardEnhance	TmpType = 9

var (
	// 總共取幾點去計算 g(Zij)
	NumOfSamplePixels	int
	NumOfImages 		int
	WidthOfImage  		int
	HeightOfImage 		int
	DataOfImages		[]image.Image
	ExposureTimes		[]float64
	// RadianceE: 0 -> R, 1 -> G, 2 -> B
	// RadianceE[]: x(width, col) RadianceE[][]: y(height, row)
	RadianceE			[][][]float64
	// LumMatrix[]: x(width, col) LumMatrix[][]: y(height, row)
	LumMatrix 			[][]float64
	LumWhite  			float64
	// OriginLum[]: x(width, col) OriginLum[][]: y(height, row)
	OriginLum   		[][]float64
	// LocalLumMatrix[]: x(width, col) LocalLumMatrix[][]: y(height, row)
	LocalLumMatrix 		[][]float64
	GlobalLumAvg		float64
	outputImage 		*image.RGBA
)

func LoadImageFiles(fileName []string, exposureTime []float64) error {
	DataOfImages = []image.Image{}
	ExposureTimes = []float64{}

	for i := 0; i < NumOfImages; i++ {
		var imageFile *os.File
		var err error
		imageFile, err = os.Open(fileName[i])
		if err != nil {
			return err
		}

		var tempImageData image.Image
		tempImageData, _, err = image.Decode(imageFile)
		if err != nil {
			return err
		}
		if i == 0 {
			WidthOfImage = tempImageData.Bounds().Dx()
			HeightOfImage = tempImageData.Bounds().Dy()
		} else if WidthOfImage != tempImageData.Bounds().Dx() ||
			HeightOfImage != tempImageData.Bounds().Dy() {
			return errors.New("width of images or height of images is not equal")
		}
		DataOfImages = append(DataOfImages, tempImageData)
		ExposureTimes = append(ExposureTimes, exposureTime[i])

		CloseFile(imageFile)
	}
	return nil
}

func GetGrayValue(red, green, blue float64) float64 {
	return 0.299*red + 0.587*green + 0.114*blue
}

func CalculateGlobalLumAvg() {
	GlobalLumAvg = 0.0
	numOfPixels := float64(WidthOfImage) * float64(HeightOfImage)

	for i := 0; i < WidthOfImage; i++ {
		for j := 0; j < HeightOfImage; j++ {
			GlobalLumAvg += GetGrayValue(
				RadianceE[ColorRed][i][j],
				RadianceE[ColorGreen][i][j],
				RadianceE[ColorBlue][i][j],
			) / numOfPixels
		}
	}
}

// Generate OriginLum, LumMatrix and LumWhite By RadianceE
func GenerateLumByRadianceE(a float64) {
	deltaAvoidSingular := 0.00001
	numOfPixels := float64(WidthOfImage) * float64(HeightOfImage)

	// init
	LumWhite = 0.0
	LumMatrix = CreateSpace2D(HeightOfImage, WidthOfImage)

	// Fill LumMatrix
	for i := 0; i < WidthOfImage; i++ {
		for j := 0; j < HeightOfImage; j++ {
			LumMatrix[i][j] = GetGrayValue(
				RadianceE[ColorRed][i][j],
				RadianceE[ColorGreen][i][j],
				RadianceE[ColorBlue][i][j],
			)
			LumWhite += math.Log(deltaAvoidSingular + LumMatrix[i][j])
		}
	}

	// LumWhite
	LumWhite = math.Exp(LumWhite / numOfPixels)
	fmt.Println("Luminance White:", LumWhite)

	// OriginLum
	OriginLum = [][]float64{}
	// x, col
	for i := 0; i < WidthOfImage; i++ {
		var temp []float64
		// y, row
		for j := 0; j < HeightOfImage; j++ {
			temp = append(temp, (a/LumWhite)*LumMatrix[i][j])
		}
		OriginLum = append(OriginLum, temp)
	}
}

func GenerateLocalLumAvgMatrix(alpha, ratio, epsilon, phi, a float64) {
	fmt.Println("Generate Local LumAvg Matrix Begin")
	nowTime := time.Now()

	// init
	runtime.GOMAXPROCS(runtime.NumCPU())
	var wg sync.WaitGroup
	LocalLumMatrix = CreateSpace2D(HeightOfImage, WidthOfImage)
	alpha1 := alpha
	alpha2 := ratio * alpha1

	kernelSize := 51
	R1 := CreateSpace3D(kernelSize, kernelSize, 9)
	R2 := CreateSpace3D(kernelSize, kernelSize, 9)

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

	V := CreateSpace3D(HeightOfImage, WidthOfImage, 9)
	V1 := CreateSpace3D(HeightOfImage, WidthOfImage, 9)
	V2 := CreateSpace3D(HeightOfImage, WidthOfImage, 9)
	// z, depth
	for k := 0; k < 9; k++ {
		wg.Add(1)
		go func(k int) {
			defer wg.Done()
			ss := math.Pow(ratio, float64(k))
			// x, col
			for i := 0; i < WidthOfImage; i++ {
				// y, row
				for j := 0; j < HeightOfImage; j++ {
					r1Sum := 0.0
					r2Sum := 0.0
					for m :=0; m < kernelSize; m++{
						for n := 0; n < kernelSize; n++ {
							var newX, newY = m - kernelSize/2 + i, n - kernelSize/2 + j
							if newX < 0 {
								newX = 0
							} else if newX >= WidthOfImage {
								newX = WidthOfImage - 1
							}
							if newY < 0 {
								newY = 0
							}else if newY >= HeightOfImage {
								newY = HeightOfImage - 1
							}
							r1Sum += R1[m][n][k] * OriginLum[newX][newY]
							r2Sum += R2[m][n][k] * OriginLum[newX][newY]
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

	for i := 0; i < WidthOfImage; i++ {
		for j := 0; j < HeightOfImage; j++ {
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
			LocalLumMatrix[i][j] = OriginLum[i][j]/(1+V1[i][j][int(p)])
		}
	}
	fmt.Println("Generate Local LumAvg Matrix End:", time.Now().Sub(nowTime))
}

func TmpOperate(color float64, tmpAction TmpAction, tmpType TmpType, i, j int) float64 {
	var lumAvg float64
	if tmpAction == LocalToneMapping {
		lumAvg = LocalLumMatrix[i][j] / LumMatrix[i][j]
	} else { // default Global
		lumAvg = GlobalLumAvg
	}

	if tmpType == Reinhard {
		if tmpAction == LocalToneMapping {
			color = ToneMapping.Reinhard(color, LumMatrix[i][j], LocalLumMatrix[i][j])
		} else { // default Global
			color = ToneMapping.Reinhard(color, GlobalLumAvg, 1)
		}
	} else if tmpType == CE {
		color = ToneMapping.CE(color, lumAvg)
	} else if tmpType == Uncharted2 {
		color = ToneMapping.Uncharted2(color, lumAvg, 11.2)
	} else if tmpType == ReinhardEnhance && tmpAction == LocalToneMapping {
		color = (color / LumMatrix[i][j]) * LocalLumMatrix[i][j]
	} else { // default ACES
		color = ToneMapping.ACES(color, lumAvg)

	}
	return color
}

func GenerateLdrImage(tmpAction TmpAction, tmpType TmpType) {
	outputImage = image.NewRGBA(image.Rect(0, 0, WidthOfImage, HeightOfImage))

	for i := 0; i < WidthOfImage; i++ {
		for j := 0; j < HeightOfImage; j++ {
			// Tone mapping
			colorR := TmpOperate(RadianceE[ColorRed][i][j], tmpAction, tmpType, i, j) * 255
			colorG := TmpOperate(RadianceE[ColorGreen][i][j], tmpAction, tmpType, i, j) * 255
			colorB := TmpOperate(RadianceE[ColorBlue][i][j], tmpAction, tmpType, i, j) * 255
			// clipping
			colorR = Clipping(colorR)
			colorG = Clipping(colorG)
			colorB = Clipping(colorB)
			// Output
			outputImage.Set(i, j, color.RGBA{
				R: uint8(colorR),
				G: uint8(colorG),
				B: uint8(colorB),
				A: 255,
			})
		}
	}
}

// clip the color which bigger than 255, or smaller than 0
func Clipping(color float64) float64 {
	if color > 255 {
		color = 255
	} else if color < 0 {
		color = 0
	}
	return color
}

func SaveAsPng() error{
	var err error
	var outputFile *os.File

	if outputFile, err = os.Create("output.png"); err != nil {
		return err
	}
	defer CloseFile(outputFile)

	if err = png.Encode(outputFile, outputImage); err != nil {
		return err
	}
	fmt.Println("Create Image:", "output.png")
	return nil
}

func SaveAsHdrFormat() {
	hdrImage := gocv.NewMatWithSize(HeightOfImage, WidthOfImage, gocv.MatTypeCV32FC3)
	for i := 0; i < WidthOfImage; i++ {
		for j := 0; j < HeightOfImage; j++ {
			// Blue
			hdrImage.SetFloatAt(j, i * 3 + 0, float32(RadianceE[ColorBlue][i][j]))
			// Green
			hdrImage.SetFloatAt(j, i * 3 + 1, float32(RadianceE[ColorGreen][i][j]))
			// Red
			hdrImage.SetFloatAt(j, i * 3 + 2, float32(RadianceE[ColorRed][i][j]))
		}
	}
	gocv.IMWrite("output.hdr", hdrImage)
	gocv.IMWrite("output.exr", hdrImage)
}

func CloseFile(file *os.File) {
	if err := file.Close(); err != nil {
		fmt.Println(err.Error())
	}
}

func CreateSpace2D(rows, cols int) [][]float64 {
	var result [][]float64
	for i := 0; i < cols; i++ {
		var temp []float64
		for j := 0; j < rows; j++ {
			temp = append(temp, 0)
		}
		result = append(result, temp)
	}
	return result
}

func CreateSpace3D(rows, cols, depth int) [][][]float64 {
	var result [][][]float64
	for i := 0; i < cols; i++ {
		var temp [][]float64
		for j := 0; j < rows; j++ {
			var tempDepth []float64
			for k := 0; k < depth; k++ {
				tempDepth = append(tempDepth, 0)
			}
			temp = append(temp, tempDepth)
		}
		result = append(result, temp)
	}
	return result
}
