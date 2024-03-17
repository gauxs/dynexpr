// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package data

import ()

type Person_ExpressionBuilder struct {
	PK            DynamoKeyAttribute[*string]
	SK            DynamoKeyAttribute[*string]
	Name          DynamoAttribute[*string]
	BankDetailsss DynamoAttribute[*data.BankDetails]
	FamilyDetails DynamoAttribute[*data.FamilyDetail]
	PhoneNos      DynamoListAttribute[*string]
}

func (o *Person_ExpressionBuilder) BuildTree(name string) *DynamoAttribute[*Person_ExpressionBuilder] {
	o = &Person_ExpressionBuilder{}
	o.PK = *NewDynamoKeyAttribute[*string]().WithName("pk")
	o.SK = *NewDynamoKeyAttribute[*string]().WithName("sk")
	o.Name = *NewDynamoAttribute[*string]().WithName("name")
	o.BankDetailsss = *(&data.BankDetails_ExpressionBuilder{}).BuildTree("bank_details")
	o.FamilyDetails = *(&data.FamilyDetail_ExpressionBuilder{}).BuildTree("family_details")
	o.PhoneNos = *NewDynamoListAttribute[*string]().WithName("phone_nos")
	return NewDynamoAttribute[*Person_ExpressionBuilder]().
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
	Children  DynamoListAttribute[*data.Child]
	IsMarried DynamoAttribute[*bool]
}

func (o *FamilyDetail_ExpressionBuilder) BuildTree(name string) *DynamoAttribute[*FamilyDetail_ExpressionBuilder] {
	o = &FamilyDetail_ExpressionBuilder{}
	o.Children = *NewDynamoListAttribute[*data.Child]().WithName("children")
	o.IsMarried = *NewDynamoAttribute[*bool]().WithName("is_married")
	return NewDynamoAttribute[*FamilyDetail_ExpressionBuilder]().
		WithAccessReference(o).
		WithName(name).
		WithChildAttribute(&o.Children).
		WithChildAttribute(&o.IsMarried)
}

type Child_ExpressionBuilder struct {
	Name DynamoAttribute[*string]
	DOB  DynamoAttribute[*time.Time]
}

func (o *Child_ExpressionBuilder) BuildTree(name string) *DynamoAttribute[*Child_ExpressionBuilder] {
	o = &Child_ExpressionBuilder{}
	o.Name = *NewDynamoAttribute[*string]().WithName("name")
	o.DOB = *(&time.Time_ExpressionBuilder{}).BuildTree("dob")
	return NewDynamoAttribute[*Child_ExpressionBuilder]().
		WithAccessReference(o).
		WithName(name).
		WithChildAttribute(&o.Name).
		WithChildAttribute(&o.DOB)
}

type BankDetails_ExpressionBuilder struct {
	Accounts DynamoAttribute[*data.BankAccount]
}

func (o *BankDetails_ExpressionBuilder) BuildTree(name string) *DynamoAttribute[*BankDetails_ExpressionBuilder] {
	o = &BankDetails_ExpressionBuilder{}
	o.Accounts = *(&data.BankAccount_ExpressionBuilder{}).BuildTree("accounts")
	return NewDynamoAttribute[*BankDetails_ExpressionBuilder]().
		WithAccessReference(o).
		WithName(name).
		WithChildAttribute(&o.Accounts)
}

type BankAccount_ExpressionBuilder struct {
	OutsidePkg DynamoAttribute[net.AddrError]
}

func (o *BankAccount_ExpressionBuilder) BuildTree(name string) *DynamoAttribute[*BankAccount_ExpressionBuilder] {
	o = &BankAccount_ExpressionBuilder{}
	o.OutsidePkg = *(&net.AddrError_ExpressionBuilder{}).BuildTree("")
	return NewDynamoAttribute[*BankAccount_ExpressionBuilder]().
		WithAccessReference(o).
		WithName(name).
		WithChildAttribute(&o.OutsidePkg)
}
