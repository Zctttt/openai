package md5

import (
	"crypto/md5"
	"encoding/base64"
)

func GetMD5(str string) string {
	hash := md5.Sum([]byte(str))
	encodedHash := base64.StdEncoding.EncodeToString(hash[:])
	return encodedHash
}

func ValidateMD5(cleartext string, ciphertext string) bool {
	md5Str := GetMD5(cleartext)
	return md5Str == ciphertext
}
