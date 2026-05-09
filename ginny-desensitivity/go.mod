module github.com/goriller/ginny-desensitivity/v2

go 1.22

require (
	github.com/goriller/ginny-encrypt/v2 v2.0.0
	github.com/stretchr/testify v1.8.0
)

replace (
	github.com/goriller/ginny-encrypt/v2 => ../ginny-encrypt
)
