package lnd

import (
	"context"
	"log"

	"github.com/lightninglabs/lndclient"
)

type Node struct {
	Alias string
	PubKey string
	Version string
	ChannelBalance string
	TotalCapacity string
	OnChainBalance string
}

func GetFromApi(service *lndclient.GrpcLndServices, ctx context.Context) Node {
	info, err := service.Client.GetInfo(ctx)
	if err != nil {
		log.Fatal(err)
	}

	nodeInfo, err := service.Client.GetNodeInfo(ctx, service.NodePubkey, false)
	if err != nil {
		log.Fatal(err)
	}
	
	channelBalance, err := service.Client.ChannelBalance(ctx)
	if err != nil {
		log.Fatal(err)
	}

	walletBalance, err := service.Client.WalletBalance(ctx)
	if err != nil {
		log.Fatal(err)
	}

	return Node{
		Alias: nodeInfo.Alias,
		PubKey: nodeInfo.PubKey.String(),
		Version: info.Version,
		ChannelBalance: channelBalance.Balance.String(),
		TotalCapacity: nodeInfo.TotalCapacity.String(),
		OnChainBalance: walletBalance.Confirmed.String(),
	}
}