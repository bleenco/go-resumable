package main

import (
	"crypto/rand"
	"fmt"
)

func fatalErrCheck(err error) {
	if err != nil {
		fmt.Printf("%v", err)
		panic(err)
	}
}

func generateSessionID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%X", b)
}
