package lnd

import "github.com/charmbracelet/bubbles/list"

type NodeData struct {
	NodeInfo        Node
	Channels        []Channel
	PendingChannels []PendingChannel
	Payments        []Payment
}

func (n NodeData) GetChannelsAsListItems(onlyOffline bool) []list.Item {
	var channelItems []list.Item
	for _, channel := range n.Channels {
		if (onlyOffline && !channel.Info.Active) || !onlyOffline {
			channelItems = append(channelItems, channel)
		}
	}

	return channelItems
}

func (n NodeData) GetPendingChannelsAsListItems() []list.Item {
	var pendingChannelItems []list.Item
	for _, pendingChannel := range n.PendingChannels {
		pendingChannelItems = append(pendingChannelItems, pendingChannel)
	}

	return pendingChannelItems
}

func (n NodeData) GetPaymentsAsListItems() []list.Item {
	var paymentItems []list.Item
	for _, payment := range n.Payments {
		paymentItems = append(paymentItems, payment)
	}

	reverseSlice(paymentItems)
	return paymentItems
}

func reverseSlice(slice []list.Item) {
	// Get the length of the slice
	length := len(slice)

	// Iterate through the slice up to its midpoint
	for i := 0; i < length/2; i++ {
		// Swap elements from the beginning and end of the slice
		slice[i], slice[length-i-1] = slice[length-i-1], slice[i]
	}
}
