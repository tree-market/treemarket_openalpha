package eth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	. "tree_service/types"
	"tree_service/utils"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/joho/godotenv"
)

var counts = make(map[int]map[int]int)
var countsB = make(map[float64][]bool)
var mutex sync.Mutex

const MATIC_ADDRESS = "0x5BBD9Ab48C6F01D9A3729e92115e444B4E5785EC"
const ETH_ADDRESS = "0x6a25E3d4C99E026023E00B69246A6983E312b07C"
const TRX_ADDRESS = "TAbuQMqepEFtfMDjmtvnsRRGkdA7zBLYnG"

// need access to eth node
var ethEndpoint = "https://eth-mainnet.g.alchemy.com/v2/"

var polygonEndpoint = "https://polygon-mainnet.g.alchemy.com/v2/"

var client *rpc.Client
var err error

func ConnectToEth() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	ethEndpoint += os.Getenv("ETH_API")
	polygonEndpoint += os.Getenv("MATIC_API")
	client, err = rpc.Dial(ethEndpoint)
	if err != nil {
		fmt.Println("Error connecting to eth", err)
		return
	}
	defer client.Close()

}

func getAllTransfers(fromBlock string, endpoint string, toAddress string) []Transfer {
	fmt.Println(fromBlock, toAddress, endpoint)
	payload := map[string]interface{}{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "alchemy_getAssetTransfers",
		"params": []map[string]interface{}{
			{
				"fromBlock":        fromBlock,
				"toBlock":          "latest",
				"toAddress":        toAddress,
				"withMetadata":     false,
				"excludeZeroValue": true,
				"maxCount":         "0x3e8",
				"category":         []string{"external", "erc20"},
			},
		},
	}

	// Convert payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Fatal("Error marshaling JSON payload:", err)
	}

	// Create a new HTTP POST request
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Fatal("Error creating HTTP request:", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create an HTTP client and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error sending HTTP request:", err)
	}
	defer resp.Body.Close()

	// Read the response body
	var response Response
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Fatal("Error decoding JSON response:", err)
	}
	fmt.Println("reponse", response)

	for _, transfer := range response.Result.Transfers {
		fmt.Printf("From: %s, To: %s, Quantity: %f\n", transfer.From, transfer.To, transfer.Quantity)
	}
	return response.Result.Transfers
	// return array of values since invoice initiation
}

func SearchEthTransfers(invoice *SeedInvoice) bool {
	invoiceTime := utils.GetUnixTime(invoice.Created)

	chains := []string{"eth", "matic"}
	for _, chain := range chains {
		var fromBlock string
		var endpoint string
		var toAddress string
		if chain == "matic" {
			fromBlock = invoice.Blocks.MATIC
			endpoint = polygonEndpoint
			toAddress = MATIC_ADDRESS
		}
		if chain == "eth" {
			fromBlock = invoice.Blocks.ETH
			endpoint = ethEndpoint
			toAddress = ETH_ADDRESS
		}
		latestTransfers := getAllTransfers(fromBlock, endpoint, toAddress) //or notupdate to have variable currency in getall
		for _, transfer := range latestTransfers {
			fmt.Println(transfer)
			fmt.Println("received: ", transfer.Quantity, strings.ToLower(transfer.Currency))
			for _, payment := range invoice.Payments {
				amount, err := strconv.ParseFloat(payment.Amount, 64)
				fmt.Println("want: ", amount, strings.ToLower(payment.Currency))
				if err != nil {
					fmt.Println("Error", err)
				}

				if transfer.Quantity == amount && strings.ToLower(transfer.Currency) == strings.ToLower(payment.Currency) { //modified tolower
					fmt.Println("we did it reddit")

					blockTime := getBlockTimeStamp(chain, transfer.Block)
					fmt.Println(blockTime, invoiceTime, invoiceTime+900)
					if blockTime > invoiceTime && blockTime < invoiceTime+900 {
						invoice.IncomingTXID = transfer.TXID
						invoice.Currency = transfer.Currency
						invoice.CryptoReceived = fmt.Sprint(amount)
						return true
					}

				}
			}

		}

	}

	return searchTron(invoice)
}

func GetNextEthID(quantity uint64) string {
	currentInterval := int(time.Now().Unix() / 900)
	fmt.Println("time", time.Now().Unix())
	mutex.Lock()
	defer mutex.Unlock()

	if counts[currentInterval] == nil {
		counts[currentInterval] = make(map[int]int)
	}

	counts[currentInterval][int(quantity)]++
	fmt.Println(counts)

	return fmt.Sprintf("%02d", counts[currentInterval][int(quantity)])

}

