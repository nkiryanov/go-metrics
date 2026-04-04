package forbidden

import (
	"log"
	"os"
)

func helper() {
	os.Exit(1) // want "os.Exit is only allowed in main\\(\\) of package main"
}

type T struct{}

func (T) method() {
	os.Exit(1) // want "os.Exit is only allowed in main\\(\\) of package main"
}

func withClosure() {
	f := func() {
		os.Exit(1) // want "os.Exit is only allowed in main\\(\\) of package main"
	}
	f()
}

func logFatals() {
	log.Fatal("x")       // want "log.Fatal is only allowed in main\\(\\) of package main"
	log.Fatalf("%s", "") // want "log.Fatalf is only allowed in main\\(\\) of package main"
	log.Fatalln("x")     // want "log.Fatalln is only allowed in main\\(\\) of package main"
}

func loggerFatals() {
	l := log.Default()
	l.Fatal("x")       // want "\\(\\*log.Logger\\).Fatal is only allowed in main\\(\\) of package main"
	l.Fatalf("%s", "") // want "\\(\\*log.Logger\\).Fatalf is only allowed in main\\(\\) of package main"
	l.Fatalln("x")     // want "\\(\\*log.Logger\\).Fatalln is only allowed in main\\(\\) of package main"
}
