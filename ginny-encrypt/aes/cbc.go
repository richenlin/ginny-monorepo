package aes

import (
	"crypto/aes"
	"crypto/cipher"

	encrypt "github.com/goriller/ginny-encrypt"
)

// NewCBC 创建实现了CBC模式的加密器
func NewCBC() encrypt.Cipher {
	return &cbc{}
}

// cbc 实现了AES CBC模式加密
type cbc struct{}

// Encrypt 加密数据
func (c *cbc) Encrypt(key []byte, data []byte) ([]byte, error) {
	// 创建加密实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// 判断加密快的大小
	blockSize := block.BlockSize()
	// 填充
	encryptBytes := pkcs7Padding(data, blockSize)
	// 初始化加密数据接收切片
	crypted := make([]byte, len(encryptBytes))
	// 使用cbc加密模式
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	// 执行加密
	blockMode.CryptBlocks(crypted, encryptBytes)
	return crypted, nil
}

// Decrypt 解密数据
func (c *cbc) Decrypt(key []byte, data []byte) ([]byte, error) {
	// 创建实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// 获取块的大小
	blockSize := block.BlockSize()
	// 使用cbc
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	// 初始化解密数据接收切片
	crypted := make([]byte, len(data))
	// 执行解密
	blockMode.CryptBlocks(crypted, data)
	// 去除填充
	crypted, err = pkcs7UnPadding(crypted)
	if err != nil {
		return nil, err
	}
	return crypted, nil
}
