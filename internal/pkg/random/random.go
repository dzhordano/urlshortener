package random

import (
	"math/rand/v2"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

// NewRandomString return random string of length l.
//
// Just returns pseudo-random string.
//
//nolint:gosec // No need to care about security here.
func NewRandomString(l int) string {
	r := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	b := make([]byte, l)
	for i := range b {
		b[i] = letters[r.IntN(len(letters))]
	}

	return string(b)
}
