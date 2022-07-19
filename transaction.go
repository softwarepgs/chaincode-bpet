package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const (
	TransactionDoc                   DocType = "transaction"
	NewTransactionEventKey                   = "transaction_created"
	TransactionStatusChangedEventKey         = "transaction_status_changed"
)

type TransactionInner struct {
	Doc

	ID             string            `json:"id"`
	Amount         uint32            `json:"amount"`
	Description    string            `json:"description"`
	Status         TransactionStatus `json:"status"`
	OrganizationID string            `json:"organization_id"`
	OrderID        string            `json:"order_id"`
}

type Transaction struct {
	ID             string            `json:"id"`
	Amount         uint32            `json:"amount"`
	Description    string            `json:"description"`
	Status         TransactionStatus `json:"status"`
	OrganizationID string            `json:"organization_id"`
	OrderID        string            `json:"order_id"`
}

type NewTransactionEvent struct {
	TransactionID string `json:"transaction_id"`
}

type TransactionStatusChangedEvent struct {
	TransactionID string            `json:"transaction_id"`
	OldStatus     TransactionStatus `json:"old_status"`
	NewStatus     TransactionStatus `json:"new_status"`
	Message       string            `json:"message"`
}

func (s *SmartContract) FromTransactionInner(_ contractapi.TransactionContextInterface, p *TransactionInner) *Transaction {
	return &Transaction{
		ID:             p.ID,
		Amount:         p.Amount,
		Description:    p.Description,
		Status:         p.Status,
		OrganizationID: p.OrganizationID,
		OrderID:        p.OrderID,
	}
}

func (s *SmartContract) GetTransactionID(_ contractapi.TransactionContextInterface, id string) string {
	return string(TransactionDoc) + "_" + id
}

func getTransactionsAmount(transactions []*TransactionInner) uint32 {
	out := uint32(0)

	for _, t := range transactions {
		if t.Status == TransactionStatusCanceled {
			continue
		}

		out += t.Amount
	}

	return out
}

func NewNewTransactionEvent(id string) ([]byte, error) {
	return json.Marshal(NewTransactionEvent{TransactionID: id})
}

func NewTransactionStatusChangedEvent(id string, oldStatus TransactionStatus, newStatus TransactionStatus, message string) ([]byte, error) {
	return json.Marshal(TransactionStatusChangedEvent{TransactionID: id, OldStatus: oldStatus, NewStatus: newStatus, Message: message})
}

func (s *SmartContract) TransactionExist(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(s.GetTransactionID(ctx, id))
	if err != nil {
		return false, fmt.Errorf("failed to read from world state:%v", err)
	}

	return assetJSON != nil, nil
}

