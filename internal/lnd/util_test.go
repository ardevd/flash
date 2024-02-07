package lnd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSantizeBoltInvoice(t *testing.T) {
	invoiceStr := "lightning:lnbc10u1pju83nypp5ffwgtf2hheeyxt69u5pxku8ww83nsvy5n8jenl239cx8xq2fq3nqdp8fe5kxetgv9eksgzyv4cx7umfwssyjmnkda5kxegcqzysxqr8pqsp5z6dfwhvzkjwh8tzggnh82zhjk2mx3eweysaj93eaeuxs2mevz55q9qyyssqtpv4pq5enzaqv5d7yhftzpzwtxlmtq7wacv5jz4she9lphe99fazqehzff73k7hh64stmnsk4dvhcldazxpjaz9l6fwu5al0w9nq5wspvncy4m"

	cleanedInvoiceStr := SantizeBoltInvoice(invoiceStr)

	assert.True(t, strings.HasPrefix(cleanedInvoiceStr, "lnbc10"))
}
