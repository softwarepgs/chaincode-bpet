package main

type DocType string

type Doc struct {
	Type      DocType `json:"doc_type"`
	CreatedBy string  `json:"created_by"`
	UpdatedBy string  `json:"updated_by"`
}
