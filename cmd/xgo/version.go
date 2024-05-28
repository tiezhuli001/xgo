package main

import "fmt"

const VERSION = "1.0.37"
const REVISION = "6f9a355d360e70a797c4ca0903fb5a6bdd1aa5db+1"
const NUMBER = 240

func getRevision() string {
	revSuffix := ""
	if isDevelopment {
		revSuffix = "_DEV"
	}
	return fmt.Sprintf("%s %s%s BUILD_%d", VERSION, REVISION, revSuffix, NUMBER)
}