func GetNextEthIDB(quantity float64) string {
	mutex.Lock()
	defer mutex.Unlock()
	if _, ok := countsB[quantity]; !ok {
		countsB[quantity] = make([]bool, 100)
	}
	for i := 0; i < 100; i++ {
		if !countsB[quantity][i] {
			fmt.Print("first false found at", i)
			countsB[quantity][i] = true
			go startEthIDTimer(quantity, i)
			return fmt.Sprintf("%02d", i)

		}
	}
	return "99"
}

func startEthIDTimer(quantity float64, id int) {
	time.Sleep(time.Second * 900)
	countsB[quantity][id] = false
}

func ClearEthID(quantity float64, id int) {
	if _, ok := countsB[quantity]; !ok {
		return
	}
	countsB[quantity][id] = false
}

func getBlockTimeStamp(chain string, blockNumber string) int64 {
	endpoint := ""
	if chain == "eth" {
		endpoint = ethEndpoint
	}
	if chain == "matic" {
		endpoint = polygonEndpoint
	}

	payload := map[string]interface{}{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "eth_getBlockByNumber",
		"params":  []interface{}{blockNumber, true},
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("error marshal: ", err)
		return 0
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		fmt.Println("error req: ", err)
		return 0
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error client: ", err)
		return 0
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error body: ", err)
		return 0
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("error marshal2: ", err)
		return 0
	}

	timestamp, err := strconv.ParseInt(response.Result.Timestamp, 0, 64)
	if err != nil {
		fmt.Println("error parse: ", err)
		return 0

	}

	return timestamp
}

func GetLatestBlock(chain string) string {
	var endpoint string
	if chain == "matic" {
		endpoint = polygonEndpoint
	}
	if chain == "eth" {
		endpoint = ethEndpoint
	}
	payload := map[string]interface{}{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "eth_getBlockByNumber",
		"params":  []interface{}{"latest", true},
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("error marshal: ", err)
		return ""
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		fmt.Println("error req: ", err)
		return ""
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error client: ", err)
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error body: ", err)
		return ""
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("error marshal2: ", err)
		return ""
	}

	return response.Result.Number
}

func searchTron(invoice *SeedInvoice) bool {
	var amount string

	for _, pmt := range invoice.Payments {
		if pmt.Currency == "trx" && pmt.Symbol == "trx" {
			amount = pmt.Amount

		}
	}
	unixTime := strconv.FormatInt(utils.GetUnixTime(invoice.Created)*1000, 10)
	url := "https://api.trongrid.io/v1/accounts/" + TRX_ADDRESS + "/transactions?only_to=true&min_timestamp=" + unixTime

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("error with res", err)
		return false
	}

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("error marshal2: ", err)
		return false

	}
	fmt.Println("tron tron response", response)
	for _, transfer := range response.Data {
		for _, contract := range transfer.Raw.Contract {
			amountStr := strconv.FormatFloat(float64(contract.Param.Value.Amount)/1e6, 'f', 5, 64)
			fmt.Println("expect:", amount, "got:", amountStr)
			if amountStr == amount {
				invoice.Currency = "trx"
				invoice.CryptoReceived = amountStr
				invoice.IncomingTXID = transfer.TXID
				return true
			}

		}

	}

	return searchTronUSDT(invoice)
}

func searchTronUSDT(invoice *SeedInvoice) bool {
	var amount string

	for _, pmt := range invoice.Payments {
		if pmt.Symbol == "USDT" {
			amount = pmt.Amount

		}
	}

	unixTime := strconv.FormatInt(utils.GetUnixTime(invoice.Created)*1000, 10)
	url := "https://api.trongrid.io/v1/accounts/" + TRX_ADDRESS + "/transactions/trc20?contract_address=TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t&only_to=true&min_timestamp=" + unixTime

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("error with res", err)
		return false
	}

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	fmt.Println("body", string(body))
	var response TRC20Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("error marshal2: ", err)
		return false

	}

	for _, transfer := range response.Data {
		fmt.Println(transfer)

		expectedFloat, _ := strconv.ParseFloat(amount, 64)
		//actualFloat, _ := strconv.ParseFloat(transfer.Amount, 64)
		fmt.Println("expect:", strconv.FormatFloat(expectedFloat*math.Pow10(6), 'f', -1, 64), "got:", transfer.Amount)
		if strconv.FormatFloat(expectedFloat*math.Pow10(6), 'f', -1, 64) == transfer.Amount {
			invoice.Currency = "USDT"
			invoice.CryptoReceived = amount
			invoice.IncomingTXID = transfer.TXID
			return true
		}

	}

	return false
}
