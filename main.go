package main

import (
	"ImageProcessing_HDR/Modules/HDR"
	"fmt"
)

func main() {
	listOfFileName := []string{
		"61.png", "62.png", "63.png", "64.png",
		"65.png", "66.png", "67.png", "68.png",
		"69.png", "70.png", "71.png", "72.png",
		"73.png", "74.png", "75.png", "76.png",
	}
	listOfExposureTime := []float64{
		32, 16, 8, 4,
		2, 1, 0.5, 0.25,
		0.125, 0.0625, 0.03125, 0.015625,
		0.0078125, 0.00390625, 0.001953125, 0.0009765625,
	}
	if err := HDR.RecoverHdrImageWithExposureTime(listOfFileName, listOfExposureTime); err != nil {
		fmt.Println(err.Error())
	}
}
