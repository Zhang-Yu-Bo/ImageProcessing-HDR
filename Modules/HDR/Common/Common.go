package Common

import (
	"errors"
	"fmt"
	"gocv.io/x/gocv"
	"image"
	"image/png"
	"os"
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

type TmoAction int
const LocalToneMapping 	TmoAction = 3
const GlobalToneMapping TmoAction = 4

type TmoType int
const Aces 				TmoType = 5
const Reinhard 			TmoType = 6
const CE 				TmoType = 7
const Uncharted2 		TmoType = 8
const ReinhardEnhance 	TmoType = 9

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
	OutputImage 		*image.RGBA
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

func SaveAsPng() error{
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
