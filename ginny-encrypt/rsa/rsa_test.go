package rsa

import (
	"bytes"
	"fmt"
	"testing"

	encrypt "github.com/goriller/ginny-encrypt"
)

var (
	testPublicKey = []byte(`-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDXJboJEAn053HlW3Z1ww31a5mt
XhimAnyXFAneW9m6AgpcwX2oG4YVRu0tl+gxojvx7jlD07uAegqbW1GEi+HGPJVf
TUHRFAaUgzXB0eYnDWccRbnrrqKaPwNjepyph1V9UJk868gVUTTix8oxmKCN9zKX
4iLPzqWdjQWk5meAYQIDAQAB
-----END PUBLIC KEY-----`)
	testPrivateKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQDXJboJEAn053HlW3Z1ww31a5mtXhimAnyXFAneW9m6AgpcwX2o
G4YVRu0tl+gxojvx7jlD07uAegqbW1GEi+HGPJVfTUHRFAaUgzXB0eYnDWccRbnr
rqKaPwNjepyph1V9UJk868gVUTTix8oxmKCN9zKX4iLPzqWdjQWk5meAYQIDAQAB
AoGAZwlr46w5QH9ZdjEL9hEQzdEW28cdQeAeABK6OTI+/0y73rlR7yEjYWxC6Zt/
OcoLMG3ZIgk0mq6YBthAnZyKZs8x+nA6eRAgE9rOxE8NxPN7VIpTUYRODXCJyCwh
ssxJNYkdPrx7uAqnWLpVj2FZ9QsukxTHOpqS7/RKVko9IcsCQQD8zeFU6LnhAmaH
Gl5xZ+pk8MqTU9TZXQ0ZMzajFv2yhaujZV3vMma4vvuUdHsddp/eb5nEACsx2YM8
Tv2bs2SvAkEA2d37cBrMS1/Ky8u5VIsbj8AKUsEVPdkaK7AjqJRGmT3dxRP/RZSd
Au5BZXQQpHLXOqzH29A2KyyxxOUCc23P7wJBAL4IRQn+pzttAoUsXTICWz/lgWGd
8rIyMFZxGPEfpzU7Jfp9iE72JCFb7uF5bdKICUS7v2qGdfHS/8Ol3R3djCECQCol
XSy0omy6XTrLcFDAkFZgqh6UJ43NX9ivvFYySO4AH9SuJ6XIOA+HE7OSnl2Rsb0y
C3+kabY0cTdLrguyZJUCQQCRQiYVRzUYTVN6/h5FHAdy6hD/vxaw95pctBbgZGkB
k60TNZTfxS+v1/QPti+PCQm6V9KQmd01gMBygCOSyUmY
-----END RSA PRIVATE KEY-----`)
	testPlainData = []byte(`123456789012345678`)
)

func testCipher(cipher encrypt.Cipher) error {
	cipherData, err := cipher.Encrypt(testPublicKey, testPlainData)
	if err != nil {
		return err
	}

	plainData, err := cipher.Decrypt(testPrivateKey, cipherData)
	if err != nil {
		return err
	}

	if !bytes.Equal(plainData, testPlainData) {
		return fmt.Errorf("data not match")
	}
	return nil
}

// TestRSA xxx
func TestRSA(t *testing.T) {
	if err := testCipher(New()); err != nil {
		t.Fatal(err)
	}
}
