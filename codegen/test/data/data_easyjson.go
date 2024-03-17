package data

import (
	dynexpr "dynexpr/v1"
	net "net"
	time "time"
)

type Person_ExpressionBuilder struct {
	PK            dynexpr.DynamoKeyAttribute[*string]
	SK            dynexpr.DynamoKeyAttribute[*string]
	Name          dynexpr.DynamoAttribute[*string]
	BankDetailsss dynexpr.DynamoAttribute[*BankDetails_ExpressionBuilder]
	FamilyDetails dynexpr.DynamoAttribute[*FamilyDetail_ExpressionBuilder]
	PhoneNos      dynexpr.DynamoListAttribute[*string]
}

func (o *Person_ExpressionBuilder) BuildTree(name string) *dynexpr.DynamoAttribute[*Person_ExpressionBuilder] {
	o = &Person_ExpressionBuilder{}
	o.PK = *dynexpr.NewDynamoKeyAttribute[*string]().WithName("pk")
	o.SK = *dynexpr.NewDynamoKeyAttribute[*string]().WithName("sk")
	o.Name = *dynexpr.NewDynamoAttribute[*string]().WithName("name")
	o.BankDetailsss = *(&BankDetails_ExpressionBuilder{}).BuildTree("bank_details")
	o.FamilyDetails = *(&FamilyDetail_ExpressionBuilder{}).BuildTree("family_details")
	o.PhoneNos = *dynexpr.NewDynamoListAttribute[*string]().WithName("phone_nos")
	return dynexpr.NewDynamoAttribute[*Person_ExpressionBuilder]().
		WithAccessReference(o).
		WithName(name).
		WithChildAttribute(&o.PK).
		WithChildAttribute(&o.SK).
		WithChildAttribute(&o.Name).
		WithChildAttribute(&o.BankDetailsss).
		WithChildAttribute(&o.FamilyDetails).
		WithChildAttribute(&o.PhoneNos)
}

type FamilyDetail_ExpressionBuilder struct {
	Children  dynexpr.DynamoListAttribute[*Child]
	IsMarried dynexpr.DynamoAttribute[*bool]
}

func (o *FamilyDetail_ExpressionBuilder) BuildTree(name string) *dynexpr.DynamoAttribute[*FamilyDetail_ExpressionBuilder] {
	o = &FamilyDetail_ExpressionBuilder{}
	o.Children = *dynexpr.NewDynamoListAttribute[*Child]().WithName("children")
	o.IsMarried = *dynexpr.NewDynamoAttribute[*bool]().WithName("is_married")
	return dynexpr.NewDynamoAttribute[*FamilyDetail_ExpressionBuilder]().
		WithAccessReference(o).
		WithName(name).
		WithChildAttribute(&o.Children).
		WithChildAttribute(&o.IsMarried)
}

type Child_ExpressionBuilder struct {
	Name dynexpr.DynamoAttribute[*string]
	DOB  dynexpr.DynamoAttribute[*time.Time_ExpressionBuilder]
}

func (o *Child_ExpressionBuilder) BuildTree(name string) *dynexpr.DynamoAttribute[*Child_ExpressionBuilder] {
	o = &Child_ExpressionBuilder{}
	o.Name = *dynexpr.NewDynamoAttribute[*string]().WithName("name")
	o.DOB = *(&time.Time_ExpressionBuilder{}).BuildTree("dob")
	return dynexpr.NewDynamoAttribute[*Child_ExpressionBuilder]().
		WithAccessReference(o).
		WithName(name).
		WithChildAttribute(&o.Name).
		WithChildAttribute(&o.DOB)
}

type BankDetails_ExpressionBuilder struct {
	Accounts dynexpr.DynamoAttribute[*BankAccount_ExpressionBuilder]
}

func (o *BankDetails_ExpressionBuilder) BuildTree(name string) *dynexpr.DynamoAttribute[*BankDetails_ExpressionBuilder] {
	o = &BankDetails_ExpressionBuilder{}
	o.Accounts = *(&BankAccount_ExpressionBuilder{}).BuildTree("accounts")
	return dynexpr.NewDynamoAttribute[*BankDetails_ExpressionBuilder]().
		WithAccessReference(o).
		WithName(name).
		WithChildAttribute(&o.Accounts)
}

type BankAccount_ExpressionBuilder struct {
	OutsidePkg dynexpr.DynamoAttribute[net.AddrError_ExpressionBuilder]
}

func (o *BankAccount_ExpressionBuilder) BuildTree(name string) *dynexpr.DynamoAttribute[*BankAccount_ExpressionBuilder] {
	o = &BankAccount_ExpressionBuilder{}
	o.OutsidePkg = *(&net.AddrError_ExpressionBuilder{}).BuildTree("")
	return dynexpr.NewDynamoAttribute[*BankAccount_ExpressionBuilder]().
		WithAccessReference(o).
		WithName(name).
		WithChildAttribute(&o.OutsidePkg)
}
