package Common

import (
	"ImageProcessing_HDR/Modules/ToneMapping"
	"errors"
	"fmt"
	"gocv.io/x/gocv"
	"image"
	"image/color"
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

var (
	// 總共取幾點去計算 g(Zij)
	NumOfSamplePixels	int
	NumOfImages 		int
	WidthOfImage  		int
	HeightOfImage 		int
	DataOfImages		[]image.Image
	ExposureTimes		[]float64
	// RadianceE: 0 -> R, 1 -> G, 2 -> B
	// RadianceE[]: x(width) RadianceE[][]: y(height)
	RadianceE			[][][]float64
	outputImage			*image.RGBA
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

func CalculateLuminanceAvg() float64 {
	// Calculate the average luminance
	var lumAvg float64
	numOfPixels := float64(WidthOfImage) * float64(HeightOfImage)

	for i := 0; i < WidthOfImage; i++ {
		for j := 0; j < HeightOfImage; j++ {
			lumAvg += (0.299*RadianceE[ColorRed][i][j] +
				0.587*RadianceE[ColorGreen][i][j] +
				0.114*RadianceE[ColorBlue][i][j]) / numOfPixels
		}
	}
	fmt.Println("Luminance Average:", lumAvg)
	return lumAvg
}

func GenerateLdrImage() {
	outputImage = image.NewRGBA(image.Rect(0, 0, WidthOfImage, HeightOfImage))
	lumAvg := CalculateLuminanceAvg()

	for i := 0; i < WidthOfImage; i++ {
		for j := 0; j < HeightOfImage; j++ {
			colorR := ToneMapping.ACES(RadianceE[ColorRed][i][j], lumAvg) * 255
			colorG := ToneMapping.ACES(RadianceE[ColorGreen][i][j], lumAvg) * 255
			colorB := ToneMapping.ACES(RadianceE[ColorBlue][i][j], lumAvg) * 255
			//colorR := ToneMapping.Reinhard(RadianceE[ColorRed][i][j], lumAvg, 1.0) * 255
			//colorG := ToneMapping.Reinhard(RadianceE[ColorGreen][i][j], lumAvg, 1.0) * 255
			//colorB := ToneMapping.Reinhard(RadianceE[ColorBlue][i][j], lumAvg, 1.0) * 255
			//colorR := ToneMapping.CE(RadianceE[ColorRed][i][j], lumAvg) * 255
			//colorG := ToneMapping.CE(RadianceE[ColorGreen][i][j], lumAvg) * 255
			//colorB := ToneMapping.CE(RadianceE[ColorBlue][i][j], lumAvg) * 255
			//colorR := ToneMapping.Uncharted2(RadianceE[ColorRed][i][j], lumAvg, 11.2) * 255
			//colorG := ToneMapping.Uncharted2(RadianceE[ColorGreen][i][j], lumAvg, 11.2) * 255
			//colorB := ToneMapping.Uncharted2(RadianceE[ColorBlue][i][j], lumAvg, 11.2) * 255

			// clipping
			if colorR > 255 {
				colorR = 255
			}
			if colorG > 255 {
				colorG = 255
			}
			if colorB > 255 {
				colorB = 255
			}
			outputImage.Set(i, j, color.RGBA{
				R: uint8(colorR),
				G: uint8(colorG),
				B: uint8(colorB),
				A: 255,
			})
		}
	}
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

func SaveAsHdr() {
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
}

func CloseFile(file *os.File) {
	if err := file.Close(); err != nil {
		fmt.Println(err.Error())
	}
}
