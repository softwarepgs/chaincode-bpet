package main

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"log"
)

func main() {
	assetChaincode, err := contractapi.NewChaincode(&SmartContract{
		checkPermissions: true,
	})
	if err != nil {
		log.Panicf("Error creating asset - transfer - basic chaincode : % v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting asset - transfer - basic chaincode : % v", err)
	}
}
