package util

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

const ErrUnauthorized = "unauthorized"
const ErrInternalError = "internal error"
const ErrInvalidToken = "invalid token"
const ErrProductNotFound = "product not found"

const CodeInvalidToken = 40

const UnlimitedSwipeProductCode = "SKU001"
const AccountVerifiedProductCode = "SKU002"

const GenderMale = "M"
const GenderFemale = "F"

const DateFormatYYYYMMDD = "2006-01-02"
const DateFormatYYYYMMDDTHHmmss = "2006-01-02T15:04:05"

var TimeNow = func() time.Time {
	return time.Now()
}

func ToDateTimeYYYYMMDD(dateString string) (dt time.Time, err error) {
	return time.Parse(DateFormatYYYYMMDD, dateString)
}

func ToDateTimeYYYYMMDDTHHmmss(dateString string) (dt time.Time, err error) {
	return time.Parse(DateFormatYYYYMMDDTHHmmss, dateString)
}

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
