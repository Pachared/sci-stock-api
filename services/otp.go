package services

import (
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
