package lnd

import (
	"math"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/lightninglabs/lndclient"
)

// A wrapper around lndclient's ChannelInfo combined with a node Alias
type Channel struct {
	Info  lndclient.ChannelInfo
	Alias string
}

// bubbletea interface function
func (c Channel) FilterValue() string {
	return c.Alias
}

// bubbletea interface function
func (c Channel) Title() string {
	titleString := c.Alias
	if len(c.Info.PendingHtlcs) > 0 {
		titleString += "*"
	}
	if !c.Info.Active {
		titleString += " (OFFLINE)"
	}

	return titleString
}

func (c Channel) UptimePct() int {
	uptimeSeconds := c.Info.Uptime.Seconds()
	totalTimeSeconds := c.Info.LifeTime.Seconds()

	if totalTimeSeconds == 0 {
		return 0 // Handle the case where totalTime is 0 to avoid division by zero
	}

	percentage := (uptimeSeconds / totalTimeSeconds) * 100
	return int(math.Ceil(percentage))
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
