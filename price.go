package main

//Structure for storing monetary values
//Amount represents how much money
//Exponent represents how many decimals
//Currency represents the type of currency (EUR,USD,...)
type Price struct {
	Amount   uint32 `json:"amount"`
	Exponent uint32 `json:"exponent"`
	Currency string `json:"currency"`
}
