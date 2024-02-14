package lnd

import (
	"math"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/lightninglabs/lndclient"
)

// ChannelType represents the type of Lightning network channel.
type PendingChannelType int

const (
	// PendingOpen indicates a pending channel opening.
	PendingOpen PendingChannelType = iota
	// CooperativeClosure indicates a cooperative channel closure.
	CooperativeClosure
	// ForceClosure indicates a forceful channel closure.
	ForceClosure
)

type PendingChannel struct {
	Capacity            btcutil.Amount
	LocalBalance        btcutil.Amount
	RecoveredBalance    btcutil.Amount
	LimboBalance        btcutil.Amount
	BlocksUntilMaturity int32
	Type                PendingChannelType
	Alias               string
}

func (c PendingChannel) FilterValue() string {
	return c.Alias
}

func (c PendingChannel) Title() string {
	return c.Alias
}

func (c PendingChannel) Description() string {
	return ""
}