func (s *SmartContract) MakeTransaction(ctx contractapi.TransactionContextInterface, id string, amount uint32, organizationID string, orderID string) error {
	if err := s.HasPermission(ctx, TransactionsCreate); err != nil {
		return err
	}

	exists, err := s.TransactionExist(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	hasOrg, err := s.OrganizationExist(ctx, organizationID)
	if err != nil {
		return err
	}
	if hasOrg {
		return fmt.Errorf("organization %s does not exist", id)
	}

	order, err := s.GetOrder(ctx, orderID)
	if err != nil {
		return err
	}

	transactions, err := s.ListTransactionsForOrderInner(ctx, orderID)
	if err != nil {
		return err
	}

	transactionsAmount := getTransactionsAmount(transactions)
	if transactionsAmount+amount > order.Amount {
		return fmt.Errorf("invalid amount to transact")
	}

	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	doc := Doc{
		Type:      TransactionDoc,
		CreatedBy: clientID,
		UpdatedBy: clientID,
	}

	transaction := TransactionInner{
		Doc:            doc,
		ID:             s.GetUnitID(ctx, id),
		Amount:         amount,
		Status:         TransactionStatusOpen,
		OrganizationID: organizationID,
		OrderID:        orderID,
	}

	assetBytes, err := json.Marshal(transaction)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(transaction.ID, assetBytes)
	if err != nil {
		return err
	}

	eventBody, err := NewNewTransactionEvent(transaction.ID)
	if err != nil {
		return err
	}

	err = ctx.GetStub().SetEvent(NewTransactionEventKey, eventBody)
	if err != nil {
		return err
	}

	return nil
}

func (s *SmartContract) ChangeStatus(ctx contractapi.TransactionContextInterface, id string, inputStatus, message string) error {
	if err := s.HasPermission(ctx, TransactionsUpdate); err != nil {
		return err
	}

	status, err := ParseTransactionStatus(inputStatus)
	if err != nil {
		return err
	}

	exists, err := s.TransactionExist(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	transaction, err := s.GetTransactionInner(ctx, id)
	if err != nil {
		return err
	}

	if transaction.Status == TransactionStatusClosed || transaction.Status == TransactionStatusCanceled {
		return fmt.Errorf("transaction already closed or canceled")
	}

	order, err := s.GetOrderInner(ctx, transaction.OrderID)
	if err != nil {
		return err
	}

	orgID, err := s.GetSubmittingClientOrganization(ctx)
	if err != nil {
		return err
	}

	switch orgID {
	case transaction.OrganizationID:
		if status == TransactionStatusPaid ||
			status == TransactionStatusInReview ||
			status == TransactionStatusWaitingPayment ||
			status == TransactionStatusReady ||
			status == TransactionStatusInProgress {
			return fmt.Errorf("you do not have permissions to do that")
		}
	case order.OrganizationID:
	default:
		return fmt.Errorf("you do not have permissions to do that")
	}

	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	oldStatus := transaction.Status
	transaction.Status = status
	transaction.Description = message
	transaction.UpdatedBy = clientID

	assetBytes, err := json.Marshal(transaction)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(transaction.ID, assetBytes)
	if err != nil {
		return err
	}

	eventBody, err := NewTransactionStatusChangedEvent(transaction.ID, oldStatus, status, message)
	if err != nil {
		return err
	}

	err = ctx.GetStub().SetEvent(TransactionStatusChangedEventKey, eventBody)
	if err != nil {
		return err
	}

	return nil
}

func (s *SmartContract) GetTransactionInner(ctx contractapi.TransactionContextInterface, id string) (*TransactionInner, error) {
	if err := s.HasPermission(ctx, TransactionsRead); err != nil {
		return nil, err
	}

	assetBytes, err := ctx.GetStub().GetState(s.GetOrderID(ctx, id))
	if err != nil {
		return nil, fmt.Errorf("failed to get asset %s:%v", id, err)
	}

	if assetBytes == nil {
		return nil, fmt.Errorf("asset %s does not exist", id)
	}

	var unit TransactionInner
	err = json.Unmarshal(assetBytes, &unit)
	if err != nil {
		return nil, err
	}

	unit.ID = strings.TrimPrefix(unit.ID, string(TransactionDoc)+"_")
	return &unit, nil
}

func (s *SmartContract) GetTransaction(ctx contractapi.TransactionContextInterface, id string) (*Transaction, error) {
	if err := s.HasPermission(ctx, TransactionsRead); err != nil {
		return nil, err
	}

	assetBytes, err := ctx.GetStub().GetState(s.GetOrderID(ctx, id))
	if err != nil {
		return nil, fmt.Errorf("failed to get asset %s:%v", id, err)
	}

	if assetBytes == nil {
		return nil, fmt.Errorf("asset %s does not exist", id)
	}

	var unit TransactionInner
	err = json.Unmarshal(assetBytes, &unit)
	if err != nil {
		return nil, err
	}

	unit.ID = strings.TrimPrefix(unit.ID, string(TransactionDoc)+"_")
	return s.FromTransactionInner(ctx, &unit), nil
}

func (s *SmartContract) ListTransactionsForOrderInner(ctx contractapi.TransactionContextInterface, orderID string) ([]*TransactionInner, error) {
	if err := s.HasPermission(ctx, TransactionsRead); err != nil {
		return nil, err
	}

	results, err := ctx.GetStub().GetQueryResult(fmt.Sprintf(`{"selector":{"doc_type":"%s","order_id":"%s"}}`, TransactionDoc, orderID))
	if err != nil {
		return nil, fmt.Errorf("failed to get assets:%v", err)
	}
	defer results.Close()

	var assets []*TransactionInner

	for results.HasNext() {
		queryResult, err := results.Next()
		if err != nil {
			return nil, err
		}
		var unit TransactionInner
		err = json.Unmarshal(queryResult.Value, &unit)
		if err != nil {
			return nil, err
		}

		unit.ID = strings.TrimPrefix(unit.ID, string(TransactionDoc)+" _ ")
		assets = append(assets, &unit)
	}

	return assets, nil
}

func (s *SmartContract) ListTransactionsForOrder(ctx contractapi.TransactionContextInterface, orderID string) ([]*Transaction, error) {
	if err := s.HasPermission(ctx, TransactionsRead); err != nil {
		return nil, err
	}

	results, err := ctx.GetStub().GetQueryResult(fmt.Sprintf(`{"selector":{"doc_type":"%s","order_id":"%s"}}`, TransactionDoc, orderID))
	if err != nil {
		return nil, fmt.Errorf("failed to get assets:%v", err)
	}
	defer results.Close()

	var assets []*Transaction
	for results.HasNext() {
		queryResult, err := results.Next()
		if err != nil {
			return nil, err
		}
		var unit TransactionInner
		err = json.Unmarshal(queryResult.Value, &unit)
		if err != nil {
			return nil, err
		}

		unit.ID = strings.TrimPrefix(unit.ID, string(TransactionDoc)+"_")
		assets = append(assets, s.FromTransactionInner(ctx, &unit))
	}

	return assets, nil
}
