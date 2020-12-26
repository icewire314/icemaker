package main

import "testing"

func TestCheckRandom(t *testing.T) {
	var tests = []struct {
		randomStr string
		logOut    string
		random    int
		outLogOut string
	}{
		{"2", "", 2, ""},
		{"false", "", 0, ""},
		{"true", "", -1, ""},
		{"-10", "oldLogOut", 0, "oldLogOutrandom should be a positive integer\n"},
		{"1e1", "", 0, "random should be either \"false\", \"true\", or a positive integer\n"},
	}
	for _, test := range tests {
		random, logOut := checkRandom(test.randomStr, test.logOut)
		if random != test.random {
			t.Error("Test Failed: {} inputted, {} expected, recieved: {}", test.randomStr, test.random, random)
		}
		if logOut != test.outLogOut {
			t.Error("Test Failed: {} inputted, {} expected, recieved: {}", test.logOut, test.outLogOut, logOut)
		}
	}
}
