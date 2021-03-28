package main

import (
	"ImageProcessing_HDR/Modules/HDR"
	"ImageProcessing_HDR/Modules/HDR/Common"
	"fmt"
)

func main() {
	imgPath := "./Images/Memorial/"
	listOfFileName := []string{
		"61.png", "62.png", "63.png", "64.png",
		"65.png", "66.png", "67.png", "68.png",
		"69.png", "70.png", "71.png", "72.png",
		"73.png", "74.png", "75.png", "76.png",
	}
	//imgPath := "./Images/NMMST/"
	//listOfFileName := []string{
	//	"1.jpg", "2.jpg", "3.jpg", "4.jpg",
	//	"5.jpg", "6.jpg", "7.jpg", "8.jpg",
	//	"9.jpg", "10.jpg", "11.jpg",
	//}
	for i := 0; i < len(listOfFileName); i++ {
		listOfFileName[i] = imgPath + listOfFileName[i]
	}
	listOfExposureTime := []float64{
		32, 16, 8, 4,
		2, 1, 0.5, 0.25,
		0.125, 0.0625, 0.03125, 0.015625,
		0.0078125, 0.00390625, 0.001953125, 0.0009765625,
	}
	//listOfExposureTime := []float64{
	//	16, 8, 4, 2,
	//	1, 0.5, 0.25, 0.125,
	//	0.0625, 0.03125, 0.015625,
	//}
	//for i := 0; i < len(listOfExposureTime); i++ {
	//	listOfExposureTime[i] /= 8
	//}

	if err := HDR.RecoverHdrImageWithExposureTime(
					listOfFileName,
					listOfExposureTime,
					900,
					Common.LocalToneMapping,
					Common.ReinhardEnhance); err != nil {
		fmt.Println(err.Error())
	}
}
