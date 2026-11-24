package tokenhash

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

func Hash(token string) string {
	sum := sha256.Sum256([]byte(token))

	return hex.EncodeToString(sum[:])
}

func Equal(storedHash, token string) bool {
	return subtle.ConstantTimeCompare([]byte(storedHash), []byte(Hash(token))) == 1
}
