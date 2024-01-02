package cipher

import (
	"testing"
)

// This function just checks if we can use our encryption as part of the Cypher interface.
func UseAsInteface(enc Cipher) {}

// This test will check the encryption function of the struct NoEncryptionCypher.
func TestNoEcryptionEncrypt(t *testing.T) {
	enc := NoEncryptionCipher{}
	// Some unicode-specific data to check if encryption is UTF-8 compatable.
	data := "Hello! World12341251123(◕‿◕)!_3123"
	// Check if we can use it as Cypher interface...
	UseAsInteface(enc)
	// And make sure that we can encrypt some data and recieve the same data after decrypting
	encrypted_data := make([]byte, 16)
	enc.Encrypt(encrypted_data, ([]byte)(data[:16]))
	decrypted_data := make([]byte, 16)
	enc.Decrypt(decrypted_data, encrypted_data)

	if (string)(decrypted_data) != data[:16] {
		t.Fatalf("decrypted_data: %s != data: %s\n", decrypted_data, data)
	}
}
