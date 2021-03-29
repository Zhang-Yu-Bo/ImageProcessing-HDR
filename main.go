package main

import (
	"ImageProcessing_HDR/Modules/HDR"
	"ImageProcessing_HDR/Modules/HDR/Common"
	"bufio"
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
