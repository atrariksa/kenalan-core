package util

import "golang.org/x/crypto/bcrypt"

func HashPassword(input string) string {
	password := []byte(input)

	// Hashing the password with the default cost of 10
	hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	// Comparing the password with the hash
	err = bcrypt.CompareHashAndPassword(hashedPassword, password)
	if err != nil {
		panic(err)
	}
	return string(hashedPassword)
}

func ValidatePassword(givenPlainTextPassword string, storedHashedPassword string) error {
	password := []byte(givenPlainTextPassword)
	hashedPassword := []byte(storedHashedPassword)
	// Comparing the password with the hash
	return bcrypt.CompareHashAndPassword(hashedPassword, password)
}
