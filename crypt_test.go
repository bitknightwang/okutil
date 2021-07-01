package okutil

import "testing"

func TestEncrypt(t *testing.T) {
	key := deriveKey("password", nil)

	encryptMsg, _ := Encrypt(key, "Hello World")
    t.Logf("encrypted: %v\n", encryptMsg)

	msg, _ := Decrypt(key, encryptMsg)
    t.Logf("decrypted: %v\n", msg)
}

