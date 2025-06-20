package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
)

func main() {
	var secretKey = ""
	var message = "hello"

	flag.StringVar(&secretKey, "s", secretKey, "secret key that will be used to sign message")
	flag.StringVar(&message, "m", message, "message to encode")
	flag.Parse()

	fmt.Printf("Secret Key: '%s'\nmessage: '%s'\n", secretKey, message)

	h := hmac.New(sha256.New, []byte(secretKey))
	_, err := h.Write([]byte(message))
	if err != nil {
		fmt.Println("error happened %w", err)
	} else {

		fmt.Println(hex.EncodeToString(h.Sum(nil)))
	}
}
