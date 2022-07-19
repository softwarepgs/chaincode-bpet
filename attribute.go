package main

import (
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const (
	UnitsCreate Attribute = "units.create"
	UnitsRead   Attribute = "units.read"
	UnitsUpdate Attribute = "units.update"
	UnitsDelete Attribute = "units.delete"

	ProductsCreate Attribute = "products.create"
	ProductsRead   Attribute = "products.read"
	ProductsUpdate Attribute = "products.update"
	ProductsDelete Attribute = "products.delete"

	OrganizationsCreate Attribute = "organizations.create"
	OrganizationsRead   Attribute = "organizations.read"
	OrganizationsUpdate Attribute = "organizations.update"
	OrganizationsDelete Attribute = "organizations.delete"

	OrdersCreate Attribute = "orders.create"
	OrdersRead   Attribute = "orders.read"
	OrdersUpdate Attribute = "orders.update"
	OrdersDelete Attribute = "orders.delete"

	TransactionsCreate Attribute = "transactions.create"
	TransactionsRead   Attribute = "transactions.read"
	TransactionsUpdate Attribute = "transactions.update"
	TransactionsDelete Attribute = "transactions.delete"

	RequestsCreate Attribute = "requests.create"
	RequestsRead   Attribute = "requests.read"
	RequestsUpdate Attribute = "requests.update"
	RequestsDelete Attribute = "requests.delete"

	OffersCreate Attribute = "offers.create"
	OffersRead   Attribute = "offers.read"
	OffersUpdate Attribute = "offers.update"
	OffersDelete Attribute = "offers.delete"
)

type Attribute string

func (a Attribute) String() string {
	return string(a)
}

func (s *SmartContract) HasPermission(ctx contractapi.TransactionContextInterface, att Attribute) error {
	if !s.checkPermissions {
		return nil
	}

	err := ctx.GetClientIdentity().AssertAttributeValue(att.String(), "true")
	if err != nil {
		return fmt.Errorf(" not authorized ")
	}

	return nil
}
