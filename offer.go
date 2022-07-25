package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const (
	OfferDoc DocType = "offer"
)

type OfferInner struct {
	Doc

	ID             string `json:"id"`
	Value          Price  `json:"value"`
	OrganizationID string `json:"organization_id"`
	RequestID      string `json:"request_id"`
}

type Offer struct {
	ID             string `json:"id"`
	Value          Price  `json:"value"`
	OrganizationID string `json:"organization_id"`
	RequestID      string `json:"request_id"`
}

func (s *SmartContract) FromOfferInner(_ contractapi.TransactionContextInterface, p *OfferInner) *Offer {
	return &Offer{
		ID: p.ID,
		Value: Price{
			Amount:   p.Value.Amount,
			Currency: p.Value.Currency,
			Exponent: p.Value.Exponent,
		},
		OrganizationID: p.OrganizationID,
		RequestID:      p.RequestID,
	}
}

func (s *SmartContract) GetOfferID(_ contractapi.TransactionContextInterface, id string) string {
	return string(OfferDoc) + "_" + id
}

func (s *SmartContract) OfferExist(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(s.GetOfferID(ctx, id))
	if err != nil {
		return false, fmt.Errorf("failed to read from world state:%v", err)
	}

	return assetJSON != nil, nil
}

func (s *SmartContract) MakeOffer(ctx contractapi.TransactionContextInterface, id string, value uint32, currency string, exponent uint32, organizationID string, requestID string) error {
	if err := s.HasPermission(ctx, OffersCreate); err != nil {
		return err
	}

	exists, err := s.OfferExist(ctx, id)
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

	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	doc := Doc{
		Type:      OfferDoc,
		CreatedBy: clientID,
		UpdatedBy: clientID,
	}

	offer := OfferInner{
		Doc: doc,
		ID:  s.GetOfferID(ctx, id),
		Value: Price{
			Amount:   value,
			Currency: currency,
			Exponent: exponent,
		},
		OrganizationID: organizationID,
		RequestID:      requestID,
	}

	assetBytes, err := json.Marshal(offer)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(offer.ID, assetBytes)
	if err != nil {
		return err
	}

	return nil
}

func (s *SmartContract) GetOfferInner(ctx contractapi.TransactionContextInterface, id string) (*OfferInner, error) {
	if err := s.HasPermission(ctx, OffersRead); err != nil {
		return nil, err
	}

	assetBytes, err := ctx.GetStub().GetState(s.GetOfferID(ctx, id))
	if err != nil {
		return nil, fmt.Errorf("failed to get asset %s:%v", id, err)
	}

	if assetBytes == nil {
		return nil, fmt.Errorf("asset %s does not exist", id)
	}

	var o OfferInner
	err = json.Unmarshal(assetBytes, &o)
	if err != nil {
		return nil, err
	}

	o.ID = strings.TrimPrefix(o.ID, string(OfferDoc)+"_")
	return &o, nil
}

func (s *SmartContract) GetOffer(ctx contractapi.TransactionContextInterface, id string) (*Offer, error) {
	if err := s.HasPermission(ctx, OffersRead); err != nil {
		return nil, err
	}

	assetBytes, err := ctx.GetStub().GetState(s.GetOfferID(ctx, id))
	if err != nil {
		return nil, fmt.Errorf("failed to get asset %s:%v", id, err)
	}

	if assetBytes == nil {
		return nil, fmt.Errorf("asset %s does not exist", id)
	}

	var o OfferInner
	err = json.Unmarshal(assetBytes, &o)
	if err != nil {
		return nil, err
	}

	o.ID = strings.TrimPrefix(o.ID, string(OfferDoc)+"_")
	return s.FromOfferInner(ctx, &o), nil
}

func (s *SmartContract) GetAllOffersForRequestInner(ctx contractapi.TransactionContextInterface, requestID string) ([]*OfferInner, error) {
	if err := s.HasPermission(ctx, OffersRead); err != nil {
		return nil, err
	}

	results, err := ctx.GetStub().GetQueryResult(fmt.Sprintf(`{"selector":{"doc_type":"%s","request_id":"%s"}}`, OfferDoc, requestID))
	if err != nil {
		return nil, fmt.Errorf("failed to get assets:%v", err)
	}
	defer results.Close()

	var assets []*OfferInner

	for results.HasNext() {
		queryResult, err := results.Next()
		if err != nil {
			return nil, err
		}
		var o OfferInner
		err = json.Unmarshal(queryResult.Value, &o)
		if err != nil {
			return nil, err
		}

		o.ID = strings.TrimPrefix(o.ID, string(OfferDoc)+" _ ")
		assets = append(assets, &o)
	}

	return assets, nil
}

func (s *SmartContract) GetAllOffersForRequest(ctx contractapi.TransactionContextInterface, requestID string) ([]*Offer, error) {
	if err := s.HasPermission(ctx, OffersRead); err != nil {
		return nil, err
	}

	results, err := ctx.GetStub().GetQueryResult(fmt.Sprintf(`{"selector":{"doc_type":"%s","request_id":"%s"}}`, OfferDoc, requestID))
	if err != nil {
		return nil, fmt.Errorf("failed to get assets:%v", err)
	}
	defer results.Close()

	var assets []*Offer
	for results.HasNext() {
		queryResult, err := results.Next()
		if err != nil {
			return nil, err
		}
		var o OfferInner
		err = json.Unmarshal(queryResult.Value, &o)
		if err != nil {
			return nil, err
		}

		o.ID = strings.TrimPrefix(o.ID, string(OfferDoc)+"_")
		assets = append(assets, s.FromOfferInner(ctx, &o))
	}

	return assets, nil
}
