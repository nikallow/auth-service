package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const verificationCodeLength = 6

func GenerateVerificationCode() (string, error) {
	code := make([]byte, verificationCodeLength)

	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", fmt.Errorf("generate verification code digit: %w", err)
		}

		code[i] = byte('0' + n.Int64())
	}

	return string(code), nil
}
