package main

import (
	"ImageProcessing_HDR/Modules/HDR"
	"ImageProcessing_HDR/Modules/HDR/Common"
	"bufio"
	"fmt"
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
				panic("ShutterTime.txt invalid at line: " + line)
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

	if err := HDR.RecoverHdrImageWithExposureTime(
					listOfFileName,
					listOfExposureTime,
					900,
					Common.LocalToneMapping,
					Common.ReinhardEnhance); err != nil {
		fmt.Println(err.Error())
	}
}
