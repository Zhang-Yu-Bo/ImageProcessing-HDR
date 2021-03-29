# ImageProcessing-HDR
* `Golang version 1.16.0`

## How to build
1. Install gonum:
```commandline
go get -u gonum.org/v1/gonum
```
2. Install opencv go version: \
   Follow the step to install opencv in this link
   * [Windows](https://gocv.io/getting-started/windows/)
   * [MacOS](https://gocv.io/getting-started/macos/)
   * [Linux](https://gocv.io/getting-started/linux/)
    
3. Build:
```commandline
go build main.go
```

## How to run
* There are two arguments that must be filled in: -path, -match
```commandline
main.exe -path ./Images/Memorial -match *.png
```

## Reference
* [HDR Tools](https://ttic.uchicago.edu/~cotter/projects/hdr_tools/)
* [Tone Mapping](https://www.phototalks.idv.tw/academic/?p=861)
* [HDR introduce](https://www.phototalks.idv.tw/academic/?p=636)
* [HDR Paper 1997](http://www.pauldebevec.com/Research/HDR/)
* [基於物理的渲染 - HDR Tone Mapping](https://zhuanlan.zhihu.com/p/26254959)
* [Tone Mapping進化論](https://zhuanlan.zhihu.com/p/21983679)
* [Vedant2311/Tone-Mapping-Library](https://github.com/Vedant2311/Tone-Mapping-Library)

## Keyword
* `Golang`
* `Image Processing`
* `HDR`
* `Tone Mapping`
