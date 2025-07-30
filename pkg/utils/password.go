package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword는 평문 비밀번호를 해시화합니다
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword는 평문 비밀번호와 해시된 비밀번호를 비교합니다
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
