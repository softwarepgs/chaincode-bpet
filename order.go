package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const (
	OrderDoc DocType = "order"
)

//Represents data stored in database
//Contains the doctype
type OrderInner struct {
	Doc

	ID             string      `json:"id"`
	Amount         uint32      `json:"amount"`
	Price          Price       `json:"price"`
	Type           OrderType   `json:"type"`
	Status         OrderStatus `json:"status"`
	OrganizationID string      `json:"organization_id"`
	ProductID      string      `json:"product_id"`
	UnitID         string      `json:"unit_id"`
}

type Order struct {
	ID           string        `json:"id"`
	Amount       uint32        `json:"amount"`
	Price        Price         `json:"price"`
	Type         OrderType     `json:"type"`
	Status       OrderStatus   `json:"status"`
	Organization *Organization `json:"organization_id"`
	Product      *Product      `json:"product_id"`
	Unit         *Unit         `json:"unit_id"`
}

//Parse order from the data on the database
func (s *SmartContract) FromOrderInner(ctx contractapi.TransactionContextInterface, p *OrderInner) *Order {
	org, _ := s.GetOrganization(ctx, p.OrganizationID)
	product, _ := s.GetProduct(ctx, p.ProductID)
	unit, _ := s.GetUnit(ctx, p.UnitID)

	return &Order{
		ID:     p.ID,
		Amount: p.Amount,
		Price: Price{
			Amount:   p.Price.Amount,
			Exponent: p.Price.Exponent,
			Currency: p.Price.Currency,
		},
		Type:         p.Type,
		Status:       p.Status,
		Organization: org,
		Product:      product,
		Unit:         unit,
	}
}

func (s *SmartContract) GetOrderID(ctx contractapi.TransactionContextInterface, id string) string {
	return string(OrderDoc) + "_" + id
}

