package main

//TO-DO: WRITE TESTING APP THAT CAN HIT THIS WITH SEVERAL TX PER SECOND

import (
	"crypto/sha1"

	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/deroproject/derohe/globals"
	"github.com/go-logr/logr"
	"github.com/ybbus/jsonrpc"
	"go.etcd.io/bbolt"
	"gopkg.in/natefinch/lumberjack.v2"

	"time"

	"github.com/gorilla/mux"

	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/rpc"

	"tree_service/eth"

	"tree_service/utils"
)

var DEROUSD float64

const PLUGIN_NAME = "tree_service"

var seedPrice = 2.99

var logger logr.Logger = logr.Discard()

var connected bool = false

func initRPCClient() {

	for !connected {
		fmt.Println("trying to connect")
		rpcClient = jsonrpc.NewClient("http://127.0.0.1:10103/json_rpc")

		var addr_result rpc.GetAddress_Result
		err := rpcClient.CallFor(&addr_result, "GetAddress")
		if err != nil || addr_result.Address == "" {
			fmt.Printf("Could not obtain address from wallet err %s\n", err)
			continue
		}

		connected = true
	}

}

func main() {

	file, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer file.Close()

	// Set the file as the output for the logger
	log.SetOutput(file)
	defer func() {
		if r := recover(); r != nil {
			// Log the panic or handle it as needed
			log.Println("Recovered from panic:", r)
			// Restart the program by calling mainLogic again
			main()
		}
	}()
	// Call your original main logic (now named mainLogic)
	mainLogic()

}

func mainLogic() {
	initRPCClient()
	eth.ConnectToEth()

	router := mux.NewRouter()
	router.HandleFunc("/service/checkInvoiceStatus/{id}", checkInvoiceStatus).Methods("GET")

	router.HandleFunc("/service/newTreeInvoice/", newTreeInvoice).Methods("POST")
	router.HandleFunc("/service/checkTransaction/{txid}", checkTransaction).Methods("GET")

	var err error

	// parse arguments and setup logging , print basic information
	globals.Arguments["--debug"] = true
	exename, _ := os.Executable()
	globals.InitializeLog(os.Stdout, &lumberjack.Logger{
		Filename:   exename + ".log",
		MaxSize:    100, // megabytes
		MaxBackups: 2,
	})
	logger = globals.Logger

	var addr_result rpc.GetAddress_Result
	err = rpcClient.CallFor(&addr_result, "GetAddress")
	if err != nil || addr_result.Address == "" {
		fmt.Printf("Could not obtain address from wallet err %s\n", err)
		return
	}

	if addr, err = rpc.NewAddress(addr_result.Address); err != nil {
		fmt.Printf("address could not be parsed: addr:%s err:%s\n", addr_result.Address, err)
		return
	}

	shasum := fmt.Sprintf("%x", sha1.Sum([]byte(addr.String())))

	db_name := fmt.Sprintf("%s_%s.bbolt.db", PLUGIN_NAME, shasum)
	db, err = bbolt.Open(db_name, 0600, nil)
	if err != nil {
		fmt.Printf("could not open db err:%s\n", err)
		return
	}
	//defer db.Close()

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("SALE"))
		fmt.Println("err", err)
		return err
	})
	if err != nil {
		fmt.Printf("err creating bucket. err %s\n", err)
	}

	fmt.Printf("Persistant store created in '%s'\n", db_name)

	fmt.Printf("Wallet Address: %s\n", addr)
	service_address_without_amount := addr.Clone()
	service_address_without_amount.Arguments = expected_arguments[:len(expected_arguments)-1]

	fmt.Printf("Integrated address to activate '%s', (without hardcoded amount) service: \n%s\n", PLUGIN_NAME, service_address_without_amount.String())

	service_address := addr.Clone()
	service_address.Arguments = expected_arguments
	fmt.Printf("Integrated address to activate '%s', service: \n%s\n", PLUGIN_NAME, service_address.String())

	go utils.GetPrices()

	go listenForDeroPayments()
	log.Fatal(http.ListenAndServe(":5001", router))

}

// TO-DO maybe no necessary after all
func retryInvoice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invoiceID := vars["id"]
	// for retrying I would check seed price based on when the payment was made
	// instead of using the "current" seed price
	checkInvoice(invoiceID)
	//w.Header().Set("Access-Control-Allow-Origin", "*")
	//w.Header().Set("Content-Type", "application/json")

	//w.Write()
}

func withdrawSeed(amount uint64) {
	fmt.Println("withdrawing ", amount, " seed")

	var result rpc.Transfer_Result
	tparams := rpc.Transfer_Params{Ringsize: 2, SC_ID: DISTRIBUTOR_SCID, SC_RPC: []rpc.Argument{{Name: "entrypoint", DataType: "S", Value: "Withdraw"}, {Name: "amount", DataType: "U", Value: amount}, {Name: "token", DataType: "S", Value: SEED_SCID}}, Transfers: []rpc.Transfer{{Destination: "deto1qyre7td6x9r88y4cavdgpv6k7lvx6j39lfsx420hpvh3ydpcrtxrxqg8v8e3z", SCID: crypto.HashHexToHash(BOT_SCID), Burn: 1}}}

	err := rpcClient.CallFor(&result, "Transfer", tparams)
	if err != nil {
		logger.Error(err, "err while transfer")
		return
	}

}

func addToTreasury(amount uint64) {
	// will spin up oao to add dero to here
}

func sendSeed(recipient string, amount uint64, change uint64, txid string) {
	time.Sleep(time.Second * 60)
	var result rpc.Transfer_Result
	tparams := rpc.Transfer_Params{Transfers: []rpc.Transfer{{Destination: recipient, SCID: crypto.HashHexToHash(SEED_SCID), Amount: amount}, {Destination: recipient, Amount: change + 1, Payload_RPC: response}}}
	err := rpcClient.CallFor(&result, "Transfer", tparams)
	if err != nil {
		logger.Error(err, "err while transfer")
		return
	}

}
