package main

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

func hashBytes(bytes []byte) string {
	h := md5.New()
	h.Write(bytes)
	return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}

func compareHash(a string, b string) bool {
	normalize := func(input string) string {
		return strings.TrimRight(strings.TrimLeft(input, "\""), "\"")
	}
	return normalize(a) == normalize(b)
}
