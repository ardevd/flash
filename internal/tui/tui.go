package tui

import (
	"context"
	"log"

	"github.com/ardevd/flash/internal/lnd"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lightninglabs/lndclient"
	"github.com/lightningnetwork/lnd/routing/route"
)

var windowSizeMsg tea.WindowSizeMsg

/* Model management */
var Models []tea.Model

// Message types
type DataLoaded lnd.NodeData

// Payments
type paymentSettled struct{}
type paymentExpired struct{}
type paymentCreated struct{}
type paymentError struct{}

// Channel
type updateChannelPolicy struct{}

func updateChannelPolicyMsg() tea.Msg {
	return updateChannelPolicy{}
}

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
	nodeData.Channels = getChannelListItems(service, ctx)

	// Load Pending channels
	nodeData.PendingChannels = getPendingChannels(service, ctx)

	// Load node data
	nodeData.NodeInfo = lnd.GetDataFromAPI(service, ctx)

	return nodeData
}

// Get list of pending channels
func getPendingChannels(service *lndclient.GrpcLndServices, ctx context.Context) []lnd.PendingChannel {
	var pendingChannels []lnd.PendingChannel
	channels, err := service.Client.PendingChannels(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Force close channels
	for _, fc := range channels.PendingForceClose {
		remotePeerAlias := getNodeAliasFromPubKey(service, ctx, fc.PubKeyBytes)

		pendingChannel := lnd.PendingChannel{
			Capacity:            fc.Capacity,
			LocalBalance:        fc.LocalBalance,
			RecoveredBalance:    fc.RecoveredBalance,
			LimboBalance:        fc.LimboBalance,
			BlocksUntilMaturity: fc.BlocksUntilMaturity,
			Type:                lnd.ForceClosure,
			Alias:               remotePeerAlias,
		}
		pendingChannels = append(pendingChannels, pendingChannel)
	}

	// Cooperative closing channels
	for _, fc := range channels.WaitingClose {
		remotePeerAlias := getNodeAliasFromPubKey(service, ctx, fc.PubKeyBytes)

		pendingChannel := lnd.PendingChannel{
			Capacity:     fc.Capacity,
			LocalBalance: fc.LocalBalance,
			Type:         lnd.CooperativeClosure,
			Alias:        remotePeerAlias,
		}
		pendingChannels = append(pendingChannels, pendingChannel)
	}

	// Pending channel opens
	for _, fc := range channels.PendingOpen {
		remotePeerAlias := getNodeAliasFromPubKey(service, ctx, fc.PubKeyBytes)

		pendingChannel := lnd.PendingChannel{
			Capacity:     fc.Capacity,
			LocalBalance: fc.LocalBalance,
			Type:         lnd.CooperativeClosure,
			Alias:        remotePeerAlias,
		}
		pendingChannels = append(pendingChannels, pendingChannel)
	}

	return pendingChannels
}

func getNodeAliasFromPubKey(service *lndclient.GrpcLndServices, ctx context.Context, pubKey route.Vertex) string {
	node, err := service.Client.GetNodeInfo(ctx, pubKey, true)
	if err != nil {
		// TODO: add logging
		return ""
	}

	return node.Alias
}

func getChannelListItems(service *lndclient.GrpcLndServices, ctx context.Context) []lnd.Channel {
	var channels []lnd.Channel
	infos, err := service.Client.ListChannels(ctx, false, false)
	if err != nil {
		log.Fatal(err)
	}

	for _, chanInfo := range infos {
		remotePeerAlias := getNodeAliasFromPubKey(service, ctx, chanInfo.PubKeyBytes)

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
	OPTION_MESSAGE_SIGN    = "sign"
	OPTION_MESSAGE_VERIFY  = "verify"
	OPTION_CHANNEL_OPEN    = "open"
	OPTION_CONNECT_TO_PEER = "connect"
)
