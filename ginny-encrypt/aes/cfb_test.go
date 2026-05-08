package aes

import "testing"

// TestCFB
func TestCFB(t *testing.T) {
	if err := testCipher(NewCFB()); err != nil {
		t.Fatal(err)
	}

}
