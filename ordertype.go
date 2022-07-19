package main

import (
	"fmt"
)

const (
	OrderTypeBuy  OrderType = "BUY"
	OrderTypeSell OrderType = "SELL"
)

type OrderType string

func (o OrderType) String() string {
	return string(o)
}

func ParseOrderType(_type string) (OrderType, error) {
	switch _type {
	case "BUY":
		return OrderTypeBuy, nil
	case "SELL":
		return OrderTypeSell, nil
	}

	return "", fmt.Errorf("invalid order type")
}
