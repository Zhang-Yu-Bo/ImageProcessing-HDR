package Common

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"

	"github.com/nfnt/resize"
	"gocv.io/x/gocv"

	_ "image/jpeg"
)

type Vec2 struct {
	X int
	Y int
}

const (
	ColorRed   = 0
	ColorGreen = 1
	ColorBlue  = 2
)

type TmoAction int

const LocalToneMapping TmoAction = 3
const GlobalToneMapping TmoAction = 4

type TmoType int

const Aces TmoType = 5
const Reinhard TmoType = 6
const CE TmoType = 7
const Uncharted2 TmoType = 8
const ReinhardEnhance TmoType = 9

var (
	// 總共取幾點去計算 g(Zij)
	NumOfSamplePixels int
	NumOfImages       int
	WidthOfImage      int
	HeightOfImage     int
	DataOfImages      []image.Image
	ExposureTimes     []float64
	// RadianceE: 0 -> R, 1 -> G, 2 -> B
	// RadianceE[]: x(width, col) RadianceE[][]: y(height, row)
	RadianceE [][][]float64
	// LumMatrix[]: x(width, col) LumMatrix[][]: y(height, row)
	LumMatrix [][]float64
	LumWhite  float64
	// OriginLum[]: x(width, col) OriginLum[][]: y(height, row)
	OriginLum [][]float64
	// LocalLumMatrix[]: x(width, col) LocalLumMatrix[][]: y(height, row)
	LocalLumMatrix [][]float64
	GlobalLumAvg   float64
	OutputImage    *image.RGBA
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

// clip the color which bigger than 255, or smaller than 0
func Clipping(color float64) float64 {
	if color > 255 {
		color = 255
	} else if color < 0 {
		color = 0
	}
	return color
}

func SaveAsPng() error {
	var err error
	var outputFile *os.File

	if outputFile, err = os.Create("output.png"); err != nil {
		return err
	}
	defer CloseFile(outputFile)

	if err = png.Encode(outputFile, OutputImage); err != nil {
		return err
	}
	fmt.Println("Create Ldr Image:", "output.png")
	return nil
}

func SaveAsHdrFormat() {
	hdrImage := gocv.NewMatWithSize(HeightOfImage, WidthOfImage, gocv.MatTypeCV32FC3)
	for i := 0; i < WidthOfImage; i++ {
		for j := 0; j < HeightOfImage; j++ {
			// Blue
			hdrImage.SetFloatAt(j, i*3+0, float32(RadianceE[ColorBlue][i][j]))
			// Green
			hdrImage.SetFloatAt(j, i*3+1, float32(RadianceE[ColorGreen][i][j]))
			// Red
			hdrImage.SetFloatAt(j, i*3+2, float32(RadianceE[ColorRed][i][j]))
		}
	}
	gocv.IMWrite("output.hdr", hdrImage)
	fmt.Println("Create Hdr Image:", "output.hdr")
	gocv.IMWrite("output.exr", hdrImage)
	fmt.Println("Create Hdr Image:", "output.exr")
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

func ConvertImageToGocvMat(img image.Image) gocv.Mat {
	gocvMatFormatImg := gocv.NewMatWithSize(img.Bounds().Dy(), img.Bounds().Dx(), gocv.MatTypeCV32SC3)
	for i := 0; i < img.Bounds().Dx(); i++ {
		for j := 0; j < img.Bounds().Dy(); j++ {
			r, g, b, _ := img.At(i, j).RGBA()
			// Blue
			gocvMatFormatImg.SetIntAt(j, i*3+0, int32(b>>8))
			// Green
			gocvMatFormatImg.SetIntAt(j, i*3+1, int32(g>>8))
			// Red
			gocvMatFormatImg.SetIntAt(j, i*3+2, int32(r>>8))
		}
	}
	return gocvMatFormatImg
}

func Gradient(img image.Image) [][]float64 {

	gradient := CreateSpace2D(img.Bounds().Dy(), img.Bounds().Dx())
	for i := 0; i < img.Bounds().Dx()-1; i++ {
		for j := 0; j < img.Bounds().Dy()-1; j++ {
			r00, g00, b00, _ := img.At(i, j).RGBA()
			r01, g01, b01, _ := img.At(i, j+1).RGBA()
			r10, g10, b10, _ := img.At(i+1, j).RGBA()
			gray00 := int(((r00 >> 8) + (g00 >> 8) + (b00 >> 8)) / 3)
			gray01 := int(((r01 >> 8) + (g01 >> 8) + (b01 >> 8)) / 3)
			gray10 := int(((r10 >> 8) + (g10 >> 8) + (b10 >> 8)) / 3)
			gradient[i][j] = math.Abs(float64(gray10-gray00)) + math.Abs(float64(gray01-gray00))
		}
	}
	return gradient
}

func MTB() {
	fmt.Println("MTB Begin")
	var moveX, moveY []int
	for j := 0; j < NumOfImages; j++ {
		moveX = append(moveX, 0)
		moveY = append(moveY, 0)
	}
	minLen := WidthOfImage
	if HeightOfImage < WidthOfImage {
		minLen = HeightOfImage
	}
	times := int(math.Log2(float64(minLen))) - 4
	fmt.Println("MTB do ", times, " times")
	for i := 0; i < times; i++ {
		scale := int(math.Pow(2, float64(times-i-1)))
		imageNewWidth := uint(WidthOfImage / scale)
		for j := 1; j < NumOfImages; j++ {
			img1 := resize.Resize(imageNewWidth, 0, DataOfImages[j-1], resize.Lanczos3)
			img2 := resize.Resize(imageNewWidth, 0, DataOfImages[j], resize.Lanczos3)
			g1 := Gradient(img1)
			g2 := Gradient(img2)

			minX, minY, minDis := 0, 0, float64(100000000)

			for x := -1; x <= 1; x++ {
				for y := -1; y <= 1; y++ {
					dis := float64(0)
					count := 0
					for ii := 1; ii < img1.Bounds().Dx()-1; ii++ {
						for jj := 1; jj < img1.Bounds().Dy()-1; jj++ {
							g2i, g2j, g1i, g1j := ii+x+moveX[j]/scale, jj+y+moveY[j]/scale, ii, jj
							if g2i < 0 || g2j < 0 || g1i < 0 || g1j < 0 || g2i >= img1.Bounds().Dx() || g2j >= img1.Bounds().Dy() || g1i >= img2.Bounds().Dx() || g1j >= img2.Bounds().Dy() {
								continue
							}
							dis += math.Pow(g2[g2i][g2j]-g1[g1i][g1j], 2)
							count += 1
						}
					}
					dis /= float64(count)
					//fmt.Println(dis, count)
					if dis < minDis {
						minX, minY = x, y
						minDis = dis
					}
				}
			}
			minX *= scale
			minY *= scale
			moveX[j] += minX
			moveY[j] += minY

		}
	}

	moveXSum, moveYSum := 0, 0
	for j := 1; j < NumOfImages; j++ {
		movedImage := image.NewRGBA(image.Rect(0, 0, WidthOfImage, HeightOfImage))
		moveXSum += moveX[j]
		moveYSum += moveY[j]
		for ii := 0; ii < WidthOfImage; ii++ {
			for jj := 0; jj < HeightOfImage; jj++ {
				oriX, oriY := ii+moveXSum, jj+moveYSum
				if oriX < 0 || oriY < 0 || oriX >= WidthOfImage || oriY >= HeightOfImage {
					continue
				}
				movedImage.Set(ii, jj, DataOfImages[j].At(oriX, oriY))
			}
		}
		DataOfImages[j] = movedImage
	}
	fmt.Println("MTB End")
}
