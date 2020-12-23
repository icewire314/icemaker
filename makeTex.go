package main

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func makeTex(problemInput, sigDigits, randomStr string, inFile, outFile fileInfo) {
	var inLines []string
	var logOut, comment string
	var texOut string
	var linesToRemove []int
	var reDeletethis = regexp.MustCompile(`(?m)\*\*deletethis\*\*`)
	var reNotBlankLine = regexp.MustCompile(`(?m)\S`)

	// using map here as I want to be able to iterate over key names as well
	// as looking at value for each key
	// these are configuration parameters
	var configParam = map[string]string{
		"paramRandom":    randomStr, // can be -1, 0, or any positive integer
		"paramSigDigits": sigDigits, // number of significant digits to print
		"paramVariation": "20:5",    // percentage varation : number of choices
		"paramFormat":    "eng",     // can be eng, sci or decimal
	}

	varAll := make(map[string]varSingle) // IMPORTANT to use this type of map assignment - tried another and it worked for a while
	// til hash table memory changed and then memory errors on run that could not be traced by debugger
	// the downside of this approach is when changing a key value struct element, need to copy struct first then change
	// struct element then copy it back into hash table.  Can not change just a single struct element without this copy first

	// WITH VARIABLE MAP, keywords variation/random/sigDigits are used and defaults set below
	// stored in MAP so user can change them using \runParam{random=true}... during run of program

	// var key string
	// for key = range configParam {
	// 	fmt.Println(configParam[key])
	// }

	inLines = strings.Split(problemInput, "\n")
	for i := range inLines {
		inLines[i], comment = deCommentLatex(inLines[i])
		logOut = syntaxWarning(inLines[i])
		if logOut != "" {
			logOut = logOutWrite(logOut, i, outFile)
		}
		inLines[i], logOut = valRunReplace(inLines[i], varAll, configParam, false)
		if logOut != "" {
			logOut = logOutWrite(logOut, i, outFile)
		}
		inLines[i] = fixParll(inLines[i])
		inLines[i] = inLines[i] + comment // add back comment that was removed above
		inLines[i] = function2Latex(inLines[i])
		if reDeletethis.MatchString(inLines[i]) {
			inLines[i] = reDeletethis.ReplaceAllString(inLines[i], "")
			if !reNotBlankLine.MatchString(inLines[i]) { // if a blank line then add line number to linesToRemove list
				linesToRemove = append(linesToRemove, i) // do it later so that line numbers still correct if an error is reported after a line removal
			}
		}
		logOut = checkLTSpice(inLines[i], inFile, outFile, sigDigits, varAll, configParam)
		if logOut != "" {
			logOut = logOutWrite(logOut, i, outFile)
		}
	}
	// remove lines that are slated for removal
	k := 0
	for i := range linesToRemove {
		inLines = remove(inLines, linesToRemove[i]-k) // need to subtract k as that those number of lines have already been removed
		k++
	}
	texOut = strings.Join(inLines, "\n")
	fileAppendString(texOut, outFile.full)
	return
}

func logOutWrite(logOut string, lineNum int, outFile fileInfo) string {
	if lineNum != -1 { // dont include line number if lineNum = -1
		logOut = logOut + " - Line number: " + strconv.Itoa(lineNum+1)
	}
	logOut = logComment(logOut, outFile.ext)
	fileAppendString(logOut, outFile.full)
	fmt.Println(logOut)
	logOut = ""
	return logOut
}

