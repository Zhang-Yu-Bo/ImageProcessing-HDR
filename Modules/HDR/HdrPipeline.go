package HDR

import (
	"ImageProcessing_HDR/Modules/HDR/Common"
	"ImageProcessing_HDR/Modules/HDR/DebevecMalik"
	"errors"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"runtime"
	"sync"
)

func RecoverHdrImageWithExposureTime(fileName []string, exposureTime []float64, numOfSampling int, tmpAction Common.TmpAction, tmpType Common.TmpType) error {
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

	// 因為hdr format的照片通常較大張，因此多執行緒去進行儲存
	runtime.GOMAXPROCS(runtime.NumCPU())
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		Common.SaveAsHdrFormat()
		wg.Done()
	}()

	if tmpAction == Common.LocalToneMapping {
		// magic number
		alpha := 1 / (2 * math.Sqrt2)
		// gaussian scale size ratio, 1.6較接近拉普拉斯
		ratio := 1.6
		// 誤差
		epsilon := 0.05
		// 整體銳度
		phi := 15.0
		// 整體亮度
		a := 0.45
		Common.GenerateLumByRadianceE(a)
		Common.GenerateLocalLumAvgMatrix(alpha, ratio, epsilon, phi, a)
	} else { // tmpAction == Common.GlobalToneMapping
		Common.CalculateGlobalLumAvg()
	}
	Common.GenerateLdrImage(tmpAction, tmpType)
	// Output Ldr Image
	if err := Common.SaveAsPng(); err != nil {
		return err
	}
	wg.Wait()

	return nil
}
