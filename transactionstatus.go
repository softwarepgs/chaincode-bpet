package main

import (
	"fmt"
)

const (
	TransactionStatusOpen           TransactionStatus = "OPEN"
	TransactionStatusClosed         TransactionStatus = "CLOSED"
	TransactionStatusCanceled       TransactionStatus = "CANCELED"
	TransactionStatusInReview       TransactionStatus = "IN_REVIEW"
	TransactionStatusWaitingPayment TransactionStatus = "WAITING_PAYMENT"
	TransactionStatusPaid           TransactionStatus = "PAID"
	TransactionStatusReady          TransactionStatus = "READY"
	TransactionStatusInProgress     TransactionStatus = "IN_PROGRESS"
	TransactionStatusNotDelivered   TransactionStatus = "NOT_DELIVERED"
	TransactionStatusDelivered      TransactionStatus = "DELIVERED"
)

type TransactionStatus string

func (o TransactionStatus) String() string {
	return string(o)
}

func ParseTransactionStatus(status string) (TransactionStatus, error) {
	switch status {
	case "OPEN":
		return TransactionStatusOpen, nil
	case "CLOSED":
		return TransactionStatusClosed, nil
	case "CANCELED":
		return TransactionStatusCanceled, nil
	case "WAITING_PAYMENT":
		return TransactionStatusWaitingPayment, nil
	case "PAID":
		return TransactionStatusPaid, nil
	case "IN_REVIEW":
		return TransactionStatusInReview, nil
	case "READY":
		return TransactionStatusReady, nil
	case "IN_PROGRESS":
		return TransactionStatusInProgress, nil
	case "NOT_DELIVERED":
		return TransactionStatusNotDelivered, nil
	case "DELIVERED":
		return TransactionStatusDelivered, nil
	}
	return " ", fmt.Errorf("invalid transaction type")
}
