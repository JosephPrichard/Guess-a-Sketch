package utils

import (
	crand "crypto/rand"
	"encoding/hex"
	mrand "math/rand"
)

var alphanumericRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func GenerateCode(len int) (string, error) {
	b := make([]byte, len/2)
	_, err := crand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func AlphaNumeric(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = alphanumericRunes[mrand.Intn(len(alphanumericRunes))]
	}
	return string(b)
}