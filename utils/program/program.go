package program

import (
	"encoding/hex"
	"math/rand"

	"golang.org/x/crypto/argon2"
)

func Hash(text, salt []byte) string {
	return hex.EncodeToString(argon2.Key([]byte(text), []byte(salt), 3, 32*1024, 4, 32))
}

func RandString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
