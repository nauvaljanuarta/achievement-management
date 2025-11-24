package utils

import (
	"errors"
	"log" 
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"achievement-backend/app/models" 
)

// Fungsi helper private
func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("FATAL: JWT_SECRET environment variable is not set. Application cannot start.")
	}
	return []byte(secret)
}

func GenerateToken(user models.User, roleName string) (string, error) {
	claims := models.JWTClaims{
		UserID: user.ID.String(),
		Email:  user.Email,
		Role:   roleName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), 
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "sistem-prestasi-mahasiswa",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

func ValidateToken(tokenString string) (*models.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return getJWTSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*models.JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}