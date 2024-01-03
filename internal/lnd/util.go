package lnd

import "fmt"

func satsToShortString(sats float64) string {
	millionSats := (sats / 1000000.0)
	return fmt.Sprintf("%.1fm", millionSats)
}