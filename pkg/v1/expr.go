package v1

import (
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

const (
	DDBAtributeNameCancatenator = "."
)

// This is implemented by structs using this package
type TreeBuilder[T any] interface {
	// BuildTree builds the tree and returns the
	// root node i.e. *DynamoAttribute
	BuildTree(string) *DynamoAttribute[T]
}

// TODO: instead of Builder and Projector we should have union of
// DynamoAttribute and DynamoListAttribute but current limitation
// doesn't allow us.
// Ref: https://stackoverflow.com/a/71378366

type Builder interface {
	// build builds `this` attribute. If `this` attribute
	// is a list then new nodes(elements of the list) are attached
	// if specified via AddListItem
	build(string) error
}

// TODO: merge this with Builder
type Projector interface {
	// addName adds 'this' attribute's document name into the projection builder
	// passed in argument and returns a new projection builder
	addName(*expression.ProjectionBuilder) (*expression.ProjectionBuilder, error)
}

type Conditioner interface {
	// addCondition adds 'this' attribute's condition into the condition builder
	// passed in argument and returns a new condition builder
	//
	// Implicitly uses `AND` between `this` attribute's condition and
	// condition passed in argument
	//
	// condition passed in argument always comes first in `AND`
	addCondition(*expression.ConditionBuilder) *expression.ConditionBuilder
}

type KeyConditioner interface {
	// addCondition adds 'this' attribute's key condition into the key condition builder
	// passed in argument and returns a new key condition builder
	//
	// Implicitly uses `AND` between `this` attribute's key condition and
	// condition passed in argument
	//
	// condition passed in argument always comes first in `AND`
	addKeyCondition(*expression.KeyConditionBuilder) *expression.KeyConditionBuilder
}

type Updater interface {
	// addUpdate adds 'this' attribute's update value into the update builder
	// passed in argument and returns a new update builder
	addUpdate(*expression.UpdateBuilder) (*expression.UpdateBuilder, error)
}

// Enforcing constraints at compile time
var _ Builder = (&DynamoAttribute[int]{})
var _ Updater = (&DynamoAttribute[int]{})
var _ Projector = (&DynamoAttribute[int]{})
var _ Conditioner = (&DynamoAttribute[int]{})

var _ Builder = (&DynamoListAttribute[int]{})
var _ Updater = (&DynamoListAttribute[int]{})
var _ Projector = (&DynamoListAttribute[int]{})
var _ Conditioner = (&DynamoListAttribute[int]{})

var _ Projector = (&DynamoKeyAttribute[int]{})
var _ KeyConditioner = (&DynamoKeyAttribute[int]{})

type DDBItemExpressionBuilder[T any] struct {
	// root of the ddb item
	root *DynamoAttribute[T]
}

func (d DDBItemExpressionBuilder[T]) DDBItemRoot() *DynamoAttribute[T] {
	return d.root
}

// Build creates additional tree structure for attributes of List/Set data type
// this has to be invoked before any builder can be built
func (d DDBItemExpressionBuilder[T]) Build() error {
	return d.root.build("")
}

// BuildProjectionBuilder builds a ProjectionBuilder by aggregating all the projection of this
// expression builder tree
func (d DDBItemExpressionBuilder[T]) BuildProjectionBuilder() (*expression.ProjectionBuilder, error) {
	return d.root.addName(&expression.ProjectionBuilder{})
}

// BuildKeyConditionBuilder builds a KeyConditionBuilder by aggregating all the KeyCondition of this
// expression builder tree
func (d DDBItemExpressionBuilder[T]) BuildKeyConditionBuilder() *expression.KeyConditionBuilder {
	var keyConditionBuilder *expression.KeyConditionBuilder
	// key condition will always be top level attribute, so we can directly traverse root's child attribute only
	// instead of a recusrsive solution
	for _, childAttribute := range d.root.childAttributes {
		switch childAttributeType := childAttribute.(type) {
		case KeyConditioner:
			if newKeyConditionBuilder := childAttributeType.addKeyCondition(keyConditionBuilder); newKeyConditionBuilder != nil {
				keyConditionBuilder = newKeyConditionBuilder
			}
		}
	}

	return keyConditionBuilder
}

// BuildConditionBuilder builds a ConditionBuilder by aggregating all the condition of this
// expression builder tree
func (d DDBItemExpressionBuilder[T]) BuildConditionBuilder() *expression.ConditionBuilder {
	// return d.root.addCondition(&expression.ConditionBuilder{})
	return d.root.addCondition(nil)
}

// BuildUpdateBuilder builds a UpdateBuilder by aggregating all the update operation of this
// expression builder tree
func (d DDBItemExpressionBuilder[T]) BuildUpdateBuilder() (*expression.UpdateBuilder, error) {
	return d.root.addUpdate(&expression.UpdateBuilder{})
}
