package service

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

func GenerateToken(username string, orgID uint, userID uint) (string, error) {
	claims := jwt.MapClaims{
		"sub":     username,
		"org_id":  orgID,
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 8).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
