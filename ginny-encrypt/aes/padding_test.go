package aes

import (
	"bytes"
	"fmt"

	encrypt "github.com/goriller/ginny-encrypt"
)

var (
	testEncryptKey = []byte(`q4L9LsrZwjuJDTnF`)
	testPlainData  = []byte(`123456789012345678`)
)

func testCipher(cipher encrypt.Cipher) error {
	cipherData, err := cipher.Encrypt(testEncryptKey, testPlainData)
	if err != nil {
		return err
	}

	plainData, err := cipher.Decrypt(testEncryptKey, cipherData)
	if err != nil {
		return err
	}

	if !bytes.Equal(plainData, testPlainData) {
		return fmt.Errorf("data not match")
	}
	return nil
}
