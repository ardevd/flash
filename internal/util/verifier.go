package util

import (
	"errors"
	"strconv"
)
// Indicates whether the provided string value
// is a valid amount
func IsAmount(s string) error {
	amount, err := strconv.Atoi(s)
	if err != nil {
		return errors.New("invalid amount")
	}

	if amount <= 0 {
		return errors.New("amount too low")
	}

	return nil
}

// Indicates whether the provided string value
// is a valid payment memo
func IsMemo(s string) error { return nil }
