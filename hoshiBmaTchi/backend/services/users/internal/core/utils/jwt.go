package utils

import (
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET_KEY"))

func GenerateTokens(userID string, email string) (string, string, error){

	if len(jwtSecret) == 0 {
		log.Fatal("FATAL: JWT_SECRET_KEY environment variable is not set.")
	}

	accessTokenClaims := jwt.MapClaims{
		"user_id": userID,
		"email": email,
		"exp": time.Now().Add(time.Minute * 15).Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString(jwtSecret)
	if err != nil{
		return "", "", err
	}

	refreshTokenClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString(jwtSecret)

	if err != nil{
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}