package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/ardevd/flash/internal/credentials"
	"github.com/ardevd/flash/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/lightninglabs/lndclient"

	"os"
)

func main() {
	// Arguments
	tlsCertFile := flag.String("c", "", "TLS Certificate file")
	adminMacaroon := flag.String("m", "", "Admin Macaroon")
	flag.Parse()
	if *tlsCertFile != ""  && *adminMacaroon != "" {
		encryptionKey := credentials.EncryptCredentials(*tlsCertFile, *adminMacaroon)
		log.Info("Encrypted credentials generated as 'auth.bin'.\n Encryption key:" + encryptionKey)
		return
	}
	// Replace these variables with your actual RPC credentials and endpoint.
	rpcServerAddress := "localhost:8888"
	macaroonDir := "macaroons/"
	tlsCertFilePath := "tls.crt"

	// Create a new gRPC client using the provided credentials.
	config := lndclient.LndServicesConfig{
		LndAddress:  rpcServerAddress,
		Network:     lndclient.NetworkMainnet,
		MacaroonDir: macaroonDir,
		TLSPath:     tlsCertFilePath,
	}
	client, err := lndclient.NewLndServices(&config)

	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	m := tui.InitLoading(client)
	p := tea.NewProgram(m)

	go func() {
		nodeData := tui.GetData(client, ctx)
		p.Send(tui.DataLoaded(nodeData))
	}()

	if _, err := p.Run(); err != nil {
		fmt.Println("error running program:", err)
		os.Exit(1)
	}
}
