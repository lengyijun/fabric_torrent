package main

import (
	"crypto/aes"
	"crypto/cipher"
	"io"
	"os"
	"log"
	"encoding/hex"
)

//sample
//./main 14189dc35ae35e75ff31d7502e245cd9bc7803838fbfd5c773cdcd79b8a28bbd encrypted-file decrypted-file

//args[0]  key
//args[1]  encryptedfile
//args[2]  decryptedfilepath
func main() {
	argsWithoutProg := os.Args[1:]
	if len((argsWithoutProg))!=3{
		log.Fatal("need 3 parameter")
		return
	}
	// Load your secret key from a safe place and reuse it across multiple
	// NewCipher calls. (Obviously don't use this example key for anything
	// real.) If you want to convert a passphrase to a key, use a suitable
	// package like bcrypt or scrypt.
	key, err := hex.DecodeString(argsWithoutProg[0])
	if err!=nil{
		log.Fatalln(err)
	}

	inFile, err := os.Open(argsWithoutProg[1])
	if err != nil {
		log.Fatalln(err)
		panic(err)
	}
	defer inFile.Close()

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatalln(err)
		panic(err)
	}

	// If the key is unique for each ciphertext, then it's ok to use a zero
	// IV.
	var iv [aes.BlockSize]byte
	stream := cipher.NewOFB(block, iv[:])

	outFile, err := os.OpenFile(argsWithoutProg[2], os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalln(err)
		panic(err)
	}
	defer outFile.Close()

	reader := &cipher.StreamReader{S: stream, R: inFile}
	// Copy the input file to the output file, decrypting as we go.
	if _, err := io.Copy(outFile, reader); err != nil {
		log.Fatalln(err)
		panic(err)
	}
}

