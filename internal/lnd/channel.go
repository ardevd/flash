package lnd

import (
	"github.com/btcsuite/btcd/btcutil"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/lightninglabs/lndclient"
)

// A wrapper around lndclient's ChannelInfo combined with a node Alias
type Channel struct {
	Info lndclient.ChannelInfo
	Alias string
}

// bubbletea interface function
func (c Channel) FilterValue() string {
	return c.Alias
}

// bubbletea interface function
func (c Channel) Title() string {
	return c.Alias
}

// bubbletea interface function
func (c Channel) Description() string {
	// Calculate node balance in percentage.
	localBalance := c.Info.LocalBalance.ToBTC()
	localBalancePercentage := localBalance / c.Info.Capacity.ToBTC()
	prog := progress.New(progress.WithoutPercentage())
	
	return satsToShortString(c.Info.LocalBalance.ToUnit(btcutil.AmountSatoshi)) + 
	" " + prog.ViewAs(localBalancePercentage) + " " + 
	satsToShortString(c.Info.RemoteBalance.ToUnit(btcutil.AmountSatoshi))
}