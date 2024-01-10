package main

import (
	"context"
	"encoding/hex"
	"flag"

	"github.com/ardevd/flash/internal/credentials"
	"github.com/ardevd/flash/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/lightninglabs/lndclient"

	"os"
)

func main() {
	logger := log.NewWithOptions(os.Stderr, log.Options{})
	styles := tui.GetDefaultStyles()
	// Arguments
	tlsCertFile := flag.String("c", "", "TLS Certificate file")
	adminMacaroon := flag.String("m", "", "Admin Macaroon")
	authFile := flag.String("a", "", "Authentication file")
	encKey := flag.String("k", "", "Encryption key")
	rpcServerAddress := flag.String("h", "", "RPC hostname:port")
	flag.Parse()

	if *tlsCertFile != "" && *adminMacaroon != "" {
		encryptionKey := credentials.EncryptCredentials(*tlsCertFile, *adminMacaroon)
		log.Info("Encrypted credentials file 'auth.bin' saved.\nEncryption key:" +
			styles.Keyword(encryptionKey) + "\n\nauth.bin with the encryption key can now be used to connect to the node")
		return
	}

	if *rpcServerAddress == "" {
		log.Fatal("No RPC hostname specified.")
	}

	var tlsData []byte
	var macData []byte
	if *authFile != "" && *encKey != "" {
		tlsData, macData = credentials.DecryptCredentials(*encKey, *authFile)
	} else {
		logger.Fatal("Auth file and encryption key required for node connection, alternatively generate them first with -a and -c")
	}

	// Create a new gRPC client using the provided credentials.
	config := lndclient.LndServicesConfig{
		LndAddress:        *rpcServerAddress,
		Network:           lndclient.NetworkRegtest,
		CustomMacaroonHex: hex.EncodeToString(macData),
		TLSData:           string(tlsData),
	}
	client, err := lndclient.NewLndServices(&config)

	if err != nil {
		logger.Fatal(err)
	}

	ctx := context.Background()

	m := tui.InitLoading(client)
	p := tea.NewProgram(m)

	go func() {
		nodeData := tui.GetData(client, ctx)
		p.Send(tui.DataLoaded(nodeData))
	}()

	if _, err := p.Run(); err != nil {
		logger.Fatal("error running program:", err)
		os.Exit(1)
	}
}
