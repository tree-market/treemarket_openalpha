package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	. "tree_service/types"
)

func pullLatestBitcart(invoice *SeedInvoice) *SeedInvoice {
	bitcart, err := getBitcartInvoice(invoice.BitcartID)
	if err != nil {
		fmt.Println("Error", err)
	}
	invoice.BitcartStatus = bitcart.Status
	if bitcart.Status == "complete" {
		if bitcart.Currency != "" {
			invoice.Currency = bitcart.Currency
		}
		invoice.CryptoReceived = bitcart.Amount
		invoice.IncomingTXID = bitcart.TXID[0]
	}

	return invoice
}

func getBitcartInvoice(id string) (*BitcartInvoice, error) {
	url := fmt.Sprintf("https://bitcart.tree.market/api/invoices/%s", id)
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	defer response.Body.Close()

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil, err
	}

	var invoice BitcartInvoice
	err = json.Unmarshal(body, &invoice)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	return &invoice, nil
}

func getBitcartInvoiceStatus(id string) (string, error) {

	url := fmt.Sprintf("https://bitcart.tree.market/api/invoices/%s", id)
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return "error", err
	}
	defer response.Body.Close()

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return "error", err
	}

	var invoice BitcartInvoice
	err = json.Unmarshal(body, &invoice)
	if err != nil {
		fmt.Println("Error:", err)
		return "error", err
	}
	return invoice.Status, nil
}

// TO-DO: RETURN INVOICE DATA ONLY
func checkInvoice(id string) {
	url := fmt.Sprintf("https://bitcart.tree.market/api/invoices/%s", id)
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer response.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	var invoice BitcartInvoice
	err = json.Unmarshal(body, &invoice)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	already_processed := alreadyProcessed(invoice.TXID[0])

	if already_processed { // if already processed skip it
		return
	}
	if invoice.Status == "complete" {
		var usdPaid float64
		for _, payment := range invoice.Payments {
			if invoice.Currency == payment.Symbol { //modified tolower
				rate, err := strconv.ParseFloat(payment.Rate, 64)
				if err != nil {
					fmt.Println("Error:", err)
				}
				cryptoPaid, _ := strconv.ParseFloat(invoice.Amount, 64)
				usdPaid = cryptoPaid * rate
				break
			}
		}
		//pricePair := invoice.Currency + "-USDT"

		//usdPaid := cryptoPaid * float64(prices[pricePair].Price)
		seedAmt := uint64(usdPaid / seedPrice)

		withdrawSeed(seedAmt)
		sendSeed(invoice.Address, seedAmt, 0, invoice.TXID[0])
		fmt.Println("Bitcart Invoice", invoice.Address, invoice.Status, invoice.Amount, usdPaid)

	}

	// Print the response body

}

func newBitcartInvoice(price float64) (*BitcartInvoice, error) {
	// Define the URL and access token
	url := "https://bitcart.tree.market/api/invoices"
	accessToken := "eI8wBGPsZNxkGOEaayDRtVGD2pkdqt0k_UoKVcdQNcA"

	priceStr := strconv.FormatFloat(price, 'f', 2, 64)

	invoice := BitcartInvoice{

		Address: "dero1qyfq8m3rju62tshju60zuc0ymrajwxqajkdh6pw888ejuv94jlfgjqq58px98",
		Price:   priceStr,
		Store:   "ggUMtJAehjHXYdoQrTTmaPfSyJzzAuIx",
	}

	// Marshal the invoice data into JSON format
	invoiceData, err := json.Marshal(invoice)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return nil, err
	}

	// Create a new HTTP POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(invoiceData))
	if err != nil {
		fmt.Println("Error creating HTTP request:", err)
		return nil, err
	}

	// Set request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Create an HTTP client and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending HTTP request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error:", resp.Status)
		return nil, err
	}

	// Read the response body
	var response BitcartInvoice
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Println("Error decoding JSON response:", err)
		return nil, err
	}

	// Print the response
	fmt.Println("Response:", response)
	return &response, nil
}
