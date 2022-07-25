package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const (
	UnitDoc DocType = "unit"
)

type UnitInner struct {
	Doc

	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Exponent    uint32 `json:"exponent"`
}

type Unit struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Exponent    uint32 `json:"exponent"`
}

func FromUnitInner(u *UnitInner) *Unit {
	if u == nil {
		return nil
	}

	return &Unit{
		ID:          u.ID,
		Name:        u.Name,
		Description: u.Description,
		Exponent:    u.Exponent,
	}
}

func (s *SmartContract) GetUnitID(_ contractapi.TransactionContextInterface, id string) string {
	return string(UnitDoc) + "_" + id
}

func (s *SmartContract) CreateUnit(ctx contractapi.TransactionContextInterface, id string, name string, description string, exponent uint32) error {
	if err := s.HasPermission(ctx, UnitsCreate); err != nil {
		return err
	}

	exists, err := s.UnitExist(ctx, id)
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
		Type:      UnitDoc,
		CreatedBy: clientID,
		UpdatedBy: clientID,
	}

	unit := UnitInner{
		ID:          s.GetUnitID(ctx, id),
		Name:        name,
		Description: description,
		Exponent:    exponent,
		Doc:         doc,
	}

	assetBytes, err := json.Marshal(unit)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(unit.ID, assetBytes)
	if err != nil {
		return err
	}

	return nil
}

func (s *SmartContract) UnitExist(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(s.GetUnitID(ctx, id))
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

func (s *SmartContract) UnitsExist(ctx contractapi.TransactionContextInterface, ids []string) error {
	for _, id := range ids {
		e, err := s.UnitExist(ctx, id)

		if err != nil {
			return err
		}

		if !e {
			return fmt.Errorf("unit %s does not exist", id)
		}
	}

	return nil
}

func (s *SmartContract) GetUnitInner(ctx contractapi.TransactionContextInterface, id string) (*UnitInner, error) {
	if err := s.HasPermission(ctx, UnitsRead); err != nil {
		return nil, err
	}

	assetBytes, err := ctx.GetStub().GetState(s.GetUnitID(ctx, id))
	if err != nil {
		return nil, fmt.Errorf("failed to get asset %s: %v", id, err)
	}

	if assetBytes == nil {
		return nil, fmt.Errorf("asset %s does not exist", id)
	}

	var unit UnitInner
	err = json.Unmarshal(assetBytes, &unit)
	if err != nil {
		return nil, err
	}

	unit.ID = strings.TrimPrefix(unit.ID, string(UnitDoc)+"_")
	return &unit, nil
}

func (s *SmartContract) GetUnit(ctx contractapi.TransactionContextInterface, id string) (*Unit, error) {
	if err := s.HasPermission(ctx, UnitsRead); err != nil {
		return nil, err
	}

	assetBytes, err := ctx.GetStub().GetState(s.GetUnitID(ctx, id))
	if err != nil {
		return nil, fmt.Errorf("failed to get asset %s: %v", id, err)
	}

	if assetBytes == nil {
		return nil, fmt.Errorf("asset %s does not exist", id)
	}

	var unit UnitInner
	err = json.Unmarshal(assetBytes, &unit)
	if err != nil {
		return nil, err
	}

	unit.ID = strings.TrimPrefix(unit.ID, string(UnitDoc)+"_")
	return FromUnitInner(&unit), nil
}

func (s *SmartContract) GetAllUnits(ctx contractapi.TransactionContextInterface) ([]*Unit, error) {
	if err := s.HasPermission(ctx, UnitsRead); err != nil {
		return nil, err
	}

	results, err := ctx.GetStub().GetQueryResult(fmt.Sprintf(`{"selector":{"doc_type":"%s"}}`, UnitDoc))
	if err != nil {
		return nil, fmt.Errorf("failed to get assets: %v", err)
	}

	defer results.Close()

	var assets []*Unit
	for results.HasNext() {
		queryResult, err := results.Next()
		if err != nil {
			return nil, err
		}
		var unit UnitInner
		err = json.Unmarshal(queryResult.Value, &unit)
		if err != nil {
			return nil, err
		}

		unit.ID = strings.TrimPrefix(unit.ID, string(UnitDoc)+"_")
		assets = append(assets, FromUnitInner(&unit))
	}

	return assets, nil
}

func (s *SmartContract) DeleteUnit(ctx contractapi.TransactionContextInterface, id string) error {
	if err := s.HasPermission(ctx, UnitsDelete); err != nil {
		return err
	}

	err := ctx.GetStub().DelState(s.GetUnitID(ctx, id))
	if err != nil {
		return fmt.Errorf("failed to delete asset %s: %v", id, err)
	}

	return nil
}
