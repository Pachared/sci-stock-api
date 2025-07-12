package services

import (
	"github.com/pquerna/otp/totp"
	"math/rand"
	"time"
)

func GenerateOTP(length int) string {
	rand.Seed(time.Now().UnixNano())
	digits := "0123456789"
	otp := ""
	for i := 0; i < length; i++ {
		otp += string(digits[rand.Intn(len(digits))])
	}
	return otp
}

func ValidateTOTP(code string, secret string) bool {
	return totp.Validate(code, secret)
}

func GenerateTwoFA(gmail string) (secret string, qrURL string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "SCI Stock",
		AccountName: gmail,
	})
	if err != nil {
		return "", "", err
	}
	return key.Secret(), key.URL(), nil
}