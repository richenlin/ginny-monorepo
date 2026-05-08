package encrypt

// Cipher 加密器
type Cipher interface {
	// Encrypt 信息加密
	Encrypt(key []byte, data []byte) ([]byte, error)
	// Decrypt 信息解密
	Decrypt(key []byte, data []byte) ([]byte, error)
}
