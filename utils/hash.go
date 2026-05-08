package utils

import "golang.org/x/crypto/bcrypt"

func HashPasswd(p string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(p), 14)
	return string(bytes), err
}

func ComparePasswd(p string, h string) error {
	err := bcrypt.CompareHashAndPassword([]byte(p), []byte(h))
	return err
}