func valRunReplace(inString string, varAll map[string]varSingle, configParam map[string]string, ltSpice bool) (string, string) {
	// ltSpice is a bool that if true implies we are replacing things in a .asc file (instead of a .prb file)
	var result []string
	var head, tail, replace, logOut, newLog, key string
	var valCmd, valCmdType, runCmd, runCmdType string
	var assignVar string
	var answer float64
	var ok bool
	var reFirstvalRunCmd = regexp.MustCompile(`(?mU)^(?P<res1>.*)\\(?P<res2>run.*|val.*)(?P<res3>{.*)$`)
	var reFirstvalCmd = regexp.MustCompile(`(?mU)^(?P<res1>.*)\\(?P<res2>val.*)(?P<res3>{.*)$`)
	var reFirstRunCmd = regexp.MustCompile(`(?mU)^(?P<res1>.*)\\(?P<res2>run.*)(?P<res3>{.*)$`)
	var reFixSpice = regexp.MustCompile(`(?mU)\\(?P<res1>\\val.*{)`)
	if ltSpice { // fix ltSpice file so all \\val.*{ are changed to \val.*{
		if reFixSpice.MatchString(inString) {
			inString = reFixSpice.ReplaceAllString(inString, "$res1")
		}
	}
	for reFirstvalRunCmd.MatchString(inString) { // chec for val or run command
		if reFirstvalCmd.MatchString(inString) { // check for a val command
			result = reFirstvalCmd.FindStringSubmatch(inString) // found a val command
			head = result[1]
			valCmdType = result[2]
			valCmd, tail = matchBrackets(result[3], "{")
			replace = "" // so the old replace is not used

			// need an error check here to ensure valCmd is only a word (no equation). Use \run{} if an equation with no assigned value needed

			switch valCmdType {
			case "val": // print out result in engineering notation (ex: 3.14159 or 23e3 or 240e-6)
				for key = range configParam {
					if valCmd == key {
						replace = configParam[key] // printing out a configParam
					}
				}
				if replace == "" { // no configParam found then do below
					_, _, answer, newLog = runCode(valCmd, varAll)
					if newLog == "" {
						replace = float2Str(answer, configParam)
					} else {
						replace = newLog
					}
				}
			case "val=": // print out var = result (with SI units).  (ex: V_1 = 3V or v_{tx} = 23mV or D = 10km)
				_, ok = varAll[valCmd]
				if ok {
					replace = "\\mbox{$" + varAll[valCmd].latex + " = " + valueInSI(valCmd, varAll, configParam) + "$}"
				} else {
					replace = "\\mbox{$" + valCmd + " \\text{ NOT DEFINED}$}"
				}
			case "valU": // print out result (with SI units). (ex: 3V or 23mV or 10km)
				_, ok = varAll[valCmd]
				if ok {
					replace = "\\mbox{$" + valueInSI(valCmd, varAll, configParam) + "$}"
				} else {
					replace = "\\mbox{$" + valCmd + " \\text{ NOT DEFINED}$}"
				}
			default:
				// if here, then \val**something else** found so an error message
				logOut = "\\" + valCmdType + " *** NOT A VALID COMMAND\n"
				inString = logOut
				return inString, logOut
			}
		}
		if !ltSpice { // also run these commands below if ltSpice is false
			if reFirstRunCmd.MatchString(inString) {
				result = reFirstRunCmd.FindStringSubmatch(inString)
				head = result[1]
				runCmdType = result[2]
				runCmd, tail = matchBrackets(result[3], "{")
				replace = "" // so the old replace is not used
				switch runCmdType {
				case "runParam": // Used for setting parameters and config parameters
					replace = "**deletethis**"
					newLog = runParamFunc(runCmd, varAll, configParam)
				case "runSilent": // run statement but do not print anything
					replace = "**deletethis**"
					_, _, _, newLog = runCode(runCmd, varAll)
				case "run": // run statement and print statement (ex: v_2 = 3*V_t)
					assignVar, runCmd, answer, newLog = runCode(runCmd, varAll)
					if assignVar == "" {
						replace = float2Str(answer, configParam) // not an assignment statment so just return  answer
					} else {
						replace = "\\mbox{$" + latexStatement(runCmd, varAll) + "$}"
					}
				case "run=": // run statement and print out statement = result (with units) (ex: v_2 = 3*V_t = 75mV)
					assignVar, runCmd, answer, newLog = runCode(runCmd, varAll)
					if assignVar == "" {
						replace = "error: not an assignment statement"
					} else {
						replace = "\\mbox{$" + latexStatement(runCmd, varAll) + " = " + valueInSI(assignVar, varAll, configParam) + "$}"
					}
				case "run()": // same as run but include = bracket values in statement (ex" v_2 = 3*V_t = 3*(25e-3))
					_, runCmd, _, newLog = runCode(runCmd, varAll)
					replace = "\\mbox{$" + latexStatement(runCmd, varAll) + bracketed(runCmd, varAll, configParam) + "$}"
				case "run()=": // same as run() but include result (ex: v_2 = 3*V_t = 3*(25e-3)=75mV)
					assignVar, runCmd, _, newLog = runCode(runCmd, varAll)
					replace = "\\mbox{$" + latexStatement(runCmd, varAll) + bracketed(runCmd, varAll, configParam) + " = " + valueInSI(assignVar, varAll, configParam) + "$}"
				default:
					// if here, then error as \run**something else** is here
					logOut = "\\" + runCmdType + " *** NOT A VALID COMMAND\n"
					inString = logOut
					return inString, logOut
				}
			}
		}
		inString = head + replace + tail
		logOut = logOut + newLog
	}
	return inString, logOut
}