//Checks if order with the given ID exists
func (s *SmartContract) OrderExist(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(s.GetOrderID(ctx, id))
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

//Checks if list of orders with the given IDs exists
func (s *SmartContract) OrdersExist(ctx contractapi.TransactionContextInterface, ids []string) error {
	for _, id := range ids {
		e, err := s.OrderExist(ctx, id)
		if err != nil {
			return err
		}

		if !e {
			return fmt.Errorf("order %s does not exist", id)
		}
	}

	return nil
}

//Creates a new order with the given ID
//User inputs the ID of the offer, the amount of product being sold, the total value of money, the exponent (number of decimals), the currency, the type of Order (BUY or SELL), the ID of the organization, the ID of the product and the ID of the unit
func (s *SmartContract) CreateOrder(ctx contractapi.TransactionContextInterface, id string, amount uint32, price uint32, priceExponent uint32, currency string, typeInput string, organizationID string, productID string, unitID string) error {
	if err := s.HasPermission(ctx, OrdersCreate); err != nil {
		return err
	}

	orgID, err := s.GetSubmittingClientOrganization(ctx)
	if err != nil || orgID != organizationID {
		return fmt.Errorf("unauthorized")
	}

	exists, err := s.OrderExist(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	hasOrg, err := s.OrganizationExist(ctx, productID)
	if err != nil {
		return err
	}
	if !hasOrg {
		return fmt.Errorf("organization %s does not exist", id)
	}

	hasProduct, err := s.ProductExist(ctx, productID)
	if err != nil {
		return err
	}
	if !hasProduct {
		return fmt.Errorf("product %s does not exist", id)
	}

	hasUnit, err := s.UnitExist(ctx, unitID)
	if err != nil {
		return err
	}
	if !hasUnit {
		return fmt.Errorf("unit %s does not exist", id)
	}

	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	_type, err := ParseOrderType(typeInput)
	if err != nil {
		return err
	}

	doc := Doc{
		Type:      ProductDoc,
		CreatedBy: clientID,
		UpdatedBy: clientID,
	}

	unit := OrderInner{
		Doc:    doc,
		ID:     s.GetOrderID(ctx, id),
		Amount: amount,
		Price: Price{
			Amount:   price,
			Exponent: priceExponent,
			Currency: currency,
		},
		Type:           _type,
		Status:         OrderStatusOpen,
		OrganizationID: organizationID,
		ProductID:      productID,
		UnitID:         unitID,
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

//Changes status of order to "CLOSED"
func (s *SmartContract) CloseOrder(ctx contractapi.TransactionContextInterface, id string) error {
	if err := s.HasPermission(ctx, OrdersUpdate); err != nil {
		return err
	}

	exists, err := s.OrderExist(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	order, err := s.GetOrderInner(ctx, id)
	if err != nil {
		return err
	}

	orgID, err := s.GetSubmittingClientOrganization(ctx)
	if err != nil || orgID != order.OrganizationID {
		return fmt.Errorf("unauthorized")
	}

	if order.Status == OrderStatusClosed {
		return fmt.Errorf("you do not belong in that org")
	}

	order.Status = OrderStatusClosed

	assetBytes, err := json.Marshal(order)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(order.ID, assetBytes)

	return nil
}

//Returns Order with given ID
func (s *SmartContract) GetOrder(ctx contractapi.TransactionContextInterface, id string) (*Order, error) {
	if err := s.HasPermission(ctx, OrdersRead); err != nil {
		return nil, err
	}

	assetBytes, err := ctx.GetStub().GetState(s.GetOrderID(ctx, id))
	if err != nil {
		return nil, fmt.Errorf("failed to get asset %s:%v", id, err)
	}

	if assetBytes == nil {
		return nil, fmt.Errorf("asset %s does not exist", id)
	}

	var unit OrderInner
	err = json.Unmarshal(assetBytes, &unit)
	if err != nil {
		return nil, err
	}

	unit.ID = strings.TrimPrefix(unit.ID, string(OrderDoc)+"_")
	return s.FromOrderInner(ctx, &unit), nil
}

//Returns OrderInner with given ID
func (s *SmartContract) GetOrderInner(ctx contractapi.TransactionContextInterface, id string) (*OrderInner, error) {
	if err := s.HasPermission(ctx, OrdersRead); err != nil {
		return nil, err
	}

	assetBytes, err := ctx.GetStub().GetState(s.GetOrderID(ctx, id))
	if err != nil {
		return nil, fmt.Errorf("failed to get asset %s:%v", id, err)
	}

	if assetBytes == nil {
		return nil, fmt.Errorf("asset %s does not exist", id)
	}

	var unit OrderInner
	err = json.Unmarshal(assetBytes, &unit)
	if err != nil {
		return nil, err
	}

	unit.ID = strings.TrimPrefix(unit.ID, string(OrderDoc)+"_")
	return &unit, nil
}

//Returns all Order in the system
func (s *SmartContract) GetAllOrders(ctx contractapi.TransactionContextInterface) ([]*Order, error) {
	if err := s.HasPermission(ctx, OrdersRead); err != nil {
		return nil, err
	}

	results, err := ctx.GetStub().GetQueryResult(fmt.Sprintf(`{"selector":{"doc_type":"%s"}}`, OrderDoc))
	if err != nil {
		return nil, fmt.Errorf("failed to get assets:%v", err)
	}
	defer results.Close()

	var assets []*Order
	for results.HasNext() {
		queryResult, err := results.Next()
		if err != nil {
			return nil, err
		}
		var unit OrderInner
		err = json.Unmarshal(queryResult.Value, &unit)
		if err != nil {
			return nil, err
		}

		unit.ID = strings.TrimPrefix(unit.ID, string(OrderDoc)+"_")
		assets = append(assets, s.FromOrderInner(ctx, &unit))
	}
	return assets, nil
}

//Returns all Order with the given status
func (s *SmartContract) GetAllOrdersByStatus(ctx contractapi.TransactionContextInterface, statusInput string) ([]*Order, error) {
	if err := s.HasPermission(ctx, OrdersRead); err != nil {
		return nil, err
	}

	status, err := ParseOrderStatus(statusInput)
	if err != nil {
		return nil, err
	}

	results, err := ctx.GetStub().GetQueryResult(fmt.Sprintf(`{"selector":{"doc_type":"%s","status":"%s"}}`, OrderDoc, status))
	if err != nil {
		return nil, fmt.Errorf("failed to get assets:%v", err)
	}
	defer results.Close()

	var assets []*Order
	for results.HasNext() {
		queryResult, err := results.Next()
		if err != nil {
			return nil, err
		}
		var unit OrderInner
		err = json.Unmarshal(queryResult.Value, &unit)
		if err != nil {
			return nil, err
		}

		unit.ID = strings.TrimPrefix(unit.ID, string(OrderDoc)+"_")
		assets = append(assets, s.FromOrderInner(ctx, &unit))
	}

	return assets, nil
}

//Returns all Order associated to the organization with the given ID
func (s *SmartContract) GetAllOrdersByOrganization(ctx contractapi.TransactionContextInterface, org string) ([]*Order, error) {
	if err := s.HasPermission(ctx, OrdersRead); err != nil {
		return nil, err
	}

	results, err := ctx.GetStub().GetQueryResult(fmt.Sprintf(`{"selector":{"doc_type":"%s","organization_id":"%s"}}`, OrderDoc, org))
	if err != nil {
		return nil, fmt.Errorf("failed to get assets:%v", err)
	}
	defer results.Close()

	var assets []*Order
	for results.HasNext() {
		queryResult, err := results.Next()
		if err != nil {
			return nil, err
		}
		var unit OrderInner
		err = json.Unmarshal(queryResult.Value, &unit)
		if err != nil {
			return nil, err
		}

		unit.ID = strings.TrimPrefix(unit.ID, string(OrderDoc)+"_")
		assets = append(assets, s.FromOrderInner(ctx, &unit))
	}

	return assets, nil
}

//Returns all Order associated to the organization with the given ID and the given status
func (s *SmartContract) GetAllOrdersByOrganizationAndStatus(ctx contractapi.TransactionContextInterface, org, statusInput string) ([]*Order, error) {
	if err := s.HasPermission(ctx, OrdersRead); err != nil {
		return nil, err
	}

	status, err := ParseOrderStatus(statusInput)
	if err != nil {
		return nil, err
	}

	results, err := ctx.GetStub().GetQueryResult(fmt.Sprintf(`{"selector":{"doc_type":"%s","organization_id":"%s","status":"%s"}}`, OrderDoc, org, status))
	if err != nil {
		return nil, fmt.Errorf("failed to get assets:%v", err)
	}
	defer results.Close()

	var assets []*Order
	for results.HasNext() {
		queryResult, err := results.Next()
		if err != nil {
			return nil, err
		}
		var unit OrderInner
		err = json.Unmarshal(queryResult.Value, &unit)
		if err != nil {
			return nil, err
		}

		unit.ID = strings.TrimPrefix(unit.ID, string(OrderDoc)+"_")
		assets = append(assets, s.FromOrderInner(ctx, &unit))
	}

	return assets, nil
}
