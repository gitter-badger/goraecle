package main

import (
	"log"

	"github.com/aeternity/aepp-sdk-go/v7/account"
	"github.com/aeternity/aepp-sdk-go/v7/aeternity"
	"github.com/aeternity/aepp-sdk-go/v7/config"
	"github.com/aeternity/aepp-sdk-go/v7/naet"
	"github.com/aeternity/aepp-sdk-go/v7/transactions"
)

const (
	networkID    = "ae_uat"
	nodeURL      = "http://sdk-testnet.aepps.com"
	oraclePubKey = "ok_2FDtT4tP8PQdakf9QQL8XAKM2eTfukTc6YUcmf54n22TCo2Uks"
)

func setupNetwork(debug bool) *naet.Node {
	config.Node.NetworkID = networkID
	return naet.NewNode(nodeURL, debug)
}

func setupAccount() *account.Account {
	acc, err := account.FromHexString(clientPrivateKey)
	if err != nil {
		log.Fatalf("setup account error: %v", err)
	}

	return acc
}

func main() {
	account := setupAccount()
	node := setupNetwork(false)
	_, _, ttlnoncer := transactions.GenerateTTLNoncer(node)

	// Queryttlnoncer
	query, err := transactions.NewOracleQueryTx(account.Address, oraclePubKey, "hello:name=Arjan van Eersel", config.Client.Oracles.QueryFee, 0, 100, 0, 100, ttlnoncer)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("prepared query: %+v", query)

	_, _, _, _, _, err = aeternity.SignBroadcastWaitTransaction(query, account, node, networkID, config.Client.WaitBlocks)
	if err != nil {
		log.Fatal(err)
	}

}
