package main

import (
	"encoding/json"
	"fmt"
	"log"
	. "tree_service/types"

	"go.etcd.io/bbolt"
)

var db *bbolt.DB

func alreadyProcessed(txid string) bool {
	var already_processed bool
	db.View(func(tx *bbolt.Tx) error {
		if b := tx.Bucket([]byte("SALE")); b != nil {
			if ok := b.Get([]byte(txid)); ok != nil { // if existing in bucket
				already_processed = true
				fmt.Println("found this m8:", string(ok))
			}
		}
		return nil
	})

	return already_processed
}

func deroTXIDAlreadyProcessed(txid string) bool {
	var already_processed bool
	db.View(func(tx *bbolt.Tx) error {
		if b := tx.Bucket([]byte("SALE")); b != nil {
			if ok := b.Get([]byte(txid)); ok != nil { // if existing in bucket
				already_processed = true
				/* fmt.Println("found this mate:", string(ok))
				if invoice := b.Get([]byte(ok)); invoice != nil {
					fmt.Println("and this oy!", string(invoice))
				} */
			}
		}
		return nil
	})

	return already_processed
}

func bitcartIDAlreadyProcessed(id string) bool {
	var already_processed bool
	db.View(func(tx *bbolt.Tx) error {
		if b := tx.Bucket([]byte("SALE")); b != nil {
			if ok := b.Get([]byte(id)); ok != nil { // if existing in bucket
				already_processed = true
				/* fmt.Println("found this mate:", string(ok))
				if invoice := b.Get([]byte(ok)); invoice != nil {
					fmt.Println("and this oy!", string(invoice))
				} */
			}
		}
		return nil
	})

	return already_processed
}

func addSeedInvoiceToDB(invoice SeedInvoice) SeedInvoice {
	id := invoice.SeedID
	fmt.Println("here's yer key massa", id)
	invoiceData, err := json.Marshal(invoice)
	if err != nil {
		log.Fatal(err)
	}

	// Store serialized user data in BoltDB
	err = db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("SALE"))
		if bucket == nil {
			return fmt.Errorf("sale bucket not found")
		}
		return bucket.Put([]byte(id), invoiceData)
	})

	if err != nil {
		logger.Error(err, "err updating db")
	}

	if invoice.Currency == "dero" {
		err = db.Update(func(tx *bbolt.Tx) error {
			bucket := tx.Bucket([]byte("SALE"))
			if bucket == nil {
				return fmt.Errorf("sale bucket not found")
			}
			return bucket.Put([]byte(invoice.IncomingTXID), []byte(id))
		})

		if err != nil {
			logger.Error(err, "err updating db")
		}
	}

	return invoice
}

func updateInvoice(invoice *SeedInvoice) *SeedInvoice {
	id := invoice.SeedID
	fmt.Println("here's ya update key massa", id)
	invoiceData, err := json.Marshal(invoice)
	if err != nil {
		log.Fatal(err)
	}

	// Store serialized user data in BoltDB
	err = db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("SALE"))
		if bucket == nil {
			return fmt.Errorf("sale bucket not found")
		}
		return bucket.Put([]byte(id), invoiceData)
	})

	if err != nil {
		logger.Error(err, "err updating db")
	}
	return invoice

}

func retrieveInvoice(id string) *SeedInvoice {
	var invoice SeedInvoice

	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("SALE"))
		if b == nil {
			return fmt.Errorf("bucket SALE not found")
		}

		invoiceBytes := b.Get([]byte(id))
		if invoiceBytes == nil {
			return fmt.Errorf("invoice not found")
		}

		if err := json.Unmarshal(invoiceBytes, &invoice); err != nil {
			return fmt.Errorf("error unmarshaling invoice: %v", err)
		}

		return nil
	})

	return &invoice
}
