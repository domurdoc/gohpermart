package utils

import "crypto/rand"

const ALPHA = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func GenerateRandomString(charSet string, length int) string {
	// https://stackoverflow.com/a/67035900
	charSetLength := len(charSet)
	buf := make([]byte, length)
	rand.Read(buf)
	for i := range length {
		buf[i] = charSet[int(buf[i])%charSetLength]
	}
	return string(buf)
}
