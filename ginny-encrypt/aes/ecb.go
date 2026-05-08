package aes

import (
	"crypto/aes"

	encrypt "github.com/goriller/ginny-encrypt"
)

// NewECB 创建实现了ECB模式的加密器
func NewECB() encrypt.Cipher {
	return &ecb{}
}

// ecb 实现AES ECB模式加密
type ecb struct{}

// Encrypt 加密数据
func (a *ecb) Encrypt(key []byte, data []byte) ([]byte, error) {
	cipherBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// 判断加密块的大小
	blockSize := cipherBlock.BlockSize()
	data = pkcs7Padding(data, blockSize)
	var encryptData = make([]byte, len(data))
	// 存储每次加密的数据
	tmpData := make([]byte, blockSize)

	// 分组分块加密
	for index := 0; index < len(data); index += blockSize {
		cipherBlock.Encrypt(tmpData, data[index:index+blockSize])
		copy(encryptData[index:index+blockSize], tmpData)
	}
	return encryptData, nil
}

// ecb Decrypt
func (a *ecb) Decrypt(key []byte, data []byte) ([]byte, error) {
	cipherBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := cipherBlock.BlockSize()
	decryptData := make([]byte, len(data))
	// 存储每次加密的数据
	tmpData := make([]byte, blockSize)

	// 分组分块加密
	for index := 0; index < len(data); index += blockSize {
		cipherBlock.Decrypt(tmpData, data[index:index+blockSize])
		copy(decryptData[index:index+blockSize], tmpData)
	}

	return pkcs7UnPadding(decryptData)
}
