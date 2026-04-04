package main

import "fmt"

func main() {
	// defer inside a closure is fine — closure has its own scope
	cleanup := func() {
		defer fmt.Println("in closure")
	}
	cleanup()

	// defer inside nested closure in a block is also fine
	if true {
		func() {
			defer fmt.Println("nested closure")
		}()
	}
}