func checkLTSpice(inString string, inFile, outFile fileInfo, sigDigits string, varAll map[string]varSingle, configParam map[string]string) string {
	var spiceFilename, spiceFile, logOut string
	var inLines []string
	var reLTSpice = regexp.MustCompile(`(?mU)\\incProbLTspice.*{\s*(?P<res1>\S*)\s*}`)
	if reLTSpice.MatchString(inString) {
		spiceFilename = reLTSpice.FindStringSubmatch(inString)[1]
		spiceFile, logOut = fileReadString(filepath.Join(inFile.path, spiceFilename+".asc"))
		if logOut != "" {
			return logOut
		}
		spiceFile, _ = convertIfUtf16(spiceFile)
		inLines = strings.Split(spiceFile, "\n")
		for i := range inLines {
			inLines[i], logOut = valRunReplace(inLines[i], varAll, configParam, true)
			if logOut != "" {
				logOut = logOut + " - error in " + spiceFilename + ".asc"
				return logOut
			}
		}
		spiceFile = strings.Join(inLines, "\n")
		fileWriteString(spiceFile, filepath.Join(outFile.path, spiceFilename+"_update.asc"))
	}
	return logOut
}

func syntaxWarning(statement string) (logOut string) {
	var reDollar = regexp.MustCompile(`(?m)\$`)
	logOut = bracketCheck(statement, "{")
	logOut = logOut + bracketCheck(statement, "(")
	logOut = logOut + bracketCheck(statement, "[")
	matches := reDollar.FindAllStringIndex(statement, -1)
	// need to count $ and see that they are even (backetCheck will not work here)
	if len(matches)%2 != 0 {
		logOut = logOut + "Uneven number of $ so likely unmatched"
	}
	return
}

func bracketCheck(inString string, leftBrac string) (logOut string) {
	var rightBrac string
	var count int
	switch leftBrac {
	case "{":
		rightBrac = "}"
	case "(":
		rightBrac = ")"
	case "[":
		rightBrac = "]"
	case "<":
		rightBrac = ">"
	default:
	}
	count = 0
	for i := range inString {
		if string(inString[i]) == leftBrac {
			count++
		}
		if string(inString[i]) == rightBrac {
			count--
		}
		if count < 0 {
			logOut = "Unmatched brackets: more " + rightBrac + " than " + leftBrac
			return
		}
	}
	if count > 0 {
		logOut = "Unmatched brackets: more " + leftBrac + " than " + rightBrac
	}
	return
}

