package lib

import "fmt"

func helper() {
	defer fmt.Println("in lib") // fine — not package main
}
