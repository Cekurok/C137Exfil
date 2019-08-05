package main

import (
	"fmt"
	"regexp"
)

// Contiains Method to check if an int is present in a int slice
func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// Contiains Method to check if an string is present in a string slice
func containsStr(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// Main logging function
func logData(data string, nl bool, err bool) {

	prefix := "[INFO]: "

	if err {
		prefix = "[ERROR]: "
	}

	if silent == false {
		if nl {
			fmt.Println(prefix + data)
		} else {
			fmt.Println(prefix + data)
		}

	}
}

// Fixes UTF Chars from a string of data.
func fixUTF(data string) string {
	re := regexp.MustCompile("[[:^ascii:]]")
	return re.ReplaceAllLiteralString(data, "")
}
