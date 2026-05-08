package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	encrypt "github.com/goriller/ginny-encrypt"
)

// NewGCM 返回加密数据携带GMAC认证码的加密器
func NewGCM() encrypt.Cipher {
	return &gcm{}
}

// gcm
type gcm struct{}

// Encrypt 加密数据
func (a *gcm) Encrypt(key []byte, data []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	bResult := gcm.Seal(nonce, nonce, data, nil)
	return bResult, nil
}

// Decrypt 解密数据
func (a *gcm) Decrypt(key []byte, data []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, err
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
