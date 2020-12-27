package main

import (
	"strconv"
	"strings"
	"testing"
)

func TestMakeTex(t *testing.T) {
	var inFile, outFile fileInfo
	var testNames = []string{"test01"}

	for i := range testNames {
		problemInput, _ := fileReadString("./bigTestInputs/" + testNames[i] + ".prb")
		expectedTexOut, _ := fileReadString("./bigTestInputs/" + testNames[i] + ".tex")
		expectedOut := strings.Split(expectedTexOut, "\n")
		texOut := makeTex(problemInput, "4", "false", inFile, outFile)
		actualOut := strings.Split(texOut, "\n")
		for j := range actualOut {
			if actualOut[j] != expectedOut[j] {
				jStr := strconv.Itoa(j)
				t.Error(testNames[i]+" line "+jStr+" Failed: {} expected {} received {} ... {}", expectedOut[j], "{}", actualOut[j], "{}")
			}
		}
	}
}
