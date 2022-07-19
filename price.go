package main

type Price struct {
	Amount   uint32 `json:"amount"`
	Exponent uint32 `json:"exponent"`
	Currency string `json:"currency"`
}
