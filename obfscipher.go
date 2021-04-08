package main

import "crypto/sha512"

// ObfsCipher is a small stream cipher intended to obfuscate the resume data.
type ObfsCipher struct {
	state [64]byte
}

// NewCipher instantiates the obfuscating cipher.
func NewCipher(passphrase string) ObfsCipher {
	sum := sha512.Sum512([]byte(passphrase))
	for i := 0; i < 10000; i++ {
		sum = sha512.Sum512(sum[:])
	}
	return ObfsCipher{state: sum}
}

// Pad encrypts or decrypts the provided byte stream.
func (c ObfsCipher) Pad(p []byte) {
	for i := range p {
		p[i] ^= c.state[0] & 0x7F
		c.state = sha512.Sum512(c.state[:])
	}
}

// Pad encrypts or decrypts the provided byte stream.
func (c ObfsCipher) PadStr(p *string) {
	b := []byte(*p)
	c.Pad(b)
	*p = string(b)
}
