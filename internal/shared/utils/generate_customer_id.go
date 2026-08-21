package utils

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
)

func GenerateCustomerId() (string, error) {
	max := big.NewInt(10000)
	randomNumber, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", errors.New("failed to generate customer id")
	}

	return fmt.Sprintf("cust_%04d", randomNumber.Int64()), nil
}
