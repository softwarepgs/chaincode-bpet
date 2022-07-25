package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const (
	ProductDoc DocType = "product"
)

type ProductInner struct {
	Doc

	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	UnitIDs     []string `json:"unit_ids"`
}

type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Units       []*Unit `json:"units"`
}

func (s *SmartContract) FromProductInner(ctx contractapi.TransactionContextInterface, p *ProductInner) *Product {
	units := make([]*Unit, 0, len(p.UnitIDs))
	for _, unitID := range p.UnitIDs {
		unit, err := s.GetUnit(ctx, unitID)
		if err != nil {
			continue
		}

		units = append(units, unit)
	}

	return &Product{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Units:       units,
	}
}

func (s *SmartContract) GetProductID(_ contractapi.TransactionContextInterface, id string) string {
	return string(ProductDoc) + "_" + id
}

func (s *SmartContract) ProductExist(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(s.GetProductID(ctx, id))
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

func (s *SmartContract) CreateProduct(ctx contractapi.TransactionContextInterface, id, name string, description string, units []string) error {
	if err := s.HasPermission(ctx, ProductsCreate); err != nil {
		return err
	}

	exists, err := s.ProductExist(ctx, id)
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	if err := s.UnitsExist(ctx, units); err != nil {
		return err
	}

	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	doc := Doc{
		Type:      ProductDoc,
		CreatedBy: clientID,
		UpdatedBy: clientID,
	}

	unit := ProductInner{
		ID:          s.GetProductID(ctx, id),
		Name:        name,
		Description: description,
		UnitIDs:     units,
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

func (s *SmartContract) GetProductInner(ctx contractapi.TransactionContextInterface, id string) (*ProductInner, error) {
	if err := s.HasPermission(ctx, ProductsRead); err != nil {
		return nil, err
	}

	assetBytes, err := ctx.GetStub().GetState(s.GetProductID(ctx, id))
	if err != nil {
		return nil, fmt.Errorf("failed to get asset %s: %v", id, err)
	}

	if assetBytes == nil {
		return nil, fmt.Errorf("asset %s does not exist", id)
	}

	var product ProductInner
	err = json.Unmarshal(assetBytes, &product)
	if err != nil {
		return nil, err
	}

	product.ID = strings.TrimPrefix(product.ID, string(ProductDoc)+"_")
	return &product, nil
}

func (s *SmartContract) GetProduct(ctx contractapi.TransactionContextInterface, id string) (*Product, error) {
	if err := s.HasPermission(ctx, ProductsRead); err != nil {
		return nil, err
	}

	assetBytes, err := ctx.GetStub().GetState(s.GetProductID(ctx, id))
	if err != nil {
		return nil, fmt.Errorf("failed to get asset %s: %v", id, err)
	}

	if assetBytes == nil {
		return nil, fmt.Errorf("asset %s does not exist", id)
	}

	var product ProductInner
	err = json.Unmarshal(assetBytes, &product)
	if err != nil {
		return nil, err
	}

	product.ID = strings.TrimPrefix(product.ID, string(ProductDoc)+"_")
	return s.FromProductInner(ctx, &product), nil
}

func (s *SmartContract) GetAllProducts(ctx contractapi.TransactionContextInterface) ([]*Product, error) {
	if err := s.HasPermission(ctx, ProductsRead); err != nil {
		return nil, err
	}

	results, err := ctx.GetStub().GetQueryResult(fmt.Sprintf(`{"selector":{"doc_type":"%s"}}`, ProductDoc))
	if err != nil {
		return nil, fmt.Errorf("failed to get assets: %v", err)
	}

	defer results.Close()

	var assets []*Product
	for results.HasNext() {
		queryResult, err := results.Next()
		if err != nil {
			return nil, err
		}
		var unit ProductInner
		err = json.Unmarshal(queryResult.Value, &unit)
		if err != nil {
			return nil, err
		}

		unit.ID = strings.TrimPrefix(unit.ID, string(ProductDoc)+"_")
		assets = append(assets, s.FromProductInner(ctx, &unit))
	}

	return assets, nil
}

func (s *SmartContract) DeleteProduct(ctx contractapi.TransactionContextInterface, id string) error {
	if err := s.HasPermission(ctx, ProductsDelete); err != nil {
		return err
	}

	err := ctx.GetStub().DelState(s.GetProductID(ctx, id))
	if err != nil {
		return fmt.Errorf("failed to delete asset %s: %v", id, err)
	}

	return nil
}
