package utils

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	"time"
	. "tree_service/types"
)

var Prices map[string]PriceData

func CheckDeroTransaction(txid string) bool {
	url := "http://tree.market:10102/json_rpc"

	txids := []string{txid}

	params := DeroReqParams{
		TXIDs: txids,
	}
	request := DeroRequest{
		Json:   "2.0",
		ID:     "1",
		Method: "DERO.GetTransaction",
		Params: params,
	}

	requestData, err := json.Marshal(request)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return false
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestData))
	if err != nil {
		fmt.Println("Error creating http request:", err)
		return false
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending HTTP request:", err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error:", resp.Status)
		return false
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return false
	}
	var respData DeroRequest
	err = json.Unmarshal(body, &respData)
	if err != nil {
		fmt.Println("Error unmarshaling:", err)
		return false
	}
	if respData.Result.TXData[0] != "" {
		return true
	}

	return false
}

func GetPrices() {
	for {
		url := "https://tradeogre.com/api/v1/markets"
		response, err := http.Get(url)
		if err != nil {
			fmt.Println("Error making GET request:", err)
			time.Sleep(time.Second * 60)
			continue
		}

		// Ensure the response body is closed to avoid resource leaks
		if response != nil {
			// Check for a successful status code
			if response.StatusCode != http.StatusOK {
				fmt.Printf("Received non-200 response code: %d\n", response.StatusCode)
				response.Body.Close()
				time.Sleep(time.Second * 60)
				continue
			}

			// Read the response body
			body, err := ioutil.ReadAll(response.Body)
			response.Body.Close() // Close the response body after reading
			if err != nil {
				fmt.Println("Error reading response body:", err)
				time.Sleep(time.Second * 60)
				continue
			}

			var allPrices []map[string]PriceData
			err = json.Unmarshal(body, &allPrices)
			if err != nil {
				fmt.Println("Error unmarshaling JSON response:", err)
				fmt.Printf("Response body: %s\n", string(body)) // Log the raw response body for debugging
				time.Sleep(time.Second * 60)
				continue
			}

			// Filter and extract specific symbols
			symbolsToExtract := []string{"DERO-USDT", "LTC-USDT", "BTC-USDT", "ETH-USDT", "XMR-USDT", "MATIC-USDT", "BNB-USDT"}

			Prices = make(map[string]PriceData)
			for _, item := range allPrices {
				for key, value := range item {
					if contains(symbolsToExtract, key) {
						Prices[key] = value
					}
				}
			}

			fmt.Println("Updated prices:", Prices) // Log the updated prices
		}

		time.Sleep(time.Second * 60)
	}
}

func GenerateRandomPassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;:',.<>/?`~"
	password := make([]byte, length)
	max := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		password[i] = charset[n.Int64()]
	}

	return string(password), nil
}

func CalculateSHA256(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

func GetUnixTime(date string) int64 {
	timeObj, err := time.Parse(time.RFC3339Nano, date)
	if err != nil {
		fmt.Println("Error parsing time string:", err)
		return 0
	}

	// Convert the time object to Unix time (seconds since Unix epoch)
	return timeObj.Unix()
}

func contains(slice []string, element string) bool {
	for _, item := range slice {
		if item == element {
			return true
		}
	}
	return false
}
