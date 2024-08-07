package types

import (
	"strconv"
	"strings"
)

type CustomFloat float64

// CustomFloat type for unmarshaling a float from a string

// UnmarshalJSON custom unmarshaling logic for CustomFloat
func (f *CustomFloat) UnmarshalJSON(data []byte) error {
	// Remove surrounding quotes from the string
	str := strings.Trim(string(data), "\"")

	// Parse the string to float64
	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return err
	}

	// Assign the parsed float value to the CustomFloat type
	*f = CustomFloat(val)
	return nil
}

type PriceData struct {
	InitialPrice string      `json:"initialprice"`
	Price        CustomFloat `json:"price"`
	High         string      `json:"high"`
	Low          string      `json:"low"`
	Volume       string      `json:"volume"`
	Bid          string      `json:"bid"`
	Ask          string      `json:"ask"`
	Basename     string      `json:"basename"`
}

type BitcartPayment struct {
	Rate     string `json:"rate"`
	Currency string `json:"currency"`
	Symbol   string `json:"symbol"`
	Amount   string `json:"amount"`
	Address  string `json:"payment_address"`
	URL      string `json:"payment_url"`
}
type BitcartInvoice struct {
	Address  string           `json:"shipping_address"`
	Currency string           `json:"paid_currency"`
	Status   string           `json:"status"`
	TXID     []string         `json:"tx_hashes"`
	Amount   string           `json:"sent_amount"`
	Payments []BitcartPayment `json:"payments"`
	Store    string           `json:"store_id"`
	Price    string           `json:"price"`
	Created  string           `json:"created"`
	Timeout  uint64           `json:"expiration_seconds"`
	ID       string           `json:"id"`
}

type SeedInvoice struct {
	SeedID             string
	Quantity           float64
	Currency           string
	Status             string
	CryptoReceived     string
	CryptoRefunded     uint64
	Email              string
	DeroAddress        string
	SeedSent           float64
	BitcartID          string
	IncomingTXID       string
	SeedOutTXID        string
	RefundTXID         string
	Created            string
	Timeout            uint64
	Integrated         string
	Payments           []BitcartPayment
	BitcartStatus      string
	Password           string
	EthID              int
	Blocks             Blocks
	CompletedTimestamp string
}

type Blocks struct {
	ETH   string
	MATIC string
	TRX   string
}

type Transfer struct {
	From     string  `json:"from"`
	To       string  `json:"to"`
	Quantity float64 `json:"value"`
	Currency string  `json:"asset"`
	Block    string  `json:"blockNum"`
	TXID     string  `json:"hash"`
}

// Define the Go struct for the result object
type Result struct {
	Transfers []Transfer `json:"transfers"`
	Timestamp string     `json:"timestamp"`
	Number    string     `json:"number"`
}

type Response struct {
	Result Result     `json:"result"`
	Data   []TronTran `json:"data"`
}

type TronTran struct {
	Raw  TronRaw `json:"raw_data"`
	TXID string  `json:"txID"`
	Time int     `json:"block_timestamp"`
}

type TronRaw struct {
	Contract []TronCon `json:"contract"`
}

type TronCon struct {
	Param TronParam `json:"parameter"`
}

type TronParam struct {
	Value TronValue `json:"value"`
}

type TronValue struct {
	Amount uint64 `json:"amount"`
}

type TRC20Response struct {
	Data []TRC20Tran `json:"data"`
}

type TRC20Tran struct {
	Amount string `json:"value"`
	TXID   string `json:"transaction_id"`
}

type DeroRequest struct {
	Json   string        `json:"jsonrpc"`
	ID     string        `json:"id"`
	Method string        `json:"method"`
	Params DeroReqParams `json:"params"`
	Result DeroResult    `json:"result"`
}

type DeroReqParams struct {
	TXIDs []string `json:"txs_hashes"`
}

type DeroResult struct {
	TXData []string `json:"txs_as_hex"`
}