func runParamFunc(statement string, varAll map[string]varSingle, configParam map[string]string) string {
	var assignVar, rightSide, key, logOut string
	var remainder, units, latex, prefix string
	var value float64
	var num, random int
	var result, values []string
	var reEqual = regexp.MustCompile(`(?m)^\s*(?P<res1>\w+)\s*=\s*(?P<res2>.*)\s*`)
	var reArray = regexp.MustCompile(`(?m)^\s*\[(?P<res1>.*)\]\s*(?P<res2>.*)`)
	var reCommaSep = regexp.MustCompile(` *, *`) // used to create a slice of possible random values input: 2,3, 5,  8 creates slice of 4 elements
	var reOptions = regexp.MustCompile(`(?m)#(?P<res1>.*)$`)
	var reUnits = regexp.MustCompile(`(?m)\\paramUnits(?P<res1>{.*)$`)
	var reLatex = regexp.MustCompile(`(?m)\\paramLatex(?P<res1>{.*)$`)
	if reEqual.MatchString(statement) {
		result = reEqual.FindStringSubmatch(statement)
		assignVar = result[1]
		rightSide = strings.TrimSpace(result[2])
		assignVar, logOut = checkReserved(assignVar, logOut)
		if logOut != "" {
			return logOut
		}
		for key = range configParam {
			if assignVar == key { // it is a configParam runParam statement
				switch assignVar {
				case "paramRandom":
					random, logOut = checkRandom(rightSide, logOut)
					if logOut == "" {
						configParam["paramRandom"] = rightSide
					}
				case "paramSigDigits":
					// put check for paramSigDigits here
					configParam["paramSigDigits"] = rightSide
				case "paramFormat":
					//
					configParam["paramFormat"] = rightSide
				case "paramVariation":
					//
					configParam["paramVariation"] = rightSide
				default:
					logOut = "should never be here 05"
				}
				return logOut
			}
		}
		tmp2, ok := varAll[assignVar]
		if !ok { // if !ok then this is the first time assigning this variable in varAll map
			varAll[assignVar] = varSingle{}
			tmp2 = varAll[assignVar]
			tmp2.latex = latexifyVar(assignVar) // add latex version of assignVar
			tmp2.units = defaultUnits(assignVar)
		}
		switch {
		case reArray.MatchString(rightSide): // it is an array runParam statement
			result = reArray.FindStringSubmatch(rightSide)
			values = reCommaSep.Split(result[1], -1)
			remainder = result[2]
		case reArray.MatchString("not here"): // put step match here
		case reArray.MatchString("not here"): // put variation match here
		default:
			logOut = "not a valid \runParam statement"
			return logOut
		}
		if reOptions.MatchString(remainder) {
			options := reOptions.FindStringSubmatch(remainder)[1] // the stuff after #
			if reUnits.MatchString(options) {
				tmp := reUnits.FindStringSubmatch(options)[1] // just the stuff {.*$
				preUnits, _ := matchBrackets(tmp, "{")        // a string that has prefix and units together
				prefix, units = getPrefixUnits(preUnits)      // separate preUnits into prefix and units
				tmp2.units = units
			}
			if reLatex.MatchString(options) {
				tmp := reLatex.FindStringSubmatch(options)[1]
				latex, _ = matchBrackets(tmp, "{")
				tmp2.latex = latex
			}
		}
		random, logOut = checkRandom(configParam["paramRandom"], logOut)
		switch random {
		case 0: // if random == 0, then num = 0 so first element is chosen
			num = 0
		case -1: // if random == -1, then  num is a random in between 0 and values-1 (based on machine time so pretty much really random)
			num = rand.Intn(len(values))
		default: // if here, random is a seed so use it to get the next random
			random = psuedoRand(random) // update random based on the last random value (treat last one as seed)
			num = randInt(len(values), random)
			configParam["paramRandom"] = strconv.Itoa(random)
		}
		value, _ = strconv.ParseFloat(values[num], 64) // convert string to float64 value
		value = value * prefix2float(prefix)
		tmp2.value = value
		varAll[assignVar] = tmp2
	}
	return logOut
}

