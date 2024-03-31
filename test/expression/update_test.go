package expression

// import (
// 	"strconv"
// 	"testing"
// 	"time"

// 	"github.com/aws/aws-sdk-go/aws"
// 	"github.com/aws/aws-sdk-go/service/dynamodb"
// 	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
// 	"github.com/gauxs/dynexpr/internal/utils"
// 	"github.com/stretchr/testify/assert"

// 	test_models "github.com/gauxs/dynexpr/test/expression/data"
// )

// // Testing update builder
// func TestAttributeUpdate(t *testing.T) {
// 	// build expression builder
// 	expBuilder := test_models.NewPerson_ExpressionBuilder()
// 	rootExpBldr := expBuilder.DDBItemRoot().AR()

// 	// add list elements
// 	rootExpBldr.PhoneNos.AddListItem(12, 14)
// 	rootExpBldr.BankDetails.AR().Accounts.AddListItem(1, 3, 5)
// 	rootExpBldr.FamilyDetails.AR().Children.AddListItem(2, 10)

// 	// Build NameBuilder, KeyBuilder,
// 	// list and constructing tree
// 	expBuilder.Build()

// 	rootExpBldr.Name.AddValue(UPDATE_SET, utils.PointerTo("New Name"))
// 	rootExpBldr.BankDetails.AR().Accounts.
// 		AndWithCondition()(expression.ConditionBuilder(rootExpBldr.BankDetails.AR().Accounts.GetNameBuilder().AttributeExists()))

// 	// setting the whole attribute
// 	rootExpBldr.BankDetails.AR().Accounts.Index(1).AddValue(UPDATE_SET, BankAccount{
// 		AccountType:       utils.PointerTo(SAVINGS),
// 		BankAccountNumber: utils.PointerTo(4857372829),
// 	})
// 	// this should be ignored by expression builder, since parent attribute i.e. index 1 is set
// 	rootExpBldr.BankDetails.AR().Accounts.Index(1).AR().BankAccountNumber.AddValue(UPDATE_SET, "NewBankAccntNumber")

// 	rootExpBldr.BankDetails.AR().Accounts.Index(3).AR().BankAccountNumber.AddValue(UPDATE_ADD, 2)
// 	rootExpBldr.BankDetails.AR().Accounts.Index(5).AddValue(UPDATE_REMOVE, nil)

// 	rootExpBldr.PhoneNos.Index(14).AddValue(UPDATE_SET, []*string{utils.PointerTo("7638927366")})

// 	// adding one more children
// 	utcLocation, _ := time.LoadLocation("UTC")
// 	rootExpBldr.FamilyDetails.AR().Children.AddValue(
// 		UPDATE_SET,
// 		rootExpBldr.FamilyDetails.AR().Children.GetNameBuilder().ListAppend(
// 			expression.Value(Child{
// 				Name: utils.PointerTo("NewChild"),
// 				DOB:  utils.PointerTo(time.Date(2023, 5, 3, 0, 0, 0, 0, utcLocation)),
// 			}),
// 		))

// 	updateBuilder, err := expBuilder.BuildUpdateBuilder()
// 	if err != nil {
// 		t.Errorf(err.Error())
// 		return
// 	}

// 	expr, err := expression.NewBuilder().WithUpdate(*updateBuilder).Build()
// 	if err != nil {
// 		t.Errorf(err.Error())
// 		return
// 	} else {
// 		expectedExprNames := map[string]*string{
// 			"#0": aws.String("bank_details"),
// 			"#1": aws.String("accounts"),
// 			"#2": aws.String("bank_account_number"),
// 			"#3": aws.String("name"),
// 			"#4": aws.String("family_details"),
// 			"#5": aws.String("children"),
// 			"#6": aws.String("phone_nos"),
// 		}

// 		assert.Equal(t, expectedExprNames, expr.Names())

// 		var expectedValuesMap = map[string]*dynamodb.AttributeValue{
// 			":0": {
// 				N: aws.String("2"),
// 			},
// 			":1": {
// 				S: aws.String("New Name"),
// 			},
// 			":2": {
// 				M: map[string]*dynamodb.AttributeValue{
// 					"account_type": {
// 						N: aws.String(strconv.Itoa(int(SAVINGS))),
// 					},
// 					"bank_account_number": {
// 						N: aws.String("4857372829"),
// 					},
// 				},
// 			},
// 			":3": {
// 				M: map[string]*dynamodb.AttributeValue{
// 					"name": {
// 						S: aws.String("NewChild"),
// 					},
// 					"dob": {
// 						S: aws.String("2023-05-03T00:00:00Z"),
// 					},
// 				},
// 			},
// 			":4": {
// 				L: []*dynamodb.AttributeValue{
// 					{
// 						S: aws.String("7638927366"),
// 					},
// 				},
// 			},
// 		}

// 		assert.Equal(t, expectedValuesMap, expr.Values())

// 		exprectedUpdateExpression := "ADD #0.#1[3].#2 :0\nREMOVE #0.#1[5]\nSET #3 = :1, #0.#1[1] = :2, #4.#5 = list_append(#4.#5, :3), #6[14] = :4\n"
// 		assert.Equal(t, exprectedUpdateExpression, *expr.Update())
// 	}
// }
