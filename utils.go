package main

import "github.com/btcsuite/btcd/btcutil/bech32"

func validateBech32Address(address string) bool {
	_, _, err := bech32.Decode(address)
	return err == nil
}
