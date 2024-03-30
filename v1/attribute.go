package v1

import (
	"errors"
	"strings"

	"github.com/gauxs/dynexpr/utils"

	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

// Represents a dyanmo db attribute
type DynamoAttribute[T any] struct {
	// True when build has been executed on 'this' attribute
	// projection, Condition and Update can only be done after
	// build is executed
	buildExecuted bool

	// Mark 'this' attribute for projection
	projection bool

	// Name of the dynamo attribute as defined in DB
	name string

	// Document path of this attribute
	documentPath string

	// Use this to build any condition on 'this' attribute
	nameBuilder expression.NameBuilder

	// Represents all the conditions applied on 'this' dynamo attribute
	conditionBuilder *expression.ConditionBuilder

	// Helps in direct member selection of child attributes
	accessReference T

	// Determines the operation which needs to performed on
	// 'this' attribute
	operation DynamoOperation

	// Represent the value which needs to be assigned to 'this'
	// attribute for update
	value any

	// Child attributes of 'this' attribute, can be DynamoAttribute
	// or DynamoListAttribute
	childAttributes []interface{}
}

func NewDynamoAttribute[T any]() *DynamoAttribute[T] {
	return &DynamoAttribute[T]{
		childAttributes: []interface{}{},
	}
}

// WithName builds `this` DynamoAttribute with a dynamo db attribute name
func (da *DynamoAttribute[T]) WithName(name string) *DynamoAttribute[T] {
	da.name = name
	return da
}

// WithAccessReference builds `this` DynamoAttribute with an access reference
// access refernece is used in direct member selection
func (da *DynamoAttribute[T]) WithAccessReference(accessReference T) *DynamoAttribute[T] {
	da.accessReference = accessReference
	return da
}

// WithChildAttribute builds `this` DynamoAttribute child attribute which are
// child nodes holding member attribute of type `T`
func (da *DynamoAttribute[T]) WithChildAttribute(childAttribute interface{}) *DynamoAttribute[T] {
	da.childAttributes = append(da.childAttributes, childAttribute)
	return da
}

// Project marks `this` attribute for projection
func (da *DynamoAttribute[T]) Project() error {
	if !da.buildExecuted {
		return errors.New("build is not yet executed on attribute [" + da.name + "], cannot mark this attribute for projection")
	}

	da.projection = true
	return nil
}

func (da *DynamoAttribute[T]) GetName() string {
	return da.name
}

func (da *DynamoAttribute[T]) GetNameBuilder() expression.NameBuilder {
	return da.nameBuilder
}

// AR returns the type `T` held by this node
// access refernece is used in direct member selection
func (da *DynamoAttribute[T]) AR() T {
	return da.accessReference
}

// AndWithCondition adds a new condition to `this` attributes existing conditions using `AND`
// NOTE: conditionBuilder represent any valid condition, zero value of struct `ConditionBuilder`
// might give build error
func (da *DynamoAttribute[T]) AndWithCondition() func(conditionBuilder expression.ConditionBuilder) {
	return func(conditionBuilder expression.ConditionBuilder) {
		if da.conditionBuilder != nil {
			da.conditionBuilder = utils.PointerTo((*da.conditionBuilder).And(conditionBuilder))
		} else {
			da.conditionBuilder = &conditionBuilder
		}
	}
}

// AddValue adds a value which will be used to update `this` attributes
func (da *DynamoAttribute[T]) AddValue(operation DynamoOperation, value any) {
	da.operation = operation

	// TODO: time.Time will automatically convert to string
	// add a switch case hee and allow clients to specify
	// which data type they want time to be in
	da.value = value
}

func (da *DynamoAttribute[T]) constructDocumentPath(documentPathOfParent string) string {
	if strings.HasSuffix(da.GetName(), "]") { // this is a top level element of a list, don't use DDBAtributeNameCancatenator
		return documentPathOfParent + da.GetName()
	}

	if documentPathOfParent == "" { // this is top level attribute
		return da.GetName()
	}

	return documentPathOfParent + DDBAtributeNameCancatenator + da.GetName()
}

func (da *DynamoAttribute[T]) build(parentDocumentPath string) error {
	if da.buildExecuted {
		return errors.New("build is already executed on attribute " + da.documentPath)
	}

	// build document path
	da.documentPath = da.constructDocumentPath(parentDocumentPath)

	// build name builder
	da.nameBuilder = expression.Name(da.documentPath)

	for _, childAttribute := range da.childAttributes {
		switch childAttributeType := childAttribute.(type) {
		case Builder:
			childAttributeType.build(da.documentPath)
		}
	}

	da.buildExecuted = true
	return nil
}

// addName recursively collects all the attributes which were marked for projection
func (da *DynamoAttribute[T]) addName(projectionBuilder *expression.ProjectionBuilder) (newProjectionBuilder *expression.ProjectionBuilder, err error) {
	if !da.buildExecuted {
		return nil, errors.New("build is not yet executed on attribute [" + da.name + "], cannot mark this attribute for projection")
	}

	if projectionBuilder == nil {
		return nil, errors.New("nil projection builder passed for attribute " + da.documentPath)
	}

	newProjectionBuilder = projectionBuilder
	if da.projection { // skipping projection of child attributes
		return utils.PointerTo(newProjectionBuilder.AddNames(expression.Name(da.documentPath))), nil
	} else {
		for _, childAttribute := range da.childAttributes {
			switch childAttributeType := childAttribute.(type) {
			case Projector:
				newProjectionBuilder, err = childAttributeType.addName(newProjectionBuilder)
			}

			if err != nil {
				return nil, err
			}
		}
	}

	return newProjectionBuilder, nil
}

func (da *DynamoAttribute[T]) addCondition(conditionBuilder *expression.ConditionBuilder) (newConditionBuilder *expression.ConditionBuilder) {
	if da.conditionBuilder != nil {
		if conditionBuilder != nil {
			newConditionBuilder = utils.PointerTo(conditionBuilder.And(*da.conditionBuilder))
		} else {
			newConditionBuilder = da.conditionBuilder
		}
	} else {
		newConditionBuilder = conditionBuilder
	}

	for _, childAttribute := range da.childAttributes {
		switch childAttributeType := childAttribute.(type) {
		case Conditioner:
			// when both passed and returned condition builder are nil
			if retConditionBuilder := childAttributeType.addCondition(newConditionBuilder); retConditionBuilder != nil {
				newConditionBuilder = retConditionBuilder
			}
		}
	}

	return newConditionBuilder
}

func (da *DynamoAttribute[T]) addUpdate(updateBuilder *expression.UpdateBuilder) (newUpdateBuilder *expression.UpdateBuilder, err error) {
	if !da.buildExecuted {
		return nil, errors.New("build is not yet executed on attribute [" + da.name + "], cannot update this attribute")
	}

	if updateBuilder == nil {
		return nil, errors.New("nil update builder passed for attribute " + da.documentPath)
	}

	newUpdateBuilder = updateBuilder
	valueBuilder := expression.Value(da.value)

	switch da.operation {
	case UPDATE_SET:
		newUpdateBuilder = utils.PointerTo((newUpdateBuilder).Set(da.nameBuilder, valueBuilder))
	case UPDATE_REMOVE:
		newUpdateBuilder = utils.PointerTo((newUpdateBuilder).Remove(da.nameBuilder))
	case UPDATE_ADD: // for numbers and set data structure
		newUpdateBuilder = utils.PointerTo((newUpdateBuilder).Add(da.nameBuilder, valueBuilder))
	case UPDATE_DELETE: // for set data structure only
		newUpdateBuilder = utils.PointerTo((newUpdateBuilder).Delete(da.nameBuilder, valueBuilder))
	case NO_OP:
		for _, childAttribute := range da.childAttributes {
			switch childAttributeType := childAttribute.(type) {
			case Updater:
				if retUpdateBuilder, err := childAttributeType.addUpdate(newUpdateBuilder); err != nil {
					return nil, err
				} else if retUpdateBuilder != nil {
					newUpdateBuilder = retUpdateBuilder
				}
			}
		}
	}

	return newUpdateBuilder, nil
}
