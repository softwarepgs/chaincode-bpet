package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const (
	RequestDoc DocType = "request"
)

type RequestInner struct {
	Doc

	ID          string        `json:"id"`
	Value       Value         `json:"value"`
	Description string        `json:"description"`
	Status      RequestStatus `json:"status"`
}

type Request struct {
	ID          string        `json:"id"`
	Value       Value         `json:"value"`
	Description string        `json:"description"`
	Status      RequestStatus `json:"status"`
}

func (s *SmartContract) FromRequestInner(ctx contractapi.TransactionContextInterface, p *RequestInner) *Request {
	return &Request{
		ID: p.ID,
		Value: Value{
			Amount:   p.Value.Amount,
			Currency: p.Value.Currency,
		},
		Description: p.Description,
		Status:      p.Status,
	}
}

func (s *SmartContract) GetRequestID(ctx contractapi.TransactionContextInterface, id string) string {
	return string(RequestDoc) + "_" + id
}

func (s *SmartContract) RequestExist(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(s.GetRequestID(ctx, id))
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

func (s *SmartContract) CreateRequest(ctx contractapi.TransactionContextInterface, id string, value uint32, currency string, description string) error {
	if err := s.HasPermission(ctx, RequestsCreate); err != nil {
		return err
	}

	exists, err := s.RequestExist(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	doc := Doc{
		Type:      RequestDoc,
		CreatedBy: clientID,
		UpdatedBy: clientID,
	}

	r := RequestInner{
		Doc: doc,
		ID:  s.GetRequestID(ctx, id),
		Value: Value{
			Amount:   value,
			Currency: currency,
		},
		Description: description,
		Status:      RequestStatusOpen,
	}

	assetBytes, err := json.Marshal(r)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(r.ID, assetBytes)
	if err != nil {
		return err
	}

	return nil
}

func (s *SmartContract) CloseRequest(ctx contractapi.TransactionContextInterface, id string) error {
	if err := s.HasPermission(ctx, RequestsUpdate); err != nil {
		return err
	}

	exists, err := s.RequestExist(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	request, err := s.GetRequestInner(ctx, id)
	if err != nil {
		return err
	}

	if request.Status == RequestStatusClosed {
		return fmt.Errorf("can't close")
	}

	request.Status = RequestStatusClosed

	assetBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(request.ID, assetBytes)

	return nil
}

func (s *SmartContract) GetRequest(ctx contractapi.TransactionContextInterface, id string) (*Request, error) {
	if err := s.HasPermission(ctx, RequestsRead); err != nil {
		return nil, err
	}

	assetBytes, err := ctx.GetStub().GetState(s.GetRequestID(ctx, id))
	if err != nil {
		return nil, fmt.Errorf("failed to get asset %s:%v", id, err)
	}

	if assetBytes == nil {
		return nil, fmt.Errorf("asset %s does not exist", id)
	}

	var r RequestInner
	err = json.Unmarshal(assetBytes, &r)
	if err != nil {
		return nil, err
	}

	r.ID = strings.TrimPrefix(r.ID, string(RequestDoc)+"_")
	return s.FromRequestInner(ctx, &r), nil
}

func (s *SmartContract) GetRequestInner(ctx contractapi.TransactionContextInterface, id string) (*RequestInner, error) {
	if err := s.HasPermission(ctx, RequestsRead); err != nil {
		return nil, err
	}

	assetBytes, err := ctx.GetStub().GetState(s.GetRequestID(ctx, id))
	if err != nil {
		return nil, fmt.Errorf("failed to get asset %s:%v", id, err)
	}

	if assetBytes == nil {
		return nil, fmt.Errorf("asset %s does not exist", id)
	}

	var r RequestInner
	err = json.Unmarshal(assetBytes, &r)
	if err != nil {
		return nil, err
	}

	r.ID = strings.TrimPrefix(r.ID, string(RequestDoc)+"_")
	return &r, nil
}

func (s *SmartContract) ListRequests(ctx contractapi.TransactionContextInterface) ([]*Request, error) {
	if err := s.HasPermission(ctx, RequestsRead); err != nil {
		return nil, err
	}

	results, err := ctx.GetStub().GetQueryResult(fmt.Sprintf(`{"selector":{"doc_type":"%s"}}`, RequestDoc))
	if err != nil {
		return nil, fmt.Errorf("failed to get assets:%v", err)
	}
	defer results.Close()

	var assets []*Request
	for results.HasNext() {
		queryResult, err := results.Next()
		if err != nil {
			return nil, err
		}
		var r RequestInner
		err = json.Unmarshal(queryResult.Value, &r)
		if err != nil {
			return nil, err
		}

		r.ID = strings.TrimPrefix(r.ID, string(RequestDoc)+"_")
		assets = append(assets, s.FromRequestInner(ctx, &r))
	}
	return assets, nil
}

func (s *SmartContract) ListRequestsByStatus(ctx contractapi.TransactionContextInterface, statusInput string) ([]*Request, error) {
	if err := s.HasPermission(ctx, RequestsRead); err != nil {
		return nil, err
	}

	status, err := ParseOrderStatus(statusInput)
	if err != nil {
		return nil, err
	}

	results, err := ctx.GetStub().GetQueryResult(fmt.Sprintf(`{"selector":{"doc_type":"%s","status":"%s"}}`, RequestDoc, status))
	if err != nil {
		return nil, fmt.Errorf("failed to get assets:%v", err)
	}
	defer results.Close()

	var assets []*Request
	for results.HasNext() {
		queryResult, err := results.Next()
		if err != nil {
			return nil, err
		}
		var r RequestInner
		err = json.Unmarshal(queryResult.Value, &r)
		if err != nil {
			return nil, err
		}

		r.ID = strings.TrimPrefix(r.ID, string(RequestDoc)+"_")
		assets = append(assets, s.FromRequestInner(ctx, &r))
	}

	return assets, nil
}
