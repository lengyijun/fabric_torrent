package main

import (
	"os"
	"crypto/aes"
	"crypto/cipher"
	"io"
	"crypto/rand"
	"github.com/syndtr/goleveldb/leveldb"
	"errors"
	"encoding/hex"
)

//encrypt file from folder origindataPath to encryptdataPath
func encryptFile(filename string) (string,error) {
	// Load your secret key from a safe place and reuse it across multiple
	// NewCipher calls. (Obviously don't use this example key for anything
	// real.) If you want to convert a passphrase to a key, use a suitable
	// package like bcrypt or scrypt.
	mykey := make([]byte, 32)
	_, err := rand.Read(mykey)
	if err != nil {
		// handle error here
	}
	db, err := leveldb.OpenFile("key.db", nil)
	defer db.Close()
	err = db.Put([]byte(filename), mykey, nil)
	if err!=nil{
		return "",errors.New("unable to put key in db")
	}

	inFile, err := os.Open(origindataPath+"/"+filename)
	if err != nil {
		panic(err)
	}
	defer inFile.Close()

	block, err := aes.NewCipher(mykey)
	if err != nil {
		panic(err)
	}

	// If the key is unique for each ciphertext, then it's ok to use a zero
	// IV.
	var iv [aes.BlockSize]byte
	stream := cipher.NewOFB(block, iv[:])

	outFile, err := os.OpenFile(encryptdataPath+"/"+filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	writer := &cipher.StreamWriter{S: stream, W: outFile}
	// Copy the input file to the output file, encrypting as we go.
	if _, err := io.Copy(writer, inFile); err != nil {
		panic(err)
	}
	return hex.EncodeToString(mykey),nil

	// Note that this example is simplistic in that it omits any
	// authentication of the encrypted data. If you were actually to use
	// StreamReader in this manner, an attacker could flip arbitrary bits in
	// the decrypted result.
}

//decrypt file from folder encryptdataPath to decryptdataPath
func decryptFile(filename string) error {
	// Load your secret key from a safe place and reuse it across multiple
	// NewCipher calls. (Obviously don't use this example key for anything
	// real.) If you want to convert a passphrase to a key, use a suitable
	// package like bcrypt or scrypt.
	db, err := leveldb.OpenFile("key.db", nil)
	defer db.Close()
	key,err := db.Get([]byte(filename),nil)
	if err!=nil{
		return errors.New("cannot get key from db")
	}

	inFile, err := os.Open(encryptdataPath+"/"+filename)
	if err != nil {
		panic(err)
	}
	defer inFile.Close()

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// If the key is unique for each ciphertext, then it's ok to use a zero
	// IV.
	var iv [aes.BlockSize]byte
	stream := cipher.NewOFB(block, iv[:])

	outFile, err := os.OpenFile(decryptdataPath+"/"+filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	reader := &cipher.StreamReader{S: stream, R: inFile}
	// Copy the input file to the output file, decrypting as we go.
	if _, err := io.Copy(outFile, reader); err != nil {
		panic(err)
	}
	return nil

	// Note that this example is simplistic in that it omits any
	// authentication of the encrypted data. If you were actually to use
	// StreamReader in this manner, an attacker could flip arbitrary bits in
	// the output.
}
