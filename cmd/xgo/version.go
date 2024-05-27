package main

import "fmt"

const VERSION = "1.0.37"
const REVISION = "9fae61a230495e8d67a045ba1d7eed0eeaade5d2+1"
const NUMBER = 240

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
