package main

import (
	"fmt"
	"math"
	"time"
	"tree_service/utils"

	"github.com/deroproject/derohe/rpc"
	"github.com/deroproject/derohe/walletapi"
	"github.com/ybbus/jsonrpc"
)

const DEST_PORT = uint64(33333)

var INTEGRATED_MESSAGE string = ""

var expected_arguments = rpc.Arguments{
	{Name: rpc.RPC_DESTINATION_PORT, DataType: rpc.DataUint64, Value: DEST_PORT},
	{Name: rpc.RPC_COMMENT,
		DataType: rpc.DataString,
		Value:    INTEGRATED_MESSAGE},
	{Name: rpc.RPC_NEEDS_REPLYBACK_ADDRESS,
		DataType: rpc.DataUint64,
		Value:    uint64(0)},
}

var response = rpc.Arguments{
	{Name: rpc.RPC_DESTINATION_PORT, DataType: rpc.DataUint64, Value: uint64(0)},
	{Name: rpc.RPC_SOURCE_PORT, DataType: rpc.DataUint64, Value: DEST_PORT}, {Name: rpc.RPC_COMMENT, DataType: rpc.DataString, Value: "Successfully purchased SEED"},
}

var rpcClient jsonrpc.RPCClient

var addr *rpc.Address

func getUniqueIntegratedAddress(amount uint64, id string) string {
	uniqueAddress := addr.Clone()
	arguments := rpc.Arguments{
		{Name: rpc.RPC_DESTINATION_PORT,
			DataType: rpc.DataUint64,
			Value:    DEST_PORT,
		},
		{Name: rpc.RPC_COMMENT,
			DataType: rpc.DataString,
			Value:    id,
		},
		{Name: rpc.RPC_NEEDS_REPLYBACK_ADDRESS,
			DataType: rpc.DataUint64,
			Value:    uint64(0),
		},
		{
			Name:     rpc.RPC_VALUE_TRANSFER,
			DataType: rpc.DataUint64,
			Value:    uint64(amount),
		}}
	uniqueAddress.Arguments = arguments

	return uniqueAddress.String()
}

func listenForDeroPayments() {

	var err error

	for { // currently we traverse entire history

		time.Sleep(time.Second)

		var transfers rpc.Get_Transfers_Result
		err = rpcClient.CallFor(&transfers, "GetTransfers", rpc.Get_Transfers_Params{In: true, DestinationPort: DEST_PORT})
		if err != nil {
			logger.Error(err, "Could not obtain gettransfers from wallet")
			continue
		}

		for _, e := range transfers.Entries {
			if e.Coinbase || !e.Incoming { // skip coinbase or outgoing, self generated transactions
				continue
			}

			// check whether the entry has been processed before, if yes skip it
			already_processed := deroTXIDAlreadyProcessed(e.TXID)
			/* db.View(func(tx *bbolt.Tx) error {
				if b := tx.Bucket([]byte("SALE")); b != nil {
					if ok := b.Get([]byte(e.TXID)); ok != nil { // if existing in bucket
						already_processed = true
					}
				}
				return nil
			}) */

			if already_processed { // if already processed skip it
				continue
			}

			// check whether this service should handle the transfer
			if !e.Payload_RPC.Has(rpc.RPC_DESTINATION_PORT, rpc.DataUint64) ||
				DEST_PORT != e.Payload_RPC.Value(rpc.RPC_DESTINATION_PORT, rpc.DataUint64).(uint64) { // this service is expecting value to be specfic
				continue

			}
			comment := e.Payload_RPC.Value(rpc.RPC_COMMENT, rpc.DataString).(string)

			openInvoice := retrieveInvoice(comment)

			if openInvoice == nil {
				//TO-DO HANDLE NO EXISTING INVOICE FOR INCOMING DERO
				/* openInvoice = &SeedInvoice{
					DeroAddress:    destination_expected,
					Currency:       "dero",
					IncomingTXID:   e.TXID,
					CryptoReceived: e.Amount,
					Status:         "new",
				}
				openInvoice = addSeedInvoiceToDB(openInvoice) */
			}

			if openInvoice.SeedSent == openInvoice.Quantity {
				continue
			}

			logger.V(1).Info("tx should be processed", "txid", e.TXID)

			if !e.Payload_RPC.Has(rpc.RPC_REPLYBACK_ADDRESS, rpc.DataAddress) {
				logger.Error(nil, fmt.Sprintf("user has not give his address so we cannot replyback")) // this is an unexpected situation
				continue
			}

			destination_expected := e.Payload_RPC.Value(rpc.RPC_REPLYBACK_ADDRESS, rpc.DataAddress).(rpc.Address).String()
			addr, err := rpc.NewAddress(destination_expected)
			if err != nil {
				logger.Error(err, "err while while parsing incoming addr")
				continue
			}
			addr.Mainnet = false // convert addresses to testnet form, by default it's expected to be mainnnet
			destination_expected = addr.String()

			logger.V(1).Info("tx should be replied", "txid", e.TXID, "replyback_address", destination_expected)

			//destination_expected := e.Sender

			// value received is what we are expecting, so time for response
			response[0].Value = e.SourcePort // source port now becomes destination port, similar to TCP
			response[2].Value = fmt.Sprintf("Sucessfully purchased SEED.You sent %s at height %d", walletapi.FormatMoney(e.Amount), e.Height)

			seedPriceInDero := seedPrice / float64(utils.Prices["DERO-USDT"].Price)
			extraDero := e.Amount - uint64(float64(e.Amount)/seedPriceInDero)
			fmt.Println("new change!!", extraDero)

			seedAmtFloat := float64(e.Amount) * float64(utils.Prices["DERO-USDT"].Price) / seedPrice / 100000
			fmt.Println("seedAmtFloat", seedAmtFloat)
			change := uint64((seedAmtFloat - math.Floor(seedAmtFloat)) * 100000)
			fmt.Println("change", change)
			seedAmt := uint64(float64(e.Amount)*float64(utils.Prices["DERO-USDT"].Price)/seedPrice) / 100000
			//withdrawSeed(seedAmt)
			openInvoice.Quantity = seedAmt

			if openInvoice.DeroAddress == "" {
				openInvoice.DeroAddress = destination_expected
			}
			withdrawSeedPublicly(openInvoice)
			//sendSeedPrivately(openInvoice)

			//_, err :=  response.CheckPack(transaction.PAYLOAD0_LIMIT)) //  we only have 144 bytes for RPC

			// sender of ping now becomes destination
			//sendSeed(destination_expected, seedAmt, change, e.TXID)
			/* var result rpc.Transfer_Result
			tparams := rpc.Transfer_Params{Transfers: []rpc.Transfer{{Destination: destination_expected, SCID: crypto.HashHexToHash(SEED_SCID), Amount: seedAmt}, {Destination: destination_expected, Amount: change + 1, Payload_RPC: response}}}
			err = rpcClient.CallFor(&result, "Transfer", tparams)
			if err != nil {
				logger.Error(err, "err while transfer")
				continue
			}

			err = db.Update(func(tx *bbolt.Tx) error {
				b := tx.Bucket([]byte("SALE"))
				return b.Put([]byte(e.TXID), []byte("done"))
			})
			if err != nil {
				logger.Error(err, "err updating db")
			} else {
				logger.Info("ping replied successfully with pong ", "result", result)
			} */

		}

	}
}
