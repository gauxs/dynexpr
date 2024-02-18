package data

import (
	"database/sql"
	"time"
)

type BankAccountType int

const (
	SAVINGS BankAccountType = iota
	CURRENT
)

type BankAccount struct {
	BankAccountNumber *int             `json:"bank_account_number,omitempty" dynamodbav:"bank_account_number,omitempty"`
	AccountType       *BankAccountType `json:"account_type,omitempty" dynamodbav:"account_type,omitempty"`
	OutsidePkg        sql.DBStats
}

type BankDetails struct {
	Accounts *[]*BankAccount `json:"accounts,omitempty" dynamodbav:"accounts,numberset,omitempty,omitemptyelem"` // This will be a set in DDB
}

type Child struct {
	Name *string    `json:"name,omitempty" dynamodbav:"name,omitempty"`
	DOB  *time.Time `json:"dob,omitempty" dynamodbav:"dob,omitempty"`
}

type FamilyDetail struct {
	Children  *[]*Child `json:"children,omitempty" dynamodbav:"children,omitempty,omitemptyelem"` // This will be a list
	IsMarried *bool     `json:"is_married,omitempty" dynamodbav:"is_married,omitempty"`
}

type Person struct {
	PK            *string       `json:"pk,omitempty" dynamodbav:"pk,omitempty"`
	SK            *string       `json:"sk,omitempty" dynamodbav:"sk,omitempty"`
	Name          *string       `json:"name,omitempty" dynamodbav:"name,omitempty"`
	BankDetails   *BankDetails  `json:"bank_details,omitempty" dynamodbav:"bank_details,omitempty"`
	FamilyDetails *FamilyDetail `json:"family_details,omitempty" dynamodbav:"family_details,omitempty"`
	PhoneNos      *[]*string    `json:"phone_nos,omitempty" dynamodbav:"phone_nos,omitempty,omitemptyelem"` // This will be a list
}