func runParamArray(statement string, varAll map[string]varSingle, configParam map[string]string) string {
	var result, values []string
	var assignVar, prefix, units, latex string
	var value float64
	var num, random int
	var logOut, sigDigits string
	var re0 = regexp.MustCompile(`(?m)^\s*(?P<res1>\w+)\s*=\s*\[(?P<res2>.*)\]`)
	var reCommaSep = regexp.MustCompile(` *, *`) // used to create a slice of possible random values input: 2,3, 5,  8 creates slice of 4 elements
	var reOptions = regexp.MustCompile(`(?m)#(?P<res1>.*)$`)
	var reUnits = regexp.MustCompile(`(?m)\\paramUnits(?P<res1>{.*)$`)
	var reLatex = regexp.MustCompile(`(?m)\\paramLatex(?P<res1>{.*)$`)
	var reRandom = regexp.MustCompile(`(?m)paramRandom\s*=\s*(?P<res1>\w+)`)
	var reSigDigits = regexp.MustCompile(`(?m)sigDigits\s*=\s*(?P<res1>\w+)`)
	var reValid = regexp.MustCompile(`(?m)[\w|=|*|/|+|\-|^|(|)|\s|\.|,|#|\\|\[|\]]+`) // valid characters in statement
	// first check if all characters in statement are valid characters
	shouldBeBlank := reValid.ReplaceAllString(statement, "")
	if shouldBeBlank != "" {
		logOut = "character(s) " + shouldBeBlank + " should not be in a \\runParam statement"
		return logOut
	}
	if reRandom.MatchString(statement) { // check if assigning "paramRandom" variable in \runParam
		configParam["paramRandom"] = reRandom.FindStringSubmatch(statement)[1] // update paramRandom
	}
	random, logOut = checkRandom(configParam["paramRandom"], logOut) //check that paramRandom is one of positive integer/true/false
	if reSigDigits.MatchString(statement) {                          // check if assigning "sigDigits" variable in \runParam
		sigDigits = reSigDigits.FindStringSubmatch(statement)[1]
		sigDigits, logOut = checkSigDigits(sigDigits, logOut)
	}
	if re0.MatchString(statement) {
		result = re0.FindStringSubmatch(statement)
		assignVar = result[1]                                // the variable to change/add in varAll map
		values = reCommaSep.Split(result[2], -1)             // slice of the possible values
		assignVar, logOut = checkReserved(assignVar, logOut) // check if assignVar is a reserved variable

		tmp2, ok := varAll[assignVar]
		if !ok { // if !ok then this is the first time assigning this variable in varAll map
			varAll[assignVar] = varSingle{}
			tmp2 = varAll[assignVar]
			tmp2.latex = latexifyVar(assignVar) // add latex version of assignVar
			tmp2.units = defaultUnits(assignVar)
		}
		if reOptions.MatchString(statement) {
			options := reOptions.FindStringSubmatch(statement)[1] // the stuff after #
			if reUnits.MatchString(options) {
				tmp := reUnits.FindStringSubmatch(options)[1] // just the stuff {.*$
				preUnits, _ := matchBrackets(tmp, "{")        // a string that has prefix and units together
				prefix, units = getPrefixUnits(preUnits)      // separate preUnits into prefix and units
				tmp2.units = units
			}
			if reLatex.MatchString(options) {
				tmp := reLatex.FindStringSubmatch(options)[1]
				latex, _ = matchBrackets(tmp, "{")
				tmp2.latex = latex
			}
		}
		switch random {
		case 0: // if random == 0, then num = 0 so first element is chosen
			num = 0
		case -1: // if random == -1, then  num is a random in between 0 and values-1 (based on machine time so pretty much really random)
			num = rand.Intn(len(values))
		default: // if here, random is a seed so use it to get the next random
			random = psuedoRand(random) // update random based on the last random value (treat last one as seed)
			num = randInt(len(values), random)
			configParam["paramRandom"] = strconv.Itoa(random)
		}
		value, _ = strconv.ParseFloat(values[num], 64) // convert string to float64 value
		value = value * prefix2float(prefix)
		tmp2.value = value
		varAll[assignVar] = tmp2
	}
	return logOut
}

func getPrefixUnits(prefixUnits string) (prefix string, units string) {
	var result []string
	var re0 = regexp.MustCompile(`(?m)^(P|T|G|M|k|m|\\mu|n|p|f|a)\s*(?P<res2>\S+)`)
	var re1 = regexp.MustCompile(`(?m)^m\s*\w`)
	if re0.MatchString(prefixUnits) {
		result = re0.FindStringSubmatch(prefixUnits)
		prefix = result[1]
		units = result[2]
		if prefix == "m" {
			if re1.MatchString(prefixUnits) {
				// "m" is a prefix so leave as is
			} else {
				// "m" is for meter so set prefix to blank and all are units
				prefix = ""
				units = prefixUnits
			}
		}
	} else {
		units = prefixUnits
	}
	return
}

// delete element of string slice while maintaing order
func remove(slice []string, s int) []string {
	return append(slice[:s], slice[s+1:]...)
}

