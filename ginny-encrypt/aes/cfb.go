package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	encrypt "github.com/goriller/ginny-encrypt"
)

// NewCFB 创建实现了CFB模式的加密器
func NewCFB() encrypt.Cipher {
	return &cfb{}
}

// cfb 实现了AES CFB模式加密
type cfb struct{}

// Encrypt 加密数据
func (c *cfb) Encrypt(key []byte, data []byte) ([]byte, error) {
	// 创建加密实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	encrypted := make([]byte, aes.BlockSize+len(data))
	iv := encrypted[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(encrypted[aes.BlockSize:], data)
	return encrypted, nil
}

// Decrypt 解密数据
func (c *cfb) Decrypt(key []byte, data []byte) ([]byte, error) {
	// 创建实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(data) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	iv := data[:aes.BlockSize]
	encrypted := data[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(encrypted, encrypted)
	return encrypted, nil
}
