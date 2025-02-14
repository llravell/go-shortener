package main

import "os"

func wrappedExit() {
	os.Exit(1)
}

func main() {
	os.Exit(1) // want "os exit call in main function"
	wrappedExit()
}
