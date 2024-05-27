package main

import "fmt"

const VERSION = "1.0.37"
const REVISION = "e0ec94d22bfdb85474715061233be5f04f723f21+1"
const NUMBER = 239

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
