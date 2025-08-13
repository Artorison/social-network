package sessions

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)

var pepper string

func MustPepperKey(key string) {
	if key == "" {
		log.Fatal("empty secret key in sessions.MustPepperKey")
	}
	pepper = key
}

func GenerateHashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pepper+password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func CheckHashPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pepper+password))
	return err == nil
}
