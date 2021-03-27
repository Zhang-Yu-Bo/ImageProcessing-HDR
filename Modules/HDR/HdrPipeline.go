package HDR

import (
	"ImageProcessing_HDR/Modules/HDR/Common"
	"ImageProcessing_HDR/Modules/HDR/DebevecMalik"
	"errors"
	_ "image/jpeg"
	_ "image/png"
	"runtime"
	"sync"
)

func RecoverHdrImageWithExposureTime(fileName []string, exposureTime []float64, numOfSampling int) error {
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

	runtime.GOMAXPROCS(runtime.NumCPU())
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		Common.SaveAsHdr()
		wg.Done()
	}()
	Common.GenerateLdrImage()
	// Output Ldr Image
	if err := Common.SaveAsPng(); err != nil {
		return err
	}
	wg.Wait()

	return nil
}