// bracketed function used to create () part of \run() where bracketed part are
// all numbers for the variables
func bracketed(statement string, varAll map[string]varSingle, configParam map[string]string) (outString string) {
	var backPart, sub string
	var result []string
	var re0 = regexp.MustCompile(`(?m)=(?P<res1>.*)$`) // get stuff after = to end
	var re1 = regexp.MustCompile(`(?m)(?P<res1>\w+)`)  // find all words
	var re2 = regexp.MustCompile(`(?m)`)               // just a declare as it will change below
	if re0.MatchString(statement) {
		backPart = re0.FindStringSubmatch(statement)[1]
		result = re1.FindAllString(backPart, -1)
		for i := range result {
			_, ok := varAll[result[i]]
			if ok {
				re2 = regexp.MustCompile(`(?m)` + result[i])
				sub = "(" + float2Str(varAll[result[i]].value, configParam) + ")"
				backPart = re2.ReplaceAllString(backPart, sub)
			}
		}
		outString = " = " + backPart
	}
	return
}

func valueInSI(variable string, varAll map[string]varSingle, configParam map[string]string) (outSI string) {
	significand, exponent, prefix := float2Parts(varAll[variable].value, strIncrement(configParam["paramSigDigits"], -1))
	if varAll[variable].units == "" {
		if exponent == "0" {
			outSI = significand
		} else {
			outSI = significand + "e" + exponent
		}
	} else {
		outSI = "\\mbox{$" + significand + " \\units{" + prefix + " " + varAll[variable].units + "}$}"
	}
	return
}

func float2Str(x float64, configParam map[string]string) (outString string) {
	outString = "should not occur 01"
	switch configParam["paramFormat"] {
	case "eng":
		significand, exponent, _ := float2Parts(x, strIncrement(configParam["paramSigDigits"], -1))
		if exponent == "0" {
			outString = significand
		} else {
			outString = significand + "e" + exponent
		}
	case "sci":
		outString = fmt.Sprintf("%."+strIncrement(configParam["paramSigDigits"], -1)+"e", x)
	case "decimal": // decimal unless values are very large or very small in which case -- sci
		outString = fmt.Sprintf("%."+strIncrement(configParam["paramSigDigits"], -1)+"g", x)
	default:
		outString = "should not occur 02"
	}
	return
}

// convert a float64 number into a string of parts in engineering form
// returns significand (ex: 2.354 or 23.54 or 235.4)
// returns exponent (ex: 0 or 3 or 9 or -3 or -6)
// returns prefix (ex: "" or k or G or m or \mu)
func float2Parts(x float64, sigDigits string) (significand string, exponent string, prefix string) {
	var xSci string
	var result []string
	var expInt int
	var signifFloat float64
	var reGetParts = regexp.MustCompile(`(?m)^\s*(?P<res1>.*)e(?P<res2>.*)$`)
	var reZeros = regexp.MustCompile(`(?m)\.?0*$`)
	xSci = fmt.Sprintf("%."+sigDigits+"e", x)
	if reGetParts.MatchString(xSci) {
		result = reGetParts.FindStringSubmatch(xSci)
		significand = result[1]
		exponent = result[2]
		expInt, _ = strconv.Atoi(exponent)
		signifFloat, _ = strconv.ParseFloat(significand, 64)
		if expInt == -1 { // special case for as prefer to write at 0.1V instead of 100mV
			expInt = expInt + 1
			signifFloat = signifFloat / 10
		} else {
			for expInt%3 != 0 { // do until exponent is a multiple of 3 (engineering and SI notation)
				expInt = expInt - 1
				signifFloat = 10 * signifFloat
			}
		}
		significand = fmt.Sprintf("%f", signifFloat)
		significand = reZeros.ReplaceAllString(significand, "")
		exponent = strconv.Itoa(expInt)
		prefix = exponent2Prefix(exponent)
	}
	return
}

func latexStatement(statement string, varAll map[string]varSingle) string {
	var result, result2 []string
	var head, tail string
	var reWord = regexp.MustCompile(`(?m)[a-zA-Z][a-zA-Z_0-9]*`) // used to find all words in statement
	var re1 = regexp.MustCompile(`(?m)`)                         // just a way to declare re1 (it changes below)
	statement = statement + " "                                  // need extra space at end so search below works correctly if word is at end of statement
	result = reWord.FindAllString(statement, -1)
	for i := range result {
		_, ok := varAll[result[i]]
		if ok {
			re1 = regexp.MustCompile(`(?m)(?P<res1>.*\W|^)` + result[i] + `(?P<res2>\W.*)$`)
			tail = statement
			statement = ""
			for re1.MatchString(tail) {
				result2 = re1.FindStringSubmatch(tail)
				head = result2[1]
				tail = result2[2]
				statement = statement + head + varAll[result[i]].latex
			}
			statement = statement + tail
		}
	}
	return statement
}

