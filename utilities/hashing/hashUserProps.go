package hashing

import (
	"crypto/sha256"
	"fmt"
)

func HashUserPassword(password string) string {
	passwordDigestSum := sha256.Sum256([]byte(password))
	passwordDigest := fmt.Sprintf("%x", passwordDigestSum)
	return passwordDigest
}