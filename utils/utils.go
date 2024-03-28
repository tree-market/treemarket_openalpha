package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"time"
	. "tree_service/types"
)

var Prices map[string]PriceData

func GetPrices() {
	for {
		//TO-DO: CHECK DATE AND UPDATE SEED PRICE

		url := "https://tradeogre.com/api/v1/markets"
		response, err := http.Get(url)
		if err != nil {
			fmt.Println("Error:", err)

		}

		// Read the response body
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)

		}
		//fmt.Println(string(body))

		var allPrices []map[string]PriceData
		err = json.Unmarshal([]byte(body), &allPrices)
		if err != nil {
			fmt.Println("Error:", err)

		}

		// Filter and extract specific symbols
		symbolsToExtract := []string{"DERO-USDT", "LTC-USDT", "BTC-USDT", "ETH-USDT", "XMR-USDT", "MATIC-USDT", "BNB-USDT"}
		//STILL NEED BCH AND TRON
		//MULTIPLE SOURCES WOULD BE IDEAL
		Prices = make(map[string]PriceData)

		for _, item := range allPrices {
			for key, value := range item {
				// Check if the symbol is in the symbolsToExtract list
				if contains(symbolsToExtract, key) {
					Prices[key] = value
					continue
				}
			}
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
