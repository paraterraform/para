package utils

import (
	"crypto/sha256"
	"fmt"
)

func HashString(input string) string {
	hash := sha256.New()
	hash.Write([]byte(input))
	return fmt.Sprintf("%x", hash.Sum(nil))
}
