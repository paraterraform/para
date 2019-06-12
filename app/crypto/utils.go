package crypto

import (
	"crypto/sha256"
	"fmt"
)

func DefaultStringHash(input string) string {
	hash := sha256.New()
	hash.Write([]byte(input))
	return fmt.Sprintf("%x", hash.Sum(nil))
}
