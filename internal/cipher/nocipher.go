package cipher

// NoEncryptionCipher structure is an type of encryption without any encryption.
type NoEncryptionCipher struct{}

const NoEncBlockSize = 16

// Encrypt function will return an unchanged data, becuase this type does not have any encryption.
func (ne NoEncryptionCipher) Decrypt(dst, src []byte) {
	copy(dst, src)
}

// Encrypt function will return an unchanged data, becuase this type does not have any encryption.
func (ne NoEncryptionCipher) Encrypt(dst, src []byte) {
	copy(dst, src)
}

func (ne NoEncryptionCipher) BlockSize() int {
	return BlockSize
}
