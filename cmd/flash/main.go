package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/ardevd/flash/internal/credentials"
	"github.com/ardevd/flash/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lightninglabs/lndclient"

	"os"
)

func main() {
	// Arguments
	tlsCertFile := flag.String("tlsCert", "", "TLS Certificate file")
	flag.Parse()
	if *tlsCertFile != "" {
		fmt.Println(credentials.GenerateKey())
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
