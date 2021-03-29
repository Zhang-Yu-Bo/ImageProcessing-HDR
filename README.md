# ImageProcessing-HDR
* `Golang version 1.16.0`

## How to build
1. Install `gonum`:
```commandline
go get -u gonum.org/v1/gonum
```
2. Follow the step to install `OpenCV` in this link:
   * Windows: https://gocv.io/getting-started/windows/
   * MacOS: https://gocv.io/getting-started/macos/
   * Linux: https://gocv.io/getting-started/linux/
    
3. Build:
```commandline
go build main.go
```

## How to use
Use the terminal to run the HDR program which you just build. \
And there are some arguments you should/could use.
* Require:
   - `-path [FolderPath]` FolderPath is the dir path in which you store the images that have different exposure times.
   - `-match [FileNamePattern]` FileNamePattern is the name pattern of your images.
* Optional:
   - `-samples [NumOfSamples]` Use how many pixels to recover the HDR curve. NumOfSamples is an integer, and the default is 900
   - `-alpha [Alpha]` The initial value of the scale of the Gaussian filter. Alpha is a float, and the default is 1 / (2 * sqrt(2))
   - `-ratio [Ratio]` The ratio between the big and small Gaussian filters. Ratio is a float, and the default is 1.6
   - `-epsilon [Epsilon]` The threshold of the local tone mapping. Epsilon is a float, and the default is 0.05
   - `-phi [Phi]` The value to control the sharpness of the LDR image. Phi is a float, and the default is 15.0
   - `-a [A]` The value to control the brightness of the LDR image. A is a float, and the default is 0.45
   - `-tmoAction [Action]` Use `local / global` tone mapping. Action is a string, and the default is "global". There are 2 actions of tone mapping methods: "local" and "global". Note that "local" is a time-consuming method, you might wait more than 5 minutes. 
   - `-tmoType [Type]` Use `reinhard / ce / uncharted2 / reinhard_enhance / aces` tone mapping. Type is a string, and the default is "aces". There are 5 types of tone mapping methods: "reinhard", "ce", "uncharted2", "reinhard_enhance" and "aces". Note that "reinhard_enhance" only use in the "local" action.
* Example:
  ```commandline
  main.exe -path ./Images/Memorial -match *.png
  main.exe -path ./Images/Exposures -match img??.jpg
  main.exe -path ./Images/Memorial -match *.png -tmoAction local -a 0.45 -tmoType reinhard_enhance -samples 900 -alpha 0.35 -ratio 1.6 -epsilon 0.05 -phi 15.0 -a 0.45
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
