package aes

import "testing"

// TestCBC
func TestCBC(t *testing.T) {
	if err := testCipher(NewCBC()); err != nil {
		t.Fatal(err)
	}
}
