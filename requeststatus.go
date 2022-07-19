package main

import (
	"fmt"
)

const (
	RequestStatusOpen   RequestStatus = "OPEN"
	RequestStatusClosed RequestStatus = "CLOSED"
)

type RequestStatus string

func (r RequestStatus) String() string {
	return string(r)
}

func ParseRequestStatus(status string) (RequestStatus, error) {
	switch status {
	case "OPEN":
		return RequestStatusOpen, nil
	case "CLOSED":
		return RequestStatusClosed, nil
	}

	return "", fmt.Errorf("invalid request status")
}
