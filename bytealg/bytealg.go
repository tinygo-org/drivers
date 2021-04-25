package bytealg

// equal checks if two byte slices are equal.
// It is equivalent to bytes.equal but no
func equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// The following code has been copied from the Go 1.15 release tree.

// primerRK is the prime base used in Rabin-Karp algorithm.
const primerRK = 16777619

// idxRabinKarpBytes uses the Rabin-Karp search algorithm to return the index of the
// first occurence of substr in s, or -1 if not present.
func IdxRabinKarpBytes(s, substr []byte) int {
	// Handle edge cases
	if len(s) == 0 {
		if len(substr) == 0 {
			return 0
		}
		return -1
	}
	// Rabin-Karp search
	hashsep, pow := hashStrBytes(substr)
	n := len(substr)
	var h uint32
	for i := 0; i < n; i++ {
		h = h*primerRK + uint32(s[i])
	}
	if h == hashsep && equal(s[:n], substr) {
		return 0
	}
	for i := n; i < len(s); {
		h *= primerRK
		h += uint32(s[i])
		h -= pow * uint32(s[i-n])
		i++
		if h == hashsep && equal(s[i-n:i], substr) {
			return i - n
		}
	}
	return -1
}

// hashStrBytes returns the hash and the appropriate multiplicative
// factor for use in Rabin-Karp algorithm.
func hashStrBytes(sep []byte) (uint32, uint32) {
	hash := uint32(0)
	for i := 0; i < len(sep); i++ {
		hash = hash*primerRK + uint32(sep[i])
	}
	var pow, sq uint32 = 1, primerRK
	for i := len(sep); i > 0; i >>= 1 {
		if i&1 != 0 {
			pow *= sq
		}
		sq *= sq
	}
	return hash, pow
}