func prefix2float(prefix string) (x float64) {
	switch prefix {
	case "a":
		x = 1e-18
	case "f":
		x = 1e-15
	case "p":
		x = 1e-12
	case "n":
		x = 1e-9
	case "\\mu":
		x = 1e-6
	case "m":
		x = 1e-3
	case "":
		x = 1
	case "k":
		x = 1e3
	case "M":
		x = 1e6
	case "G":
		x = 1e9
	case "T":
		x = 1e12
	case "P":
		x = 1e15
	default: // unrecognized prefix
	}
	return
}

func exponent2Prefix(exponent string) string {
	exp2Prefix := make(map[string]string)
	exp2Prefix["-18"] = "a"
	exp2Prefix["-15"] = "f"
	exp2Prefix["-12"] = "p"
	exp2Prefix["-9"] = "n"
	exp2Prefix["-6"] = "\\mu"
	exp2Prefix["-3"] = "m"
	exp2Prefix["0"] = ""
	exp2Prefix["3"] = "k"
	exp2Prefix["6"] = "M"
	exp2Prefix["9"] = "G"
	exp2Prefix["12"] = "T"
	exp2Prefix["15"] = "P"
	return exp2Prefix[exponent]
}

func deCommentLatex(inString string) (string, string) {
	// remove latex comments but leave \%
	// This is done for ALL lines so it takes out % in JULIA CODE AS WELL
	// might have to modify this so this func does not affect Julia code if % needed in Julia
	var comment string
	var re0 = regexp.MustCompile(`(?m)^(?P<res1>%.*)$`)     // strip off comments where % is at beginning of line
	var re1 = regexp.MustCompile(`(?m)(?P<res1>[^\\]%.*)$`) // strip off rest unless \% (since that is a real %)
	if re0.MatchString(inString) {
		comment = re0.FindStringSubmatch(inString)[1]
		inString = re0.ReplaceAllString(inString, "")
		return inString, comment
	}
	if re1.MatchString(inString) {
		comment = re1.FindStringSubmatch(inString)[1]
		inString = re1.ReplaceAllString(inString, "")
		return inString, comment
	}
	return inString, comment
}

func matchBrackets(inString, leftBrac string) (string, string) {
	// returns the first enclosed values inside outside matching brackets
	// as well as rest of string after outside matching brackets
	var inside, rightBrac, tail string
	switch leftBrac {
	case "{":
		rightBrac = "}"
	case "(":
		rightBrac = ")"
	case "[":
		rightBrac = "]"
	case "<":
		rightBrac = ">"
	default:
	}
	openBr := 0
	for i := 0; i < len(inString); i++ {
		if string(inString[i]) == leftBrac {
			openBr++
			for j := i + 1; j < len(inString); j++ {
				switch string(inString[j]) {
				case leftBrac:
					openBr++
				case rightBrac:
					openBr--
				default:
				}
				if openBr == 0 {
					inside = inString[i+1 : j]
					if j+1 <= len(inString) {
						tail = inString[j+1 : len(inString)]
					}
					return inside, tail
				}

			}

		}
	}
	return inside, tail
}

func strIncrement(inString string, k int) string {
	// take in a string representing an integer, add k to it and return incremented value as string
	i, _ := strconv.Atoi(inString)
	i = i + k
	outString := strconv.FormatInt(int64(i), 10)
	return outString
}

func function2Latex(inString string) string {
	var result []string
	var funcInput, head, tail string
	var re0 = regexp.MustCompile(`(?m)`)
	for key := range func1 {
		re0 = regexp.MustCompile(`(?mU)^(?P<res1>.*\W)(?P<res2>` + key + `\(.*)$`)
		for re0.MatchString(inString) {
			result = re0.FindStringSubmatch(inString)
			head = result[1]
			tail = result[2]
			funcInput, tail = matchBrackets(tail, "(")
			if key == "log10" { // latex command cannot be \log10 since numbers not allowed
				key = "logten" // change to logten so that \logten{} is used for latex
			}
			inString = head + "\\" + key + "{" + funcInput + "}" + tail
		}
	}
	return inString
}

