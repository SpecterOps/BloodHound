package alluregen

import "crypto/sha256"

func hashString(testtCaseName string) []byte {
	hash := sha256.New()
	hash.Reset()
	hash.Write([]byte(testtCaseName))
	return hash.Sum(nil)
}
