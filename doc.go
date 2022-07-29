package main

type DocType string

//Helper structure for couchDB
//DocType represents the document type - making it easier to search
//CreatedBy stores the ID of the user that created the document
//UpdatedBy stores the ID of the user that updated the document
type Doc struct {
	Type      DocType `json:"doc_type"`
	CreatedBy string  `json:"created_by"`
	UpdatedBy string  `json:"updated_by"`
}