func fixParll(inString string) string {
	var result []string
	var outString, head, tail, inside, var1, var2 string
	var reParll = regexp.MustCompile(`(?mU)^(?P<res1>.*)parll(?P<res2>\(.*)$`)
	var reInside = regexp.MustCompile(`(?m)^(?P<res1>.*),(?P<res2>.*)$`)
	outString = inString // default if matching below does not occur
	for reParll.MatchString(outString) {
		if reParll.MatchString(outString) {
			result = reParll.FindStringSubmatch(outString)
			head = result[1]
			inside, tail = matchBrackets(result[2], "(")
			inside = fixParll(inside)
			if reInside.MatchString(inside) {
				result = reInside.FindStringSubmatch(inside)
				var1 = result[1]
				var2 = result[2]
				outString = head + var1 + "||" + var2 + tail
			}
		}
	}
	return outString
}

func randInt(N, random int) int {
	// based on random (a random number), choose an
	// int from 0 to N-1
	var choice int
	choice = random % N
	return choice
}

func psuedoRand(x0 int) int {
	// A linear congruential generator (LCG) based on
	// https://en.wikipedia.org/wiki/Linear_congruential_generator
	// it returns an psuedorandom integer between 100000 and 999999
	var a, c, m, x1 int
	if x0 < 0 { // correct x0 if it happens to be less than 0
		x0 = -1 * x0
	}
	a = 707106 // 1e6/sqrt(2) and truncated
	c = 1
	m = 999999
	x1 = 0
	for x1 < 100000 {
		x1 = (a*x0 + c) % m
		x0 = x1
	}
	return x1
}

func checkReserved(variable, logOut string) (string, string) {
	var key string
	switch variable {
	case "parll": // can add more reserved variables here
		logOut = logOut + variable + " is a reserved variable and cannot be assigned"
		variable = variable + "IsReservedVariable"
	default:
	}
	for key = range func1 {
		if variable == key {
			logOut = logOut + key + " is a reserved variable and cannot be assigned"
			variable = key + "IsReservedVariable"
		}
	}
	return variable, logOut
}

func syntaxError(statement, cmdType string) string {
	var errCode, shouldBeBlank string
	// valid characters in statement
	var reValidRunCode0 = regexp.MustCompile(`(?m)[\w|=|*|/|+|\-|^|(|)|\s|\.|,]+`) // valid characters in statement
	var reValidVal0 = regexp.MustCompile(`(?m)[\w|\s]+`)
	var reValidVal1 = regexp.MustCompile(`(?m)[\w|*|/|+|\-|^|(|)|\s|\.|,]+`)
	var reValidRunParam0 = regexp.MustCompile(`(?m)[\w|=|\s|\.|,|#|\\|\[|\]|:]+`)

	// first check if all characters in statement are valid characters
	switch cmdType {
	case "val":
		shouldBeBlank = reValidVal1.ReplaceAllString(statement, "")
		if shouldBeBlank != "" {
			errCode = "character(s) " + shouldBeBlank + " should not be in a \\val statement"
		}
	case "val=":
		shouldBeBlank = reValidVal0.ReplaceAllString(statement, "")
		if shouldBeBlank != "" {
			errCode = "character(s) " + shouldBeBlank + " should not be in a \\val= statement"
		}
	case "runCode":
		shouldBeBlank = reValidRunCode0.ReplaceAllString(statement, "")
		if shouldBeBlank != "" {
			errCode = "character(s) " + shouldBeBlank + " should not be in a \\run statement"
		}
	case "runParam":
		shouldBeBlank = reValidRunParam0.ReplaceAllString(statement, "")
		if shouldBeBlank != "" {
			errCode = "character(s) " + shouldBeBlank + " should not be in a \\runParam statement"
		}
	default:
		errCode = "should not be here in syntaxError"
	}
	return errCode
}
