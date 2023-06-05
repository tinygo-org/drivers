package lorawan

import "crypto/rand"

// reverseBytes reverses order of a given byte slice
func reverseBytes(s []byte) []byte {
	result := make([]byte, len(s))
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = s[j], s[i]
	}
	return result
}

// GetRand16 returns 2 random bytes
func GetRand16() ([2]uint8, error) {
	var randomBytes [2]byte
	_, err := rand.Read(randomBytes[:])

	return randomBytes, err
}
