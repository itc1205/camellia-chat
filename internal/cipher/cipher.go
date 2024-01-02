// Package cipher describes an Cipher interface for encrypting and decrypting data,
// that will be used to implement an different types of encryption algorithms.
package cipher

// Cipher interface is used internally for decrypting and encrypting data inside of TCP-chat.
type Cipher interface {
	Decrypt(dst, src []byte)
	Encrypt(dst, src []byte)
	BlockSize() int
}
