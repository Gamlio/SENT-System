// internal/auth/security.go
package auth

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateToken(username string, orgID uint) (string, error) { // <--- THÊM orgID VÀO ĐÂY
	claims := jwt.MapClaims{
		"sub":    username,
		"org_id": orgID, // <--- ĐÓNG DẤU MÃ CÔNG TY VÀO TOKEN
		"exp":    time.Now().Add(time.Hour * 8).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("SECRET_KEY")))
}
