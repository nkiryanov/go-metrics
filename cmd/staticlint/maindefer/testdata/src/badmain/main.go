package main

import "fmt"

func main() {
	defer fmt.Println("cleanup") // want "defer in main\\(\\): use a run\\(\\) helper instead"

	if true {
		defer fmt.Println("nested") // want "defer in main\\(\\): use a run\\(\\) helper instead"
	}
}
