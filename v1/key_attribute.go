package v1

import (
	"errors"

	"github.com/gauxs/dynexpr/utils"

	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

// Represents a dyanmo db primary key attribute
type DynamoKeyAttribute[T any] struct {
	// True when build has been executed on 'this' attribute
	// projection, Condition and Update can only be done after
	// build is executed
	buildExecuted bool

	// Mark 'this' attribute for projection
	projection bool

	// Name of the dynamo attribute as defined in DB
	Name string

	// Document path of this attribute
	documentPath string

	// Use this to build any condition on 'this' attribute
	keyBuilder expression.KeyBuilder

	// Represents all the conditions applied on 'this' dynamo attribute
	keyConditionBuilder *expression.KeyConditionBuilder
}

func NewDynamoKeyAttribute[T any]() *DynamoKeyAttribute[T] {
	return &DynamoKeyAttribute[T]{}
}

func (dka *DynamoKeyAttribute[T]) constructDocumentPath(documentPathOfParent string) string {
	return dka.GetName()
}

// WithName builds `this` DynamoKeyAttribute with a dynamo db attribute name
func (dka *DynamoKeyAttribute[T]) WithName(name string) *DynamoKeyAttribute[T] {
	dka.Name = name
	return dka
}

func (dka *DynamoKeyAttribute[T]) build(parentDocumentPath string) error {
	if dka.buildExecuted {
		return errors.New("build is already executed on attribute " + dka.documentPath)
	}

	// build document path
	dka.documentPath = dka.constructDocumentPath(parentDocumentPath)

	// build key builder
	dka.keyBuilder = expression.Key(dka.documentPath)

	dka.buildExecuted = true
	return nil
}

// Project marks `this` attribute for projection
func (dka *DynamoKeyAttribute[T]) Project() error {
	if !dka.buildExecuted {
		return errors.New("build is not yet executed on attribute [" + dka.Name + "], cannot mark this attribute for projection")
	}

	dka.projection = true
	return nil
}

func (dka *DynamoKeyAttribute[T]) GetName() string {
	return dka.Name
}

func (dka *DynamoKeyAttribute[T]) GetKeyBuilder() expression.KeyBuilder {
	return dka.keyBuilder
}

func (dka *DynamoKeyAttribute[T]) addName(projectionBuilder *expression.ProjectionBuilder) (*expression.ProjectionBuilder, error) {
	if !dka.buildExecuted {
		return nil, errors.New("build is not yet executed on attribute [" + dka.Name + "], cannot mark this attribute for projection")
	}

	if projectionBuilder == nil {
		return nil, errors.New("nil projection builder passed for attribute " + dka.documentPath)
	}

	return utils.PointerTo(projectionBuilder.AddNames(expression.Name(dka.documentPath))), nil
}

// AndWithCondition adds a new condition to `this` attributes existing conditions using `AND`
// NOTE: keyConditionBuilder represent any valid condition, zero value of struct `KeyConditionBuilder`
// might give build error
func (dka *DynamoKeyAttribute[T]) AndWithCondition() func(keyConditionBuilder expression.KeyConditionBuilder) {
	return func(keyConditionBuilder expression.KeyConditionBuilder) {
		if dka.keyConditionBuilder != nil {
			// if keyConditionBuilder.IsSet() {
			dka.keyConditionBuilder = utils.PointerTo((*dka.keyConditionBuilder).And(keyConditionBuilder))
			// }
		} else {
			// dka.isKeyConditionBuilderSet = true
			dka.keyConditionBuilder = &keyConditionBuilder
		}
	}
}

func (dka *DynamoKeyAttribute[T]) addKeyCondition(keyConditionBuilder *expression.KeyConditionBuilder) (newKeyConditionBuilder *expression.KeyConditionBuilder) {
	if dka.keyConditionBuilder != nil {
		if keyConditionBuilder != nil {
			// parent condition will be L.H.S
			newKeyConditionBuilder = utils.PointerTo(keyConditionBuilder.And(*dka.keyConditionBuilder))
		} else {
			newKeyConditionBuilder = dka.keyConditionBuilder
		}
	} else {
		newKeyConditionBuilder = keyConditionBuilder
	}

	return newKeyConditionBuilder
}
