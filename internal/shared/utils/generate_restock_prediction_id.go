package utils

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
)

func GenerateRestockPredictionId() (string, error) {
	max := big.NewInt(10000)
	randomNumber, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", errors.New("failed to generate restock prediction id")
	}

	return fmt.Sprintf("restock_%04d", randomNumber.Int64()), nil
}
