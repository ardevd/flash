package lnd

import (
	"fmt"

	"github.com/lightninglabs/lndclient"
)

type Payment struct {
	Payment lndclient.Payment
}

func (p Payment) FilterValue() string {
	return p.Payment.Hash.String()
}

func (p Payment) Title() string {
	return fmt.Sprintf("%d sats", p.Payment.Amount.ToSatoshis())
}

func (p Payment) Description() string {
	return p.Payment.Status.State.String()
}