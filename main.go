package main

import (
	"ImageProcessing_HDR/Modules/HDR"
	"ImageProcessing_HDR/Modules/HDR/Common"
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func CheckArgusIsEnough(i *int, length int) {
	if *i++; *i >= length {
		panic(os.Args[*i] + ": argument not enough.")
	}
}

func main() {
	argsWithoutProg := os.Args[1:]
	argsLen := len(argsWithoutProg)
	var mFilepath string
	var mMatchPattern string
	var listOfFileName []string
	numOfSampling := 900
	alpha := 1 / (2 * math.Sqrt2)
	ratio := 1.6
	epsilon := 0.05
	phi := 15.0
	a := 0.45
	tmoAction := Common.GlobalToneMapping
	tmoType := Common.Aces

	// Read arguments
	// ex1: ./main.exe -path ./Images/Memorial -match *.png
	// ex2: ./main.exe -path ./Images/Exposures -match img??.jpg
	if argsLen <= 0 {
		panic("There is no arguments.")
	}
	for i :=0; i < argsLen;i++ {
		if argsWithoutProg[i] == "-path" {
			CheckArgusIsEnough(&i, argsLen)

			mFilepath = argsWithoutProg[i]
			if _, err := os.Stat(mFilepath); /*os.IsNotExist(err)*/ err != nil {
				panic(err)
			}
			if mFilepath[len(mFilepath) - 1] != '/' {
				mFilepath += "/"
			}
		} else if argsWithoutProg[i] == "-match" {
			CheckArgusIsEnough(&i, argsLen)

			mMatchPattern = argsWithoutProg[i]
			if mMatch, err := filepath.Glob(mFilepath + mMatchPattern); err != nil {
				panic(err)
			} else {
				for _, k := range mMatch {
					listOfFileName = append(listOfFileName, k)
				}
			}
		} else if argsWithoutProg[i] == "-samples" {
			CheckArgusIsEnough(&i, argsLen)

			var err error
			numOfSampling, err = strconv.Atoi(argsWithoutProg[i])
			if err != nil {
				fmt.Println(err)
				numOfSampling = 900
			}
		} else if argsWithoutProg[i] == "-alpha" {
			CheckArgusIsEnough(&i, argsLen)

			var err error
			alpha, err = strconv.ParseFloat(argsWithoutProg[i], 64)
			if err != nil {
				fmt.Println(err)
				alpha = 1 / (2 * math.Sqrt2)
			}
		} else if argsWithoutProg[i] == "-ratio" {
			CheckArgusIsEnough(&i, argsLen)

			var err error
			ratio, err = strconv.ParseFloat(argsWithoutProg[i], 64)
			if err != nil {
				fmt.Println(err)
				ratio = 1.6
			}
		} else if argsWithoutProg[i] == "-epsilon" {
			CheckArgusIsEnough(&i, argsLen)

			var err error
			epsilon, err = strconv.ParseFloat(argsWithoutProg[i], 64)
			if err != nil {
				fmt.Println(err)
				epsilon = 0.05
			}
		} else if argsWithoutProg[i] == "-phi" {
			CheckArgusIsEnough(&i, argsLen)

			var err error
			phi, err = strconv.ParseFloat(argsWithoutProg[i], 64)
			if err != nil {
				fmt.Println(err)
				phi = 15.0
			}
		} else if argsWithoutProg[i] == "-a" {
			CheckArgusIsEnough(&i, argsLen)

			var err error
			a, err = strconv.ParseFloat(argsWithoutProg[i], 64)
			if err != nil {
				fmt.Println(err)
				a = 0.45
			}
		} else if argsWithoutProg[i] == "-tmoAction" {
			CheckArgusIsEnough(&i, argsLen)

			if argsWithoutProg[i] == "local" {
				tmoAction = Common.LocalToneMapping
			} else {
				tmoAction = Common.GlobalToneMapping
			}
		} else if argsWithoutProg[i] == "-tmoType" {
			CheckArgusIsEnough(&i, argsLen)

			if argsWithoutProg[i] == "reinhard" {
				tmoType = Common.Reinhard
			} else if argsWithoutProg[i] == "ce"{
				tmoType = Common.CE
			} else if argsWithoutProg[i] == "uncharted2"{
				tmoType = Common.Uncharted2
			} else if argsWithoutProg[i] == "reinhard_enhance"{
				tmoType = Common.ReinhardEnhance
			} else {
				tmoType = Common.Aces
			}
		}
	}

	// Read ShutterTime.txt
	shutterInfo := make(map[string]float64)
	var listOfExposureTime []float64
	if shutterFile, err := os.Open(mFilepath+"ShutterTime.txt"); err != nil {
		panic(err)
	} else {
		defer Common.CloseFile(shutterFile)

		scanner := bufio.NewScanner(shutterFile)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			line := scanner.Text()
			keyValue := strings.Split(line, " ")
			if len(keyValue) != 2 {
				panic("ShutterTime.txt invalid format at line: " + line)
			}
			if shutterInfo[keyValue[0]], err = strconv.ParseFloat(keyValue[1], 64); err != nil {
				panic(err)
			}
		}
	}
	if len(listOfFileName) != len(shutterInfo) {
		panic("The num of the images is not equal to the num of the shutter time.")
	}
	for _, value := range listOfFileName {
		_, fileName := filepath.Split(value)
		if shutterTime, exist := shutterInfo[fileName]; !exist {
			panic("Image [" + fileName + "] is not exist in the ShutterTime.txt")
		} else {
			listOfExposureTime = append(listOfExposureTime, shutterTime)
		}
	}

	// HDR Recover
	if err := HDR.RecoverHdrByDebevecMalik(listOfFileName, listOfExposureTime, numOfSampling); err != nil {
		panic(err)
	}
	// Export Hdr Image
	if err := HDR.ExportHdrImage(); err != nil {
		panic(err)
	}
	// Tone Mapping
	if err := HDR.ToneMappingOperate(tmoAction, tmoType, alpha, ratio, epsilon, phi, a); err != nil {
		panic(err)
	}
	// Export Ldr Image
	if err:= HDR.ExportLdrImage(); err != nil {
		panic(err)
	}

}
