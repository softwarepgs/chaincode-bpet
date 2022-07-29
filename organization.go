package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const (
	OrganizationDoc DocType = "organization"
)

//Represents data stored in database
//Contains the doctype
type OrganizationInner struct {
	Doc

	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
	PhoneNumber string `json:"phone_number"`
}

type Organization struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
	PhoneNumber string `json:"phone_number"`
}

//Parse organization from the data on the database
func (s *SmartContract) FromOrganizationInner(_ contractapi.TransactionContextInterface, p *OrganizationInner) *Organization {
	return &Organization{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Address:     p.Address,
		PhoneNumber: p.PhoneNumber,
	}
}

func (s *SmartContract) GetOrganizationID(_ contractapi.TransactionContextInterface, id string) string {
	return string(UnitDoc) + "_" + id
}

//Checks if organization with the given ID exists
func (s *SmartContract) OrganizationExist(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(s.GetOrganizationID(ctx, id))
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

//Creates a new organization with the given ID
//User inputs the ID of the organization, the name of the organization, a description of the organization, the address of the organization and the phone number
func (s *SmartContract) CreateOrganization(ctx contractapi.TransactionContextInterface, id string, name string, description string, address string, phoneNumber string) error {
	if err := s.HasPermission(ctx, OrganizationsCreate); err != nil {
		return err
	}

	exists, err := s.OrganizationExist(ctx, id)
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

	unit := OrganizationInner{
		ID:          s.GetOrganizationID(ctx, id),
		Name:        name,
		Description: description,
		Address:     address,
		PhoneNumber: phoneNumber,
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

//Updates information regarding the organization
//Updates the name, description, address and phone number of the organization with the given ID
func (s *SmartContract) UpdateOrganization(ctx contractapi.TransactionContextInterface, id string, name string, description string, address string, phoneNumber string) error {
	if err := s.HasPermission(ctx, OrganizationsCreate); err != nil {
		if innerErr := s.HasPermission(ctx, OrganizationsUpdate); innerErr != nil {
			return innerErr
		}

		if orgID, innerErr := s.GetSubmittingClientOrganization(ctx); innerErr != nil && orgID != id {
			return err
		}
	}

	exists, err := s.OrganizationExist(ctx, id)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	org, err := s.GetOrganizationInner(ctx, id)
	if err != nil {
		return err
	}

	org.Name = name
	org.Description = description
	org.Address = address
	org.PhoneNumber = phoneNumber
	org.UpdatedBy = clientID

	assetBytes, err := json.Marshal(org)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(org.ID, assetBytes)
	if err != nil {
		return err
	}

	return nil
}

//Returns OrganizationInner with the given ID
func (s *SmartContract) GetOrganizationInner(ctx contractapi.TransactionContextInterface, id string) (*OrganizationInner, error) {
	if err := s.HasPermission(ctx, OrganizationsRead); err != nil {
		return nil, err
	}

	assetBytes, err := ctx.GetStub().GetState(s.GetOrganizationID(ctx, id))
	if err != nil {
		return nil, fmt.Errorf("failed to get asset %s: %v", id, err)
	}

	if assetBytes == nil {
		return nil, fmt.Errorf("asset %s does not exist", id)
	}

	var product OrganizationInner
	err = json.Unmarshal(assetBytes, &product)
	if err != nil {
		return nil, err
	}

	product.ID = strings.TrimPrefix(product.ID, string(OrganizationDoc)+"_")
	return &product, nil
}

//Returns Organization with the given ID
func (s *SmartContract) GetOrganization(ctx contractapi.TransactionContextInterface, id string) (*Organization, error) {
	if err := s.HasPermission(ctx, OrganizationsRead); err != nil {
		return nil, err
	}

	assetBytes, err := ctx.GetStub().GetState(s.GetOrganizationID(ctx, id))
	if err != nil {
		return nil, fmt.Errorf("failed to get asset %s: %v", id, err)
	}

	if assetBytes == nil {
		return nil, fmt.Errorf("asset %s does not exist", id)
	}

	var product OrganizationInner
	err = json.Unmarshal(assetBytes, &product)
	if err != nil {
		return nil, err
	}

	product.ID = strings.TrimPrefix(product.ID, string(OrganizationDoc)+"_")
	return s.FromOrganizationInner(ctx, &product), nil
}

//Returns all organizations in the system
func (s *SmartContract) GetAllOrganizations(ctx contractapi.TransactionContextInterface) ([]*Organization, error) {
	if err := s.HasPermission(ctx, OrganizationsRead); err != nil {
		return nil, err
	}

	results, err := ctx.GetStub().GetQueryResult(fmt.Sprintf(`{"selector":{"doc_type":"%s"}}`, OrganizationDoc))
	if err != nil {
		return nil, fmt.Errorf("failed to get assets: %v", err)
	}

	defer results.Close()

	var assets []*Organization
	for results.HasNext() {
		queryResult, err := results.Next()
		if err != nil {
			return nil, err
		}
		var unit OrganizationInner
		err = json.Unmarshal(queryResult.Value, &unit)
		if err != nil {
			return nil, err
		}

		unit.ID = strings.TrimPrefix(unit.ID, string(OrganizationDoc)+"_")
		assets = append(assets, s.FromOrganizationInner(ctx, &unit))
	}

	return assets, nil
}

//Deletes the organization from the system
func (s *SmartContract) DeleteOrganization(ctx contractapi.TransactionContextInterface, id string) error {
	if err := s.HasPermission(ctx, OrganizationsDelete); err != nil {
		return err
	}

	err := ctx.GetStub().DelState(s.GetOrganizationID(ctx, id))
	if err != nil {
		return fmt.Errorf("failed to delete asset %s: %v", id, err)
	}

	return nil
}
