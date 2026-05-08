package rsa

import (
	"crypto/rand"
	stdRSA "crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	encrypt "github.com/goriller/ginny-encrypt"
)

// New 获取RSA加密器
//  使用公钥加密、私钥解密
func New() encrypt.Cipher {
	return &rsa{}
}

type rsa struct{}

// Encrypt 加密数据
func (r *rsa) Encrypt(key []byte, data []byte) ([]byte, error) {
	block, _ := pem.Decode(key)
	if block == nil {
		return nil, fmt.Errorf("no pem data found")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	publicKey, ok := pubInterface.(*stdRSA.PublicKey)
	if !ok {
		return nil, fmt.Errorf("unrecognized public key")
	}

	return stdRSA.EncryptPKCS1v15(rand.Reader, publicKey, data)
}

// Decrypt 解密数据
func (r *rsa) Decrypt(key []byte, data []byte) ([]byte, error) {
	block, _ := pem.Decode(key)
	if block == nil {
		return nil, fmt.Errorf("no pem data found")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return stdRSA.DecryptPKCS1v15(rand.Reader, privateKey, data)
}
