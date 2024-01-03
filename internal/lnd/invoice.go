package lnd

import (
	"context"
	"crypto/rand"
	"log"

	"github.com/lightninglabs/lndclient"
	"github.com/lightningnetwork/lnd/lnrpc/invoicesrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/lightningnetwork/lnd/lnwire"
)

func GeneratePaymentInvoice(service *lndclient.GrpcLndServices, ctx context.Context, memo string,
	satsAmount uint64, expiry int64) (lntypes.Hash, string, error) {

	preimage, hash, err := generateRandomPreimageAndHash()
	if err != nil {
		log.Fatal(err)
	}

	invoice := invoicesrpc.AddInvoiceData{
		Memo:     memo,
		Value:    lnwire.MilliSatoshi(satsAmount * 1000),
		Expiry:   int64(expiry),
		Hash:     &hash,
		Preimage: preimage,
	}
	return service.Client.AddInvoice(ctx, &invoice)
}

func generateRandomPreimageAndHash() (*lntypes.Preimage,
	lntypes.Hash, error) {
	var (
		paymentPreimage *lntypes.Preimage
		paymentHash     lntypes.Hash
	)

	paymentPreimage = &lntypes.Preimage{}
	if _, err := rand.Read(paymentPreimage[:]); err != nil {
		return nil, lntypes.Hash{}, err
	}
	paymentHash = paymentPreimage.Hash()

	return paymentPreimage, paymentHash, nil
}
