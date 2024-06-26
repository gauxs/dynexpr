package expression

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/stretchr/testify/assert"

	dynexprv1 "github.com/gauxs/dynexpr/pkg/v1"
	test_models "github.com/gauxs/dynexpr/test/expression/data"
	test_helpers "github.com/gauxs/dynexpr/test/expression/helpers"
)

const (
	destinationDirPath string = "/test/expression/data"
)

func TestTransaction(t *testing.T) {
	// build expression builder
	dynexprBldr := test_models.NewTransaction_ExpressionBuilder()
	dynexprBldr.Build()

	ddbItem := dynexprBldr.DDBItemRoot()

	// project attributes
	ddbItem.AR().UserID.Project()
	ddbItem.AR().TransactionID.Project()
	ddbItem.AR().Amount.Project()

	// add condition
	ddbItem.AR().TransactionID.AndWithCondition()(
		ddbItem.AR().TransactionID.GetKeyBuilder().Equal(expression.Value("userID#123")))

	// update attributes
	ddbItem.AR().Amount.AddValue(dynexprv1.UPDATE_SET, 9000)

	projBldr, _ := dynexprBldr.BuildProjectionBuilder()
	updtBldr, _ := dynexprBldr.BuildUpdateBuilder()
	dynamoDBExpr, err := expression.NewBuilder().
		WithProjection(*projBldr).
		WithKeyCondition(*(dynexprBldr.BuildKeyConditionBuilder())).
		WithUpdate(*updtBldr).
		Build()
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	fmt.Println(string(*(dynamoDBExpr.Projection())))
	fmt.Println(string(*(dynamoDBExpr.KeyCondition())))
	fmt.Println(string(*(dynamoDBExpr.Update())))

	nM, _ := json.Marshal(dynamoDBExpr.Names())
	vM, _ := json.Marshal(dynamoDBExpr.Values())
	fmt.Println(string(nM))
	fmt.Println(string(vM))

}

func TestAttributeCondition(t *testing.T) {
	assert.Nil(t, test_helpers.GenerateExpressionBuilder(destinationDirPath))

	// build expression builder
	expBuilder := test_models.NewPerson_ExpressionBuilder()
	rootExpBldr := expBuilder.DDBItemRoot().AR()

	rootExpBldr.BankDetails.AR().Accounts.AddListItem(1, 3, 5) // this will be a set
	rootExpBldr.FamilyDetails.AR().Children.AddListItem(2, 10) // this will be a list

	// Build NameBuilder, KeyBuilder,
	// list and constructing tree
	expBuilder.Build()

	rootExpBldr.PK.AndWithCondition()(
		rootExpBldr.PK.GetKeyBuilder().Equal(expression.Value("PartionValue")))

	rootExpBldr.SK.AndWithCondition()(
		rootExpBldr.SK.GetKeyBuilder().BeginsWith("SortKeyPrefix"))

	rootExpBldr.Name.AndWithCondition()(
		rootExpBldr.Name.GetNameBuilder().BeginsWith("NamePrefix"))

	rootExpBldr.BankDetails.AR().Accounts.AndWithCondition()(
		expression.ConditionBuilder(rootExpBldr.BankDetails.AR().Accounts.GetNameBuilder().AttributeExists()))

	rootExpBldr.BankDetails.AR().Accounts.Index(1).AndWithCondition()(
		expression.ConditionBuilder(rootExpBldr.BankDetails.GetNameBuilder().AttributeExists()))

	rootExpBldr.BankDetails.AR().Accounts.Index(1).AR().BankAccountNumber.AndWithCondition()(
		expression.ConditionBuilder(rootExpBldr.BankDetails.AR().Accounts.Index(1).AR().
			BankAccountNumber.GetNameBuilder().Equal(expression.Value("SomeBankAccntNumber"))))

	rootExpBldr.FamilyDetails.AR().Children.Index(2).AR().Name.AndWithCondition()(
		expression.ConditionBuilder(rootExpBldr.FamilyDetails.AR().Children.Index(2).AR().
			Name.GetNameBuilder().Equal(expression.Value("ChildrenName"))))

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

	expr, err := expression.NewBuilder().
		WithProjection(*projBuilder).
		WithKeyCondition(*keyConditionBuilder).
		WithCondition(*conditionBuilder).
		Build()
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
