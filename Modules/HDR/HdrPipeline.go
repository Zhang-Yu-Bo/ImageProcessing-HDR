package HDR

import (
	"ImageProcessing_HDR/Modules/HDR/Common"
	"ImageProcessing_HDR/Modules/HDR/DebevecMalik"
	"ImageProcessing_HDR/Modules/ToneMapping"
	"errors"
	_ "image/jpeg"
	_ "image/png"
)

type hdrState int

const notInit 		hdrState = 0
const recoverState 	hdrState = 1
const tmoState 		hdrState = 2

var NowState = notInit

func RecoverHdrByDebevecMalik(fileName []string, exposureTime []float64, numOfSampling int) error {
	Common.NumOfSamplePixels = numOfSampling
	Common.NumOfImages = len(fileName)
	if Common.NumOfImages != len(exposureTime) {
		return errors.New("num of file is not equal to num of exposure time")
	}
	if err := Common.LoadImageFiles(fileName, exposureTime); err != nil {
		return err
	}

	// HDR pipeline: Sampling -> Fill Matrix -> Calculate g(Zij) -> Calculate RadianceE
	DebevecMalik.PixelSampling(Common.NumOfSamplePixels)
	if err := DebevecMalik.GenerateFunctionGz(); err != nil {
		return err
	}
	DebevecMalik.CalculateRadianceE()
	NowState = recoverState
	return nil
}

/**
 alpha := 1 / (2 * math.Sqrt2)	// magic number
 ratio := 1.6  					// gaussian scale size ratio, 1.6較接近拉普拉斯
 epsilon := 0.05				// 誤差
 phi := 15.0 					// 整體銳度
 a := 0.45						// 整體亮度
 */
func ToneMappingOperate(tmoAction Common.TmoAction, tmoType Common.TmoType, alpha, ratio, epsilon, phi, a float64) error {
	if NowState == notInit {
		return errors.New("There is no hdr data. ")
	}
	if tmoAction == Common.LocalToneMapping {
		ToneMapping.GenerateLumByRadianceE(a)
		ToneMapping.GenerateLocalLumAvgMatrix(alpha, ratio, epsilon, phi, a)
	} else { // tmpAction == Common.GlobalToneMapping
		ToneMapping.CalculateGlobalLumAvg()
	}
	ToneMapping.GenerateLdrImage(tmoAction, tmoType)
	NowState = tmoState
	return nil
}

// Output Ldr Image
func ExportLdrImage() error {
	if NowState == notInit || NowState == recoverState {
		return errors.New("There is no ldr data. ")
	}
	if err := Common.SaveAsPng(); err != nil {
		return err
	}
	return nil
}

func ExportHdrImage() error {
	if NowState == notInit {
		return errors.New("There is no hdr data. ")
	}
	Common.SaveAsHdrFormat()
	return nil
}
