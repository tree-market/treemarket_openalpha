package main

import (
	"fmt"
	"time"
	. "tree_service/types"
	"tree_service/utils"

	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/rpc"
)

// this one sends direct from contract
func withdrawSeedPublicly(invoice *SeedInvoice) {
	if invoice.DeroAddress == "" {
		if invoice.Email == "" {
			invoice.Status = "missing_address"

		} else {
			invoice.Status = "unredeemed"
			password, err := utils.GenerateRandomPassword(18)
			hash := utils.CalculateSHA256(password)
			invoice.Password = password
			if err != nil {
				fmt.Println("Error generating password: ", err)
			}

			//withdraw and store for user
			//or keep in contract with secret code
			var result rpc.Transfer_Result
			tparams := rpc.Transfer_Params{Ringsize: 2, SC_ID: DISTRIBUTOR_SCID, SC_RPC: []rpc.Argument{{Name: "entrypoint", DataType: "S", Value: "ReserveTokens"}, {Name: "amount", DataType: "U", Value: invoice.Quantity}, {Name: "token", DataType: "S", Value: SEED_SCID}, {Name: "hash", DataType: "S", Value: hash}}}

			err = rpcClient.CallFor(&result, "Transfer", tparams)
			if err != nil {
				logger.Error(err, "err while transfer")
				return
			}

			//withdraw on behalf of user
			//get txid
			//if successful,

			invoice.SeedOutTXID = result.TXID
			invoice.SeedSent = invoice.Quantity
		}

	} else {
		var result rpc.Transfer_Result
		tparams := rpc.Transfer_Params{Ringsize: 2, SC_ID: DISTRIBUTOR_SCID, SC_RPC: []rpc.Argument{{Name: "entrypoint", DataType: "S", Value: "WithdrawPublic"}, {Name: "amount", DataType: "U", Value: invoice.Quantity}, {Name: "token", DataType: "S", Value: SEED_SCID}, {Name: "recipient", DataType: "S", Value: invoice.DeroAddress}}}

		err := rpcClient.CallFor(&result, "Transfer", tparams)
		if err != nil {
			logger.Error(err, "err while transfer")
			return
		}
		//withdraw on behalf of user
		//get txid
		//if successful,
		invoice.Status = "complete"
		invoice.SeedOutTXID = result.TXID
		invoice.SeedSent = invoice.Quantity

		// else invoice.Status = "failed"
	}
	updateInvoice(invoice)
}

// this one sends wallet to wallet
func sendSeedPrivately(invoice *SeedInvoice) {
	time.Sleep(time.Second * 60)
	if invoice.DeroAddress == "" {
		if invoice.Email == "" {
			invoice.Status = "missing_address"
		} else {
			invoice.Status = "unredeemed"
			//withdraw and store for user
			//or keep in contract with secret code
		}

	} else {
		var result rpc.Transfer_Result
		tparams := rpc.Transfer_Params{Transfers: []rpc.Transfer{{Destination: invoice.DeroAddress, SCID: crypto.HashHexToHash(SEED_SCID), Amount: invoice.Quantity}, {Destination: invoice.DeroAddress, Amount: 1, Payload_RPC: response}}}
		err := rpcClient.CallFor(&result, "Transfer", tparams)
		if err != nil {
			logger.Error(err, "err while transfer")
			return
		}
		fmt.Println("seed-out-tx", result)

		//if successful,
		invoice.Status = "complete"
		invoice.SeedOutTXID = result.TXID
		invoice.SeedSent = invoice.Quantity
		// else invoice.Status = "failed"
	}
	updateInvoice(invoice)
}

// TO-DO
func checkSeedRemainingInContract() {

}

// Function to update the seed price at midnight UTC
func updateSeedPriceTicker() {
	// Update the seed price immediately upon starting the server
	updateSeedPrice()
	// Calculate the duration until the next midnight UTC
	now := time.Now().UTC()
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	durationUntilMidnight := nextMidnight.Sub(now)

	// 24hr ticker
	ticker := time.NewTicker(60 * 60 * 24)

	time.Sleep(durationUntilMidnight)

	// Update the seed price at first midnight
	updateSeedPrice()

	// Schedule the ticker to update the seed price at midnight UTC
	for range ticker.C {
		updateSeedPrice()
	}
} //TO-DO: MAKE SURE THIS ACTUALLY HITS MIDNIGHT

// Function to update the seed price based on the current date
func updateSeedPrice() {
	currentTime := time.Now().Format(time.RFC3339)
	seedPrice, err := calculateSeedPrice(currentTime)
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Printf("Updated Seed Price at %s (UTC): %.2f\n", currentTime, seedPrice)
}

// Function to calculate seed price based on date (similar to previous example)
func calculateSeedPrice(dateString string) (float64, error) {
	// Parse the date string into a time.Time object
	date, err := time.Parse(time.RFC3339, dateString)
	if err != nil {
		return 0, err
	}

	// Define the date ranges and corresponding seed prices
	// Adjust the date ranges and prices as needed
	dateRanges := []struct {
		startDate time.Time
		endDate   time.Time
		seedPrice float64
	}{
		{time.Date(2024, time.March, 7, 0, 0, 0, 0, time.UTC), time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC), 1.99},
		{time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC), time.Date(2024, time.April, 15, 0, 0, 0, 0, time.UTC), 2.99},
		{time.Date(2024, time.April, 15, 0, 0, 0, 0, time.UTC), time.Date(2024, time.April, 21, 0, 0, 0, 0, time.UTC), 3.99},
	}

	// Iterate through the date ranges to find the corresponding seed price
	for _, rangeData := range dateRanges {
		if date.After(rangeData.startDate) && date.Before(rangeData.endDate) {
			return rangeData.seedPrice, nil
		}
	}

	// Return an error if the date is outside the defined ranges
	return 0, fmt.Errorf("seed price unavailable for the given date")
}

/* func withdrawSeed(amount uint64) {
	fmt.Println("withdrawing ", amount, " seed")

	var result rpc.Transfer_Result
	tparams := rpc.Transfer_Params{Ringsize: 2, SC_ID: DISTRIBUTOR_SCID, SC_RPC: []rpc.Argument{{Name: "entrypoint", DataType: "S", Value: "Withdraw"}, {Name: "amount", DataType: "U", Value: amount}, {Name: "token", DataType: "S", Value: SEED_SCID}}, Transfers: []rpc.Transfer{{Destination: "deto1qyre7td6x9r88y4cavdgpv6k7lvx6j39lfsx420hpvh3ydpcrtxrxqg8v8e3z", SCID: crypto.HashHexToHash(BOT_SCID), Burn: 1}}}

	err := rpcClient.CallFor(&result, "Transfer", tparams)
	if err != nil {
		logger.Error(err, "err while transfer")
		return
	}

} */
