package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"

	"golang.org/x/text/encoding/unicode"
)

// get flag info and argument
// NOTE: arg MUST occur AFTER flags when calling program
// icemaker -export=tmp/outfilename.tex -sigDigits=3 infilename.prb
func commandFlags(version string) (inFile fileInfo, outFile fileInfo, symPath, randomStr, sigDigits, txtMode, dotsMode, logOut string) {
	var inFileStr string

	outFilePtr := flag.String("export", "", "outFile - REQUIRED FLAG\nFile extension should be .tex or .svg")
	symPathPtr := flag.String("symPath", "", "symPath - path for where to look for LTSpice symbols")
	sigDigitsPtr := flag.String("sigDigits", "4", "number of significant digits for output\n")
	// determines sig digits number for prob2tex
	randomPtr := flag.String("random", "false", "If \"false\", first element in each parameter set is used\nIf \"true\", parameters are randomly chosen from parameter sets\nIf a positive integer, that integer is the seed for random generator\n")
	// determines whether parameters are default or random chosen from a set

	txtModePtr := flag.String("text", "noChange", "If \"noChange\", text is unchanged\nIf \"latex\", instantiation names will be put in latex equations\nIf \"subscript\", instantiation names will have subscripts\n   and _{x1} will put x1 as a subscript\nIf \"symbol\", an svg output will be created for the symbol input file\n")
	// text flag sets test output as well as changing to symbol svg generation
	// - if left blank or not used, text is unchanged
	// - if "latex", instantiation names will be put in latex equations (nothing else is changed)
	// - if "subscript", instantiation names will have subscripts and if _{x1} is used, it will also be put as a subscript
	//   labels do not accept "{" so x_dd will be equivalent to x_{dd}
	// - if "symbol", the input file should be an ltspice symbol and the output will be the svg output for that symbol to be
	//   included in the symDefn file
	dotsModePtr := flag.String("Tdots", "true", "true or false\nPlace dots on wired T connections\n")
	versionPtr := flag.Bool("version", false, "Print out version")

	flag.Parse()
	if *versionPtr == true {
		fmt.Println("Icemaker ", version)
	}
	exitCode := 0
	inFileStr = flag.Arg(0)
	if inFileStr == "" {
		exitCode = 1
		fmt.Println("No input file name given\nRun with -help to see inputs required")
		os.Exit(exitCode)
	}
	if *outFilePtr == "" {
		exitCode = 1
		fmt.Println("No outFile given\nRun with -help to see inputs required")
		os.Exit(exitCode)
	}

	inFile = getFileInfo(inFileStr)
	outFile = getFileInfo(*outFilePtr)
	symPath = *symPathPtr
	if symPath == "" {
		symPath = "no symPath given"
	}
	randomStr = *randomPtr
	_, logOut = checkRandom(randomStr, logOut)
	sigDigits, logOut = checkSigDigits(*sigDigitsPtr, logOut)
	sigDigits = strIncrement(sigDigits, -1) // needed so that TOTAL significant digits is sigDigits
	if outFile.ext == "" {
		outFile.ext = ".log"
		outFile.full = filepath.Join(outFile.path, outFile.name+outFile.ext)
		logOut = logOut + "Output file needs a file extension of either .tex or .svg\n"
	}
	switch outFile.ext {
	case ".tex":
		if inFile.ext != ".prb" {
			logOut = logOut + "Input should be .prb file since output is .tex\n"
		}
	case ".svg":
		if inFile.ext != ".asc" && inFile.ext != ".asy" {
			logOut = logOut + "Input should be .asc or .asy file since output is .svg\n"
		}
	default:
		logOut = logOut + "Output file needs a file extension of either .svg or .tex\n"
		outFile.ext = ".log"
		outFile.full = filepath.Join(outFile.path, outFile.name+outFile.ext)
	}

	txtMode = *txtModePtr
	dotsMode = *dotsModePtr
	return
}

