package main

import (
	"encoding/base64"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
	checkPermissions bool
}

//Returns current user's ID
func (s *SmartContract) GetSubmittingClientIdentity(ctx contractapi.TransactionContextInterface) (string, error) {
	b64ID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return " ", fmt.Errorf("failed to read clientID : % v", err)
	}

	decodeID, err := base64.StdEncoding.DecodeString(b64ID)
	if err != nil {
		return " ", fmt.Errorf("failed to base64 decode clientID : % v", err)
	}

	return string(decodeID), nil
}

//Returns current user's organization
func (s *SmartContract) GetSubmittingClientOrganization(ctx contractapi.TransactionContextInterface) (string, error) {
	return ctx.GetClientIdentity().GetMSPID()
}
