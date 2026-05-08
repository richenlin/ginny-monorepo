package aes

import "testing"

// TestECB xxx
func TestECB(t *testing.T) {
	if err := testCipher(NewECB()); err != nil {
		t.Fatal(err)
	}
}
