package utils

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"
)

func GenerateTransactionId() (string, error) {
	// Use milliseconds-based counter to avoid collisions during rapid seeding.
	// The old "txn_%04d" (0-9999) collided when creating many transactions at once.
	now := time.Now().UnixMilli()

	// Random suffix between 0-9999 for extra entropy.
	max := big.NewInt(10000)
	randomNumber, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", errors.New("failed to generate transaction id")
	}

	// Format: txn_{millis}_{random4digits} — fits varchar(36).
	return fmt.Sprintf("txn_%d_%04d", now, randomNumber.Int64()), nil
}

func GenerateTransactionItemId() (string, error) {
	now := time.Now().UnixNano()
	max := big.NewInt(10000)
	randomNumber, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", errors.New("failed to generate transaction item id")
	}

	return fmt.Sprintf("txni_%d_%04d", now, randomNumber.Int64()), nil
}
