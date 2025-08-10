package cmd

import (
	"context"
	"encoding/hex"
	"os"

	"github.com/ardevd/flash/internal/credentials"
	"github.com/ardevd/flash/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/log"
	"github.com/lightninglabs/lndclient"
	"github.com/spf13/cobra"
)

var (
	// Used for flags.
	tlsCertFile      string
	adminMacaroon    string
	authFile         string
	encKey           string
	rpcServerAddress string
	rootCmd          = &cobra.Command{
		Use:   "flash",
		Short: "A TUI based Lightning node management tool",
		Long: `Flash is a TUI based Bitcoin Ligthing Network management tool
			written in Go.`,

		Run: func(cmd *cobra.Command, args []string) {
			runApp()
		},
	}
)

// Execute executes the root command.
func Execute() {
	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		os.Exit(1)
	}
}

func runApp() {

	logger := log.NewWithOptions(os.Stderr, log.Options{})
	styles := tui.GetDefaultStyles()
	if tlsCertFile != "" && adminMacaroon != "" {
		encryptionKey := credentials.EncryptCredentials(tlsCertFile, adminMacaroon)
		logger.Info("Encrypted credentials file 'auth.bin' saved.\nEncryption key:" +
			styles.Keyword(encryptionKey) + "\n\nauth.bin with the encryption key can now be used to connect to the node")
		return
	}

	if rpcServerAddress == "" {
		log.Fatal("No RPC hostname specified.")
	}

	var tlsData []byte
	var macData []byte
	if authFile != "" && encKey != "" {
		tlsData, macData = credentials.DecryptCredentials(encKey, authFile)
	} else {
		logger.Fatal("Auth file and encryption key required for node connection, alternatively generate them first with -a and -c")
	}

	// Create a new gRPC client using the provided credentials.
	config := lndclient.LndServicesConfig{
		LndAddress:        rpcServerAddress,
		Network:           lndclient.NetworkMainnet,
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

func init() {

	rootCmd.PersistentFlags().StringVarP(&tlsCertFile, "tlscert", "c", "", "TLS Certificate file")
	rootCmd.PersistentFlags().StringVarP(&adminMacaroon, "macaroon", "m", "", "Admin Macaroon")
	rootCmd.PersistentFlags().StringVarP(&authFile, "auth", "a", "", "Authentication file")
	rootCmd.PersistentFlags().StringVarP(&encKey, "enckey", "k", "", "Encryption key")
	rootCmd.PersistentFlags().StringVar(&rpcServerAddress, "host", "", "RPC hostname:port")

}
