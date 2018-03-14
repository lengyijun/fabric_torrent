package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func main(){
	mykey := make([]byte, 32)
	rand.Read(mykey)
	m:=hex.EncodeToString(mykey)
	fmt.Printf(m)
}