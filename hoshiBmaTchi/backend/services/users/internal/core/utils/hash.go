package utils

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error){
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool{
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Inside internal/core/utils/ or strictly inside the handler if you prefer
func ParseUUID(id string) uuid.UUID {
    uid, _ := uuid.Parse(id)
    return uid
}
