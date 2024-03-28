package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"

	. "tree_service/types"
	"tree_service/utils"

	"tree_service/eth"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func newTreeInvoice(w http.ResponseWriter, r *http.Request) {
	var invoice SeedInvoice
	err := json.NewDecoder(r.Body).Decode(&invoice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := uuid.New().String()
	invoice.SeedID = id
	invoice.Status = "open"
	deroUSDT := utils.Prices["DERO-USDT"].Price
	deroRate := strconv.FormatFloat(float64(deroUSDT), 'f', 2, 64)

	//need to get dero price here

	usdPrice := seedPrice * float64(invoice.Quantity)

	deroPrice := usdPrice / float64(deroUSDT)
	deroPrice = math.Ceil(deroPrice*math.Pow10(5)) / math.Pow10(5)
	deroAmount := strconv.FormatFloat(deroPrice, 'f', 5, 64)
	integratedAddress := getUniqueIntegratedAddress(uint64(deroPrice*100000), invoice.SeedID)

	invoice.Integrated = integratedAddress
	deroPayment := BitcartPayment{Rate: deroRate, Currency: "dero", Amount: deroAmount, Address: integratedAddress}

	bitcartInvoice, err := newBitcartInvoice(usdPrice)
	if err != nil {
		fmt.Println("Error", err)
	}
	invoice.Payments = append(bitcartInvoice.Payments, deroPayment)

	ethID := eth.GetNextEthIDB(invoice.Quantity)
	for i := range invoice.Payments {
		if invoice.Payments[i].Currency == "eth" || invoice.Payments[i].Currency == "matic" || invoice.Payments[i].Currency == "trx" {

			amount, err := strconv.ParseFloat(invoice.Payments[i].Amount, 64)
			if err != nil {
				fmt.Println("Error converting string to float:", err)
				return
			}
			amount = math.Round(amount*math.Pow10(3)) / math.Pow10(3)
			newAmount := strconv.FormatFloat(amount, 'f', 3, 64)
			invoice.Payments[i].Amount = newAmount + ethID
		}
	}
	ethid, err := strconv.ParseInt(ethID, 10, 64)
	if err != nil {
		fmt.Println("error converting eth id", err)
	}
	invoice.EthID = int(ethid)
	invoice.Blocks.ETH = eth.GetLatestBlock("eth")
	invoice.Blocks.MATIC = eth.GetLatestBlock("matic")
	invoice.Timeout = bitcartInvoice.Timeout
	invoice.Created = bitcartInvoice.Created
	invoice.BitcartID = bitcartInvoice.ID
	invoice.BitcartStatus = bitcartInvoice.Status
	invoice = addSeedInvoiceToDB(invoice)
	// Marshal the invoice object into JSON
	invoiceJSON, err := json.Marshal(invoice)
	if err != nil {
		// Handle the error, perhaps by returning an error response
		http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON response to the client
	w.WriteHeader(http.StatusOK) // Optional: Set the HTTP status code
	w.Write(invoiceJSON)

}

func checkInvoiceStatus(w http.ResponseWriter, r *http.Request) {
	//here we can check bitcart status so we need bitcartid
	//or we can check dero status which should be in db
	//or
	vars := mux.Vars(r)
	invoiceID := vars["id"]
	invoice := retrieveInvoice(invoiceID)

	//CASE 1: COMPLETE. DO NOTHING

	//CASE 2: OPEN. CHECK BITCART
	if invoice.Status == "open" {
		invoice = pullLatestBitcart(invoice)
		//if bitstatus is complete
		//send seed
		ethPaid := eth.SearchEthTransfers(invoice)
		if ethPaid {
			invoice.BitcartStatus = "complete"
		}
		if invoice.BitcartStatus == "complete" {
			withdrawSeedPublicly(invoice)
			eth.ClearEthID(invoice.Quantity, invoice.EthID)
		}
		if invoice.BitcartStatus == "expired" {
			invoice.Status = "expired"
		}

	}

	//CASE 3: UNREDEEMED. CHECK CONTRACT FOR REDEMPTION

	invoiceJSON, err := json.Marshal(invoice)
	if err != nil {
		// Handle the error, perhaps by returning an error response
		http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON response to the client
	w.WriteHeader(http.StatusOK) // Optional: Set the HTTP status code
	w.Write(invoiceJSON)

}
