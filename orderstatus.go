package main

import (
	"fmt"
)

const (
	OrderStatusOpen   OrderStatus = "OPEN"
	OrderStatusClosed OrderStatus = "CLOSED"
)

type OrderStatus string

func (o OrderStatus) String() string {
	return string(o)
}

func ParseOrderStatus(status string) (OrderStatus, error) {
	switch status {
	case "OPEN":
		return OrderStatusOpen, nil
	case "CLOSED":
		return OrderStatusClosed, nil
	}

	return "", fmt.Errorf("invalid order status")
}
