package main

import (
	"math"
	"math/rand"
	"os"
	"time"
)

type tokenAndType struct {
	token     string
	tokenType string
}

var func2 = map[string]struct {
	name func(float64, float64) float64
	prec int // precedence value (higher is more priority)
}{
	"+":     {add, 2},
	"-":     {sub, 2},
	"*":     {mult, 3},
	"/":     {div, 3},
	"^":     {pow, 5},
	"parll": {parll, 5},
}

var func1 = map[string]func(float64) float64{
	"abs":   math.Abs,
	"asin":  math.Asin,
	"asinh": math.Asinh,
	"acos":  math.Acos,
	"acosh": math.Acosh,
	"atan":  math.Atan,
	"atand": atand,
	"atanh": math.Atanh,
	"ceil":  math.Ceil,
	"cos":   math.Cos,
	"cosh":  math.Cosh,
	"exp":   math.Exp,
	"floor": math.Floor,
	"log":   math.Log,
	"log10": math.Log10,
	"round": math.Round,
	"sin":   math.Sin,
	"sinh":  math.Sinh,
	"sqrt":  math.Sqrt,
	"tan":   math.Tan,
	"tanh":  math.Tanh,
	"neg":   neg,
	"pos":   pos,
	"dB":    dB,
	"dBV":   dBV,
}

type varSingle struct { // a structure for each variable hold info below
	latex string  // the latex equivalent of the variable
	units string  // Just the units part (without the prefix)
	value float64 //The value in float64 format (should equal value)
}

// fileInfo is structure that contains file info (example: path1/path2/filename.ext)
type fileInfo struct {
	path string // the relative Directory location (path/path2)
	name string // the filename without extension (filename)
	ext  string // the filename extension (ext)
	full string // full path/path2/filename.ext
}

// input file name is a command line arg (no flag) (path/infile.prb or path/infile.asc)
// output file name is -export=path/outfile.tex or -export=path/outfile.svg
// infile.prb should have outfile.tex while infile.asc should have outfile.svg
// (output file extension needs to be appropriate for input file extension)
// in both cases, error log is at beginning of output file as commented lines
// also software version is output at beginning of output file

func main() {
	var inFileStr, logOut, header, symPath, fontType string
	var sigDigits, txtMode, dotsMode, randomStr string
	var inFile, outFile fileInfo
	var version string
	var ltSpiceInput string

	rand.Seed(time.Now().UnixNano()) // needed so a new seed occurs every time the program is run
	//currentTime := time.Now()
	//	todayDate = currentTime.Format("2006-01-02")
	version = "0.7.5" + " (" + "2020-08-05" + ")"

	inFile, outFile, symPath, randomStr, sigDigits, txtMode, dotsMode, fontType, logOut = commandFlags(version) // outFile depends on inFile file extension
	fileWriteString("", outFile.full)
	if logOut != "" {
		logOut = logOutWrite(logOut, -1, outFile)
		os.Exit(1)
	}
	header = "Created with icemaker: version = " + version
	switch outFile.ext {
	case ".tex":
		_ = logOutWrite(header, -1, outFile)
		inFileStr, logOut = fileReadString(inFile.full)
		if logOut != "" {
			logOut = logOutWrite(logOut, -1, outFile)
		}
		makeTex(inFileStr, sigDigits, randomStr, inFile, outFile)
	case ".svg":
		header0 := `<?xml version="1.0" encoding="UTF-8" standalone="no"?>`
		fileAppendString(header0, outFile.full)
		_ = logOutWrite(header, -1, outFile)
		ltSpiceInput, logOut = getInputFiles(inFile.full)
		if logOut != "" {
			logOut = logOutWrite(logOut, -1, outFile)
		}
		switch txtMode {
		case "symbol":
			_ = ltSymbol2svg(ltSpiceInput, inFile.name, outFile, true)
		default:
			ltSpice2svg(ltSpiceInput, symPath, txtMode, dotsMode, fontType, outFile)
		}
	default:

	}

}

// *******************************************************************************************
// *******************************************************************************************
