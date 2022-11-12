package utils

import (
	"math/rand"
	"time"
)

var allowedCharacters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

// Random returns a random string of the given length
func Random(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, length)
	for i := range b {
		b[i] = allowedCharacters[rand.Intn(len(allowedCharacters))]
	}

	return string(b)
}
