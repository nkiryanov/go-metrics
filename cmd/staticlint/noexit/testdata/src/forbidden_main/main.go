package main

import "os"

func init() {
	os.Exit(1) // want "os.Exit is only allowed in main\\(\\) of package main"
}

func main() {
	os.Exit(0)
}