func checkRandom(randomStr, logOut string) (int, string) {
	var random int
	var err error
	switch randomStr {
	case "false":
		random = 0
	case "true":
		random = -1
	default: //check that string is a positive integer
		random, err = strconv.Atoi(randomStr)
		if err != nil {
			logOut = logOut + "random should be either \"false\", \"true\", or a positive integer\n"
		} else {
			if random < 1 {
				logOut = logOut + "random should be a positive integer\n"
			}
		}
	}
	return random, logOut
}

func checkSigDigits(sigDigits, logOut string) (string, string) {
	i, err := strconv.Atoi(sigDigits)
	if err != nil {
		logOut = logOut + "sigDigits should be a positive integer\n"
		sigDigits = "4"
	} else {
		if i < 1 {
			logOut = logOut + "sigDigits should be a positive integer\n"
			sigDigits = "4"
		}
	}
	return sigDigits, logOut
}

func getFileInfo(inString string) (file fileInfo) {
	var base string
	var re0 = regexp.MustCompile(`(?m)^(?P<res1>\w*)`)
	//var result []string

	file.path = filepath.Dir(inString)
	file.ext = filepath.Ext(inString)
	file.full = inString

	base = filepath.Base(inString)
	if re0.MatchString(base) {
		file.name = re0.FindStringSubmatch(base)[1]
	}
	return
}

func fileWriteString(inString, fileNameandPath string) {
	// write inString to file "fileNameandPath" (does NOT append, it overwrites)
	outbytes := []byte(inString)
	err := ioutil.WriteFile(fileNameandPath, outbytes, 0644)
	if err != nil { // if error, then create an ERROR.log file and write to it the error
		outbytes := []byte("Cannot write " + fileNameandPath + "\n")
		_ = ioutil.WriteFile("ERROR.log", outbytes, 0644) // ERROR log file created
		os.Exit(1)
	}
}

func fileAppendString(inString, fileNameandPath string) {
	// append inString to file "fileNameandPath" (will create it if it does not exist)
	f, err := os.OpenFile(fileNameandPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write([]byte(inString + "\n")); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func fileReadString(fileNameandPath string) (string, string) {
	var fileString, logOut string
	inbytes, err := ioutil.ReadFile(fileNameandPath) //
	if err != nil {
		//	fmt.Print(err)
		logOut = fmt.Sprint(err)
	}
	fileString = string(inbytes)
	return fileString, logOut
}

func checkIfNewer(newFile, oldFile string) bool {
	var newer bool
	var commandTest string
	// var svgFile, pdfTexFile string
	// svgFile = inFile.path + inFile.name + inFile.extension
	// pdfTexFile = outPath + inFile.name + ".pdf_tex"
	commandTest = "test " + newFile + " -nt " + oldFile + " ; echo $?"
	out, errout, err := shellout(commandTest)
	if err != nil {
		log.Printf("error: %v\n", err)
	}
	_ = errout
	// fmt.Println("--- stdout ---")
	// fmt.Println(out)
	// fmt.Println("--- stderr ---")
	// fmt.Println(errout)
	var re = regexp.MustCompile(`(?m)0`)
	if re.MatchString(out) {
		newer = true
	} else {
		newer = false
	}
	return newer
}

func shellout(command string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// Checks if file is utf16 encoded and if so, it converts it to utf8 for better regex matching
func convertIfUtf16(inString string) (string, bool) {
	// requires import "golang.org/x/text/encoding/unicode"
	var inBytes []byte
	var codeUtf16 bool
	inBytes = []byte(inString)
	if len(inBytes) > 7 {
		if inBytes[1] == 0 && inBytes[3] == 0 && inBytes[5] == 0 && inBytes[7] == 0 { // VERY likely utf16 encoded so need to change to utf8
			codeUtf16 = true
			decoder := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()
			inString, _ = decoder.String(inString)
		}
	}
	return inString, codeUtf16
}

// makes string into a comment - intended for log comments
// type of comment depends on file extension (either tex or svg)
func logComment(logOut string, outExt string) string {
	switch outExt {
	case ".tex":
		logOut = "% " + logOut
	case ".svg":
		logOut = "<!-- " + logOut + " -->"
	default:
	}
	return logOut
}
