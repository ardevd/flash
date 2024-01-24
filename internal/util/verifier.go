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

func IsMessage(s string) error {
	if len(s) >0 {
		return nil
	} else {
		return errors.New("no message provided")
	}
}

func IsSignature(s string) error {
	if len(s) >0 {
		return nil
	} else {
		return errors.New("no signature provided")
	}
}

// Indicates whether the provided string value
// is a valid payment memo
func IsMemo(s string) error { return nil }
