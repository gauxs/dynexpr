package v1

import (
	"dynexpr/utils"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/stretchr/testify/assert"
)

// --------------------------------------------- PERSON MODEL ---------------------------------------------
type BankAccountType int

const (
	SAVINGS BankAccountType = iota
	CURRENT
)

type BankAccount struct {
	BankAccountNumber *int             `json:"bank_account_number,omitempty" dynamodbav:"bank_account_number,omitempty"`
	AccountType       *BankAccountType `json:"account_type,omitempty" dynamodbav:"account_type,omitempty"`
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

// ---------------------------------------- EXPRESSION BUILDER FOR PERSON MODEL ----------------------------------------
type BankAccount_ExpressionBuilder struct {
	BankAccountNumber DynamoAttribute[*int]
	AccountType       DynamoAttribute[*BankAccountType]
}

func (bankAccntExpBldr *BankAccount_ExpressionBuilder) BuildTree(name string) *DynamoAttribute[*BankAccount_ExpressionBuilder] {
	bankAccntExpBldr = &BankAccount_ExpressionBuilder{}
	bankAccntExpBldr.BankAccountNumber = *NewDynamoAttribute[*int]().WithName("bank_account_number")
	bankAccntExpBldr.AccountType = *NewDynamoAttribute[*BankAccountType]().WithName("account_type")
	return NewDynamoAttribute[*BankAccount_ExpressionBuilder]().
		WithAccessReference(bankAccntExpBldr).
		WithName(name).
		WithChildAttribute(&bankAccntExpBldr.BankAccountNumber).
		WithChildAttribute(&bankAccntExpBldr.AccountType)
}

type BankDetails_ExpressionBuilder struct {
	Accounts DynamoListAttribute[*BankAccount_ExpressionBuilder] // This will be a set in DDB
}

func (bankDetailsExpBldr *BankDetails_ExpressionBuilder) BuildTree(name string) *DynamoAttribute[*BankDetails_ExpressionBuilder] {
	bankDetailsExpBldr = &BankDetails_ExpressionBuilder{}
	bankDetailsExpBldr.Accounts = *NewDynamoListAttribute[*BankAccount_ExpressionBuilder]().
		WithListItemAccessReference(&BankAccount_ExpressionBuilder{}).
		WithName("accounts")
	return NewDynamoAttribute[*BankDetails_ExpressionBuilder]().
		WithAccessReference(bankDetailsExpBldr).
		WithName(name).
		WithChildAttribute(&bankDetailsExpBldr.Accounts)
}

type Child_ExpressionBuilder struct {
	Name DynamoAttribute[*string]
	DOB  DynamoAttribute[*time.Time]
}

func (childExpBldr *Child_ExpressionBuilder) BuildTree(name string) *DynamoAttribute[*Child_ExpressionBuilder] {
	childExpBldr = &Child_ExpressionBuilder{}
	childExpBldr.Name = *NewDynamoAttribute[*string]().WithName("name")
	childExpBldr.DOB = *NewDynamoAttribute[*time.Time]().WithName("dob")
	return NewDynamoAttribute[*Child_ExpressionBuilder]().
		WithAccessReference(childExpBldr).
		WithName(name).
		WithChildAttribute(&childExpBldr.Name).
		WithChildAttribute(&childExpBldr.DOB)
}

type FamilyDetail_ExpressionBuilder struct {
	Children  DynamoListAttribute[*Child_ExpressionBuilder] // This will be a list
	IsMarried DynamoAttribute[*bool]
}

func (familyExpBldr *FamilyDetail_ExpressionBuilder) BuildTree(name string) *DynamoAttribute[*FamilyDetail_ExpressionBuilder] {
	familyExpBldr = &FamilyDetail_ExpressionBuilder{}
	familyExpBldr.Children = *NewDynamoListAttribute[*Child_ExpressionBuilder]().
		WithListItemAccessReference(&Child_ExpressionBuilder{}).
		WithName("children")
	familyExpBldr.IsMarried = *NewDynamoAttribute[*bool]().WithName("is_married")
	return NewDynamoAttribute[*FamilyDetail_ExpressionBuilder]().
		WithAccessReference(familyExpBldr).
		WithName(name).
		WithChildAttribute(&familyExpBldr.Children).
		WithChildAttribute(&familyExpBldr.IsMarried)
}

type Person_ExpressionBuilder struct {
	PK            DynamoKeyAttribute[*string]
	SK            DynamoKeyAttribute[*string]
	Name          DynamoAttribute[*string]
	BankDetails   DynamoAttribute[*BankDetails_ExpressionBuilder]
	FamilyDetails DynamoAttribute[*FamilyDetail_ExpressionBuilder]
	PhoneNos      DynamoListAttribute[*string]
}

func (personExpBldr *Person_ExpressionBuilder) BuildTree(name string) *DynamoAttribute[*Person_ExpressionBuilder] {
	personExpBldr = &Person_ExpressionBuilder{} // Creating new object is necessary because some attributes can be list/set
	personExpBldr.PK = *NewDynamoKeyAttribute[*string]().WithName("pk")
	personExpBldr.SK = *NewDynamoKeyAttribute[*string]().WithName("sk")
	personExpBldr.Name = *NewDynamoAttribute[*string]().WithName("name")
	personExpBldr.BankDetails = *(&BankDetails_ExpressionBuilder{}).BuildTree("bank_details")
	personExpBldr.FamilyDetails = *(&FamilyDetail_ExpressionBuilder{}).BuildTree("family_details")
	personExpBldr.PhoneNos = *NewDynamoListAttribute[*string]().WithName("phone_nos")
	return NewDynamoAttribute[*Person_ExpressionBuilder]().
		WithAccessReference(personExpBldr).
		WithName(name).
		WithChildAttribute(&personExpBldr.PK).
		WithChildAttribute(&personExpBldr.SK).
		WithChildAttribute(&personExpBldr.Name).
		WithChildAttribute(&personExpBldr.BankDetails).
		WithChildAttribute(&personExpBldr.FamilyDetails).
		WithChildAttribute(&personExpBldr.PhoneNos)
}

func NewPerson_ExpressionBuilder() DDBItemExpressionBuilder[*Person_ExpressionBuilder] {
	return NewDDBItemExpressionBuilder(&Person_ExpressionBuilder{})
}

// -------------------------------------------------------------------------------------------------------------

// Testing projection of all the top level attributes
// Projection of attribute at bottom level should be ignored
func TestTopLevelAttributeProjection(t *testing.T) {
	// new expression builder
	expBuilder := NewPerson_ExpressionBuilder()

	// Build NameBuilder, KeyBuilder,
	// list and constructing tree
	expBuilder.Build()

	rootExpBldr := expBuilder.DDBItemRoot().AR()
	rootExpBldr.PK.Project()
	rootExpBldr.SK.Project()
	rootExpBldr.Name.Project()
	rootExpBldr.BankDetails.Project()
	rootExpBldr.FamilyDetails.Project()

	// mark a deep level attribute for projection
	// which should be ignored
	rootExpBldr.FamilyDetails.AR().IsMarried.Project()

	// generate expression
	projBuilder, err := expBuilder.BuildProjectionBuilder()
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	expr, err := expression.NewBuilder().WithProjection(*projBuilder).Build()
	if err != nil {
		t.Errorf(err.Error())
		return
	} else {
		expectedExprNames := map[string]*string{
			"#0": aws.String("pk"),
			"#1": aws.String("sk"),
			"#2": aws.String("name"),
			"#3": aws.String("bank_details"),
			"#4": aws.String("family_details"),
		}

		assert.Equal(t, expectedExprNames, expr.Names())

		exprectedProjectionExpression := "#0, #1, #2, #3, #4"
		assert.Equal(t, exprectedProjectionExpression, *expr.Projection())

		var expectedValuesMap map[string]*dynamodb.AttributeValue = nil
		assert.Equal(t, expectedValuesMap, expr.Values())

		var exprectedUpdateExpression *string = nil
		assert.Equal(t, exprectedUpdateExpression, expr.Update())
	}
}

// Testing projection of all the primitive attributes
// and not complex object attributes
func TestPrimitiveAttributeProjection(t *testing.T) {
	// build expression builder
	expBuilder := NewPerson_ExpressionBuilder()
	rootExpBldr := expBuilder.DDBItemRoot().AR()

	// add list elements
	rootExpBldr.BankDetails.AR().Accounts.AddListItem(1, 3, 5)
	rootExpBldr.FamilyDetails.AR().Children.AddListItem(2, 10)

	// Build NameBuilder, KeyBuilder,
	// list and constructing tree
	expBuilder.Build()

	rootExpBldr.PK.Project()
	rootExpBldr.SK.Project()
	rootExpBldr.Name.Project()

	rootExpBldr.BankDetails.AR().Accounts.Index(1).AR().AccountType.Project()
	rootExpBldr.BankDetails.AR().Accounts.Index(1).AR().BankAccountNumber.Project()
	rootExpBldr.BankDetails.AR().Accounts.Index(3).AR().AccountType.Project()
	rootExpBldr.BankDetails.AR().Accounts.Index(3).AR().BankAccountNumber.Project()
	rootExpBldr.BankDetails.AR().Accounts.Index(5).AR().AccountType.Project()
	rootExpBldr.BankDetails.AR().Accounts.Index(5).AR().BankAccountNumber.Project()

	rootExpBldr.FamilyDetails.AR().IsMarried.Project()
	rootExpBldr.FamilyDetails.AR().Children.Index(2).AR().DOB.Project()
	rootExpBldr.FamilyDetails.AR().Children.Index(2).AR().Name.Project()
	rootExpBldr.FamilyDetails.AR().Children.Index(10).AR().DOB.Project()
	rootExpBldr.FamilyDetails.AR().Children.Index(10).AR().Name.Project()

	// generate expression
	projBuilder, err := expBuilder.BuildProjectionBuilder()
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	expr, err := expression.NewBuilder().WithProjection(*projBuilder).Build()
	if err != nil {
		t.Errorf(err.Error())
		return
	} else {
		expectedExprNames := map[string]*string{
			"#0":  aws.String("pk"),
			"#1":  aws.String("sk"),
			"#2":  aws.String("name"),
			"#3":  aws.String("bank_details"),
			"#4":  aws.String("accounts"),
			"#5":  aws.String("bank_account_number"),
			"#6":  aws.String("account_type"),
			"#7":  aws.String("family_details"),
			"#8":  aws.String("children"),
			"#9":  aws.String("dob"),
			"#10": aws.String("is_married"),
		}

		assert.Equal(t, expectedExprNames, expr.Names())

		exprectedProjectionExpression := "#0, #1, #2, #3.#4[1].#5, #3.#4[1].#6, #3.#4[3].#5, #3.#4[3].#6, #3.#4[5].#5," +
			" #3.#4[5].#6, #7.#8[2].#2, #7.#8[2].#9, #7.#8[10].#2, #7.#8[10].#9, #7.#10"
		assert.Equal(t, exprectedProjectionExpression, *expr.Projection())

		// var expectedValuesMap map[string]types.AttributeValue = nil
		var expectedValuesMap map[string]*dynamodb.AttributeValue = nil

		assert.Equal(t, expectedValuesMap, expr.Values())

		var exprectedUpdateExpression *string = nil
		assert.Equal(t, exprectedUpdateExpression, expr.Update())
	}
}

// Testing condition builder
func TestAttributeCondition(t *testing.T) {
	// build expression builder
	expBuilder := NewPerson_ExpressionBuilder()
	rootExpBldr := expBuilder.DDBItemRoot().AR()

	rootExpBldr.BankDetails.AR().Accounts.AddListItem(1, 3, 5) // this will be a set
	rootExpBldr.FamilyDetails.AR().Children.AddListItem(2, 10) // this will be a list

	// Build NameBuilder, KeyBuilder,
	// list and constructing tree
	expBuilder.Build()

	rootExpBldr.PK.AndWithCondition()(rootExpBldr.PK.GetKeyBuilder().Equal(expression.Value("PartionValue")))
	rootExpBldr.SK.AndWithCondition()(rootExpBldr.SK.GetKeyBuilder().BeginsWith("SortKeyPrefix"))

	rootExpBldr.Name.AndWithCondition()(rootExpBldr.Name.GetNameBuilder().BeginsWith("NamePrefix"))

	rootExpBldr.BankDetails.AR().Accounts.
		AndWithCondition()(expression.ConditionBuilder(rootExpBldr.BankDetails.AR().Accounts.GetNameBuilder().AttributeExists()))

	rootExpBldr.BankDetails.AR().Accounts.Index(1).
		AndWithCondition()(expression.ConditionBuilder(rootExpBldr.BankDetails.GetNameBuilder().AttributeExists()))

	rootExpBldr.BankDetails.AR().Accounts.Index(1).AR().BankAccountNumber.
		AndWithCondition()(expression.ConditionBuilder(rootExpBldr.BankDetails.AR().Accounts.Index(1).AR().BankAccountNumber.GetNameBuilder().Equal(expression.Value("SomeBankAccntNumber"))))

	rootExpBldr.FamilyDetails.AR().Children.Index(2).AR().Name.
		AndWithCondition()(expression.ConditionBuilder(rootExpBldr.FamilyDetails.AR().Children.Index(2).AR().Name.GetNameBuilder().Equal(expression.Value("ChildrenName"))))

	// generate expression
	projBuilder, err := expBuilder.BuildProjectionBuilder()
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	keyConditionBuilder := expBuilder.BuildKeyConditionBuilder()
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	conditionBuilder := expBuilder.BuildConditionBuilder()
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	expr, err := expression.NewBuilder().WithProjection(*projBuilder).WithKeyCondition(*keyConditionBuilder).WithCondition(*conditionBuilder).Build()
	if err != nil {
		t.Errorf(err.Error())
		return
	} else {
		expectedExprNames := map[string]*string{
			"#0": aws.String("name"),
			"#1": aws.String("bank_details"),
			"#2": aws.String("accounts"),
			"#3": aws.String("bank_account_number"),
			"#4": aws.String("family_details"),
			"#5": aws.String("children"),
			"#6": aws.String("pk"),
			"#7": aws.String("sk"),
		}

		assert.Equal(t, expectedExprNames, expr.Names())

		exprectedProjectionExpression := "#6, #7"
		assert.Equal(t, exprectedProjectionExpression, *expr.Projection())

		var expectedValuesMap = map[string]*dynamodb.AttributeValue{
			":0": {
				S: aws.String("NamePrefix"),
			},
			":1": {
				S: aws.String("SomeBankAccntNumber"),
			},
			":2": {
				S: aws.String("ChildrenName"),
			},
			":3": {
				S: aws.String("PartionValue"),
			},
			":4": {
				S: aws.String("SortKeyPrefix"),
			},
		}

		assert.Equal(t, expectedValuesMap, expr.Values())

		var exprectedUpdateExpression *string = nil
		assert.Equal(t, exprectedUpdateExpression, expr.Update())

		exprectedCondition := "((((begins_with (#0, :0)) AND (attribute_exists (#1.#2))) AND (attribute_exists (#1))) AND (#1.#2[1].#3 = :1)) AND (#4.#5[2].#0 = :2)"
		assert.Equal(t, exprectedCondition, *expr.Condition())

		exprectedKeyCondition := "(#6 = :3) AND (begins_with (#7, :4))"
		assert.Equal(t, exprectedKeyCondition, *expr.KeyCondition())
	}
}

// Testing update builder
func TestAttributeUpdate(t *testing.T) {
	// build expression builder
	expBuilder := NewPerson_ExpressionBuilder()
	rootExpBldr := expBuilder.DDBItemRoot().AR()

	// add list elements
	rootExpBldr.PhoneNos.AddListItem(12, 14)
	rootExpBldr.BankDetails.AR().Accounts.AddListItem(1, 3, 5)
	rootExpBldr.FamilyDetails.AR().Children.AddListItem(2, 10)

	// Build NameBuilder, KeyBuilder,
	// list and constructing tree
	expBuilder.Build()

	rootExpBldr.Name.AddValue(UPDATE_SET, utils.PointerTo("New Name"))
	rootExpBldr.BankDetails.AR().Accounts.
		AndWithCondition()(expression.ConditionBuilder(rootExpBldr.BankDetails.AR().Accounts.GetNameBuilder().AttributeExists()))

	// setting the whole attribute
	rootExpBldr.BankDetails.AR().Accounts.Index(1).AddValue(UPDATE_SET, BankAccount{
		AccountType:       utils.PointerTo(SAVINGS),
		BankAccountNumber: utils.PointerTo(4857372829),
	})
	// this should be ignored by expression builder, since parent attribute i.e. index 1 is set
	rootExpBldr.BankDetails.AR().Accounts.Index(1).AR().BankAccountNumber.AddValue(UPDATE_SET, "NewBankAccntNumber")

	rootExpBldr.BankDetails.AR().Accounts.Index(3).AR().BankAccountNumber.AddValue(UPDATE_ADD, 2)
	rootExpBldr.BankDetails.AR().Accounts.Index(5).AddValue(UPDATE_REMOVE, nil)

	rootExpBldr.PhoneNos.Index(14).AddValue(UPDATE_SET, []*string{utils.PointerTo("7638927366")})

	// adding one more children
	utcLocation, _ := time.LoadLocation("UTC")
	rootExpBldr.FamilyDetails.AR().Children.AddValue(
		UPDATE_SET,
		rootExpBldr.FamilyDetails.AR().Children.GetNameBuilder().ListAppend(
			expression.Value(Child{
				Name: utils.PointerTo("NewChild"),
				DOB:  utils.PointerTo(time.Date(2023, 5, 3, 0, 0, 0, 0, utcLocation)),
			}),
		))

	updateBuilder, err := expBuilder.BuildUpdateBuilder()
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	expr, err := expression.NewBuilder().WithUpdate(*updateBuilder).Build()
	if err != nil {
		t.Errorf(err.Error())
		return
	} else {
		expectedExprNames := map[string]*string{
			"#0": aws.String("bank_details"),
			"#1": aws.String("accounts"),
			"#2": aws.String("bank_account_number"),
			"#3": aws.String("name"),
			"#4": aws.String("family_details"),
			"#5": aws.String("children"),
			"#6": aws.String("phone_nos"),
		}

		assert.Equal(t, expectedExprNames, expr.Names())

		var expectedValuesMap = map[string]*dynamodb.AttributeValue{
			":0": {
				N: aws.String("2"),
			},
			":1": {
				S: aws.String("New Name"),
			},
			":2": {
				M: map[string]*dynamodb.AttributeValue{
					"account_type": {
						N: aws.String(strconv.Itoa(int(SAVINGS))),
					},
					"bank_account_number": {
						N: aws.String("4857372829"),
					},
				},
			},
			":3": {
				M: map[string]*dynamodb.AttributeValue{
					"name": {
						S: aws.String("NewChild"),
					},
					"dob": {
						S: aws.String("2023-05-03T00:00:00Z"),
					},
				},
			},
			":4": {
				L: []*dynamodb.AttributeValue{
					{
						S: aws.String("7638927366"),
					},
				},
			},
		}

		assert.Equal(t, expectedValuesMap, expr.Values())

		exprectedUpdateExpression := "ADD #0.#1[3].#2 :0\nREMOVE #0.#1[5]\nSET #3 = :1, #0.#1[1] = :2, #4.#5 = list_append(#4.#5, :3), #6[14] = :4\n"
		assert.Equal(t, exprectedUpdateExpression, *expr.Update())
	}
}
