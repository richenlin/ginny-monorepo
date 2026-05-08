package aes

import "testing"

// TestGCM xxx
func TestGCM(t *testing.T) {
	if err := testCipher(NewGCM()); err != nil {
		t.Fatal(err)
	}
}
