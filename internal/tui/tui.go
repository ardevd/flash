package tui

import (
	"context"
	"log"

	"github.com/ardevd/flash/internal/lnd"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lightninglabs/lndclient"
)

var windowSizeMsg tea.WindowSizeMsg

/* Model management */
var Models []tea.Model

// Message types
type DataLoaded lnd.NodeData
type paymentSettled struct{}
type paymentExpired struct{}
type paymentCreated struct{}

func GetData(service *lndclient.GrpcLndServices, ctx context.Context) lnd.NodeData {
	var nodeData lnd.NodeData

	request := lndclient.ListPaymentsRequest{MaxPayments: 10, Reversed: true, IncludeIncomplete: false}
	payments, err := service.Client.ListPayments(ctx, request)
	if err != nil {
		log.Fatal(err)
	}

	// Load Payments
	var paymentsSlice []lnd.Payment
	for _, payment := range payments.Payments {
		paymentsSlice = append(paymentsSlice, lnd.Payment{Payment: payment})
	}

	nodeData.Payments = paymentsSlice

	// Load Channels
	nodeData.Channels = GetChannelListItems(service, ctx)

	// Load node data
	nodeData.NodeInfo = lnd.GetDataFromAPI(service, ctx)

	return nodeData
}

func GetChannelListItems(service *lndclient.GrpcLndServices, ctx context.Context) []lnd.Channel {
	var channels []lnd.Channel
	infos, err := service.Client.ListChannels(ctx, false, false)
	if err != nil {
		log.Fatal(err)
	}

	for _, chanInfo := range infos {
		remotePeerAlias := ""
		channelNode, err := service.Client.GetNodeInfo(ctx, chanInfo.PubKeyBytes, true)
		if err != nil {
			// TODO: add logging
		} else {
			remotePeerAlias = channelNode.Alias
		}

		channels = append(channels, lnd.Channel{Info: chanInfo, Alias: remotePeerAlias})
	}

	return channels
}

func Init(service *lndclient.GrpcLndServices) []tea.Model {
	progress := InitLoading(service)
	Models = []tea.Model{progress}
	return Models
}

const (
	OPTION_PAYMENT_RECEIVE = "receive"
	OPTION_PAYMENT_SEND    = "send"
)
