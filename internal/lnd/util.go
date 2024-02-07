package lnd

import (
	"fmt"
	"strings"
)

func satsToShortString(sats float64) string {
	millionSats := (sats / 1000000.0)
	return fmt.Sprintf("%.1fm", millionSats)
}

func SantizeBoltInvoice(invoice string) string {
	cleanedInvoice := strings.Replace(invoice, "lightning:", "", -1)
	return cleanedInvoice
}
