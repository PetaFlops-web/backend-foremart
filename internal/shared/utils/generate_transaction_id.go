package utils

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
)

func GenerateTransactionId() (string, error) {
	max := big.NewInt(10000)
	randomNumber, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", errors.New("failed to generate transaction id")
	}

	return fmt.Sprintf("txn_%04d", randomNumber.Int64()), nil
}

func GenerateTransactionItemId() (string, error) {
	max := big.NewInt(10000)
	randomNumber, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", errors.New("failed to generate transaction item id")
	}

	return fmt.Sprintf("txni_%04d", randomNumber.Int64()), nil
}
