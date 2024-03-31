package expression

// import (
// 	"testing"

// 	"github.com/aws/aws-sdk-go/aws"
// 	"github.com/aws/aws-sdk-go/service/dynamodb"
// 	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
// 	"github.com/stretchr/testify/assert"

// 	test_models "github.com/gauxs/dynexpr/test/expression/data"
// )

// // Testing projection of all the top level attributes
// // Projection of attribute at bottom level should be ignored
// func TestTopLevelAttributeProjection(t *testing.T) {
// 	// new expression builder
// 	expBuilder := test_models.NewPerson_ExpressionBuilder()

// 	// Build NameBuilder, KeyBuilder,
// 	// list and constructing tree
// 	expBuilder.Build()

// 	rootExpBldr := expBuilder.DDBItemRoot().AR()
// 	rootExpBldr.PK.Project()
// 	rootExpBldr.SK.Project()
// 	rootExpBldr.Name.Project()
// 	rootExpBldr.BankDetails.Project()
// 	rootExpBldr.FamilyDetails.Project()

// 	// mark a deep level attribute for projection
// 	// which should be ignored
// 	rootExpBldr.FamilyDetails.AR().IsMarried.Project()

// 	// generate expression
// 	projBuilder, err := expBuilder.BuildProjectionBuilder()
// 	if err != nil {
// 		t.Errorf(err.Error())
// 		return
// 	}

// 	expr, err := expression.NewBuilder().WithProjection(*projBuilder).Build()
// 	if err != nil {
// 		t.Errorf(err.Error())
// 		return
// 	} else {
// 		expectedExprNames := map[string]*string{
// 			"#0": aws.String("pk"),
// 			"#1": aws.String("sk"),
// 			"#2": aws.String("name"),
// 			"#3": aws.String("bank_details"),
// 			"#4": aws.String("family_details"),
// 		}

// 		assert.Equal(t, expectedExprNames, expr.Names())

// 		exprectedProjectionExpression := "#0, #1, #2, #3, #4"
// 		assert.Equal(t, exprectedProjectionExpression, *expr.Projection())

// 		var expectedValuesMap map[string]*dynamodb.AttributeValue = nil
// 		assert.Equal(t, expectedValuesMap, expr.Values())

// 		var exprectedUpdateExpression *string = nil
// 		assert.Equal(t, exprectedUpdateExpression, expr.Update())
// 	}
// }

// // Testing projection of all the primitive attributes
// // and not complex object attributes
// func TestPrimitiveAttributeProjection(t *testing.T) {
// 	// build expression builder
// 	expBuilder := NewPerson_ExpressionBuilder()
// 	rootExpBldr := expBuilder.DDBItemRoot().AR()

// 	// add list elements
// 	rootExpBldr.BankDetails.AR().Accounts.AddListItem(1, 3, 5)
// 	rootExpBldr.FamilyDetails.AR().Children.AddListItem(2, 10)

// 	// Build NameBuilder, KeyBuilder,
// 	// list and constructing tree
// 	expBuilder.Build()

// 	rootExpBldr.PK.Project()
// 	rootExpBldr.SK.Project()
// 	rootExpBldr.Name.Project()

// 	rootExpBldr.BankDetails.AR().Accounts.Index(1).AR().AccountType.Project()
// 	rootExpBldr.BankDetails.AR().Accounts.Index(1).AR().BankAccountNumber.Project()
// 	rootExpBldr.BankDetails.AR().Accounts.Index(3).AR().AccountType.Project()
// 	rootExpBldr.BankDetails.AR().Accounts.Index(3).AR().BankAccountNumber.Project()
// 	rootExpBldr.BankDetails.AR().Accounts.Index(5).AR().AccountType.Project()
// 	rootExpBldr.BankDetails.AR().Accounts.Index(5).AR().BankAccountNumber.Project()

// 	rootExpBldr.FamilyDetails.AR().IsMarried.Project()
// 	rootExpBldr.FamilyDetails.AR().Children.Index(2).AR().DOB.Project()
// 	rootExpBldr.FamilyDetails.AR().Children.Index(2).AR().Name.Project()
// 	rootExpBldr.FamilyDetails.AR().Children.Index(10).AR().DOB.Project()
// 	rootExpBldr.FamilyDetails.AR().Children.Index(10).AR().Name.Project()

// 	// generate expression
// 	projBuilder, err := expBuilder.BuildProjectionBuilder()
// 	if err != nil {
// 		t.Errorf(err.Error())
// 		return
// 	}

// 	expr, err := expression.NewBuilder().WithProjection(*projBuilder).Build()
// 	if err != nil {
// 		t.Errorf(err.Error())
// 		return
// 	} else {
// 		expectedExprNames := map[string]*string{
// 			"#0":  aws.String("pk"),
// 			"#1":  aws.String("sk"),
// 			"#2":  aws.String("name"),
// 			"#3":  aws.String("bank_details"),
// 			"#4":  aws.String("accounts"),
// 			"#5":  aws.String("bank_account_number"),
// 			"#6":  aws.String("account_type"),
// 			"#7":  aws.String("family_details"),
// 			"#8":  aws.String("children"),
// 			"#9":  aws.String("dob"),
// 			"#10": aws.String("is_married"),
// 		}

// 		assert.Equal(t, expectedExprNames, expr.Names())

// 		exprectedProjectionExpression := "#0, #1, #2, #3.#4[1].#5, #3.#4[1].#6, #3.#4[3].#5, #3.#4[3].#6, #3.#4[5].#5," +
// 			" #3.#4[5].#6, #7.#8[2].#2, #7.#8[2].#9, #7.#8[10].#2, #7.#8[10].#9, #7.#10"
// 		assert.Equal(t, exprectedProjectionExpression, *expr.Projection())

// 		// var expectedValuesMap map[string]types.AttributeValue = nil
// 		var expectedValuesMap map[string]*dynamodb.AttributeValue = nil

// 		assert.Equal(t, expectedValuesMap, expr.Values())

// 		var exprectedUpdateExpression *string = nil
// 		assert.Equal(t, exprectedUpdateExpression, expr.Update())
// 	}
// }
