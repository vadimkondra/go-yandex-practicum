package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

const Header = "HashSHA256"

func Calculate(data []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)

	return hex.EncodeToString(h.Sum(nil))
}

func Check(data []byte, key string, expectedHash string) bool {
	if key == "" {
		return true
	}

	actualHash := Calculate(data, key)

	return hmac.Equal([]byte(actualHash), []byte(expectedHash))
}
