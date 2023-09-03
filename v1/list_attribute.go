package v1

import (
	"dynexpr/utils"
	"errors"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

// Represents list/set datastructure in dynamo db
// T specifies the type of data stored in the list
//
// Current Limitations:
// 1 - Lists in dynamo db can store attributes of different types but we are LIMITING this to single type
// 2 - List of list is NOT supported currently
// 3 - List of object is supported but list of object containing another list is NOT
type DynamoListAttribute[T any] struct {
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

	// Helps in direct member selection of child attributes of list item
	// This will also be used in creating new list item
	listItemAccessReference T

	// Determines the operation which needs to performed on
	// 'this' attribute
	operation DynamoOperation

	// Represent the value which needs to be assigned to 'this'
	// attribute for update
	value any

	// Defines the order of index added, this helps in maintaining
	// order when generating projection
	orderOfListItems []int

	// List of item of 'this' list attribute
	listAttributes map[int]interface{}
}

func NewDynamoListAttribute[T any]() *DynamoListAttribute[T] {
	return &DynamoListAttribute[T]{
		listAttributes: map[int]interface{}{},
	}
}

// WithName builds `this` DynamoListAttribute with a dynamo db attribute name
func (dla *DynamoListAttribute[T]) WithName(name string) *DynamoListAttribute[T] {
	dla.name = name
	return dla
}

// WithListItemAccessReference builds `this` DynamoListAttribute with an access reference
// access refernece is used in direct member selection
//
// NOTE: List/Set will hold same type `T`
func (dla *DynamoListAttribute[T]) WithListItemAccessReference(listItemAccessReference T) *DynamoListAttribute[T] {
	dla.listItemAccessReference = listItemAccessReference
	return dla
}

// Project marks `this` attribute for projection
func (dla *DynamoListAttribute[T]) Project() error {
	if !dla.buildExecuted {
		return errors.New("build is not yet executed on attribute [" + dla.name + "], cannot mark this attribute for projection")
	}

	dla.projection = true
	return nil
}

func (dla *DynamoListAttribute[T]) GetName() string {
	return dla.name
}

func (dla *DynamoListAttribute[T]) GetNameBuilder() expression.NameBuilder {
	return dla.nameBuilder
}

// AR returns the type `T` held by this node
// access refernece is used in direct member selection
func (dla *DynamoListAttribute[T]) AR() T {
	return dla.listItemAccessReference
}

// AddListItem add a node in the list
func (dla *DynamoListAttribute[T]) AddListItem(listItemsIndex ...int) error {
	if dla.buildExecuted {
		return errors.New("build is already executed on attribute " + dla.documentPath)
	}

	// currently, we are not allowing nested lists
	for _, index := range listItemsIndex {
		dla.orderOfListItems = append(dla.orderOfListItems, index)
		switch listItemType := interface{}(dla.listItemAccessReference).(type) {
		case TreeBuilder[T]:
			dla.listAttributes[index] = listItemType.BuildTree("[" + strconv.Itoa(int(index)) + "]")
		default: // it's a primitive
			dla.listAttributes[index] = NewDynamoAttribute[T]().WithName("[" + strconv.Itoa(int(index)) + "]")
		}
	}

	return nil
}

func (dla *DynamoListAttribute[T]) Index(listAttributeIndex int) *DynamoAttribute[T] {
	// currently, we are limiting the type to DynamoAttribute
	if listAttribute, ok := dla.listAttributes[listAttributeIndex].(*DynamoAttribute[T]); !ok {
		return nil
	} else {
		return listAttribute
	}
}

// NOTE: conditionBuilder represent any valid condition, zero value of struct `ConditionBuilder`
// might give build error
func (dla *DynamoListAttribute[T]) AndWithCondition() func(conditionBuilder expression.ConditionBuilder) {
	return func(conditionBuilder expression.ConditionBuilder) {
		if dla.conditionBuilder != nil {
			dla.conditionBuilder = utils.PointerTo((*dla.conditionBuilder).And(conditionBuilder))
		} else {
			dla.conditionBuilder = &conditionBuilder
		}
	}
}

func (dla *DynamoListAttribute[T]) AddValue(operation DynamoOperation, value any) {
	dla.operation = operation
	dla.value = value
}

func (dla *DynamoListAttribute[T]) constructDocumentPath(documentPathOfParent string) string {
	if strings.HasSuffix(dla.GetName(), "]") { // this is a top level element of a list, we can skip DDBAtributeNameCancatenator
		return documentPathOfParent + dla.GetName()
	}

	if documentPathOfParent == "" { // this is root's child
		return dla.GetName()
	}

	return documentPathOfParent + DDBAtributeNameCancatenator + dla.GetName()
}

func (dla *DynamoListAttribute[T]) build(parentDocumentPath string) error {
	if dla.buildExecuted {
		return errors.New("build is already executed on attribute " + dla.documentPath)
	}

	// build document path
	dla.documentPath = dla.constructDocumentPath(parentDocumentPath)

	// build name builder
	dla.nameBuilder = expression.Name(dla.documentPath)

	for _, listItem := range dla.listAttributes {
		switch listItemType := listItem.(type) {
		case Builder:
			listItemType.build(dla.documentPath)
		}
	}

	dla.buildExecuted = true
	return nil
}

func (dla *DynamoListAttribute[T]) addName(projectionBuilder *expression.ProjectionBuilder) (newProjectionBuilder *expression.ProjectionBuilder, err error) {
	if !dla.buildExecuted {
		return nil, errors.New("build is not yet executed on attribute [" + dla.name + "], cannot mark this attribute for projection")
	}

	if projectionBuilder == nil {
		return nil, errors.New("nil projection builder passed for attribute " + dla.documentPath)
	}

	newProjectionBuilder = projectionBuilder
	if dla.projection { // skipping projection of child attributes
		return utils.PointerTo(newProjectionBuilder.AddNames(expression.Name(dla.documentPath))), nil
	} else {
		for _, listAttributeIndex := range dla.orderOfListItems {
			listItem := dla.listAttributes[listAttributeIndex]
			switch listItemType := listItem.(type) {
			case Projector:
				newProjectionBuilder, err = listItemType.addName(newProjectionBuilder)
			}

			if err != nil {
				return nil, err
			}
		}
	}

	return newProjectionBuilder, nil
}

func (dla *DynamoListAttribute[T]) addCondition(conditionBuilder *expression.ConditionBuilder) (newConditionBuilder *expression.ConditionBuilder) {
	if dla.conditionBuilder != nil {
		if conditionBuilder != nil {
			newConditionBuilder = utils.PointerTo(conditionBuilder.And(*dla.conditionBuilder))
		} else {
			newConditionBuilder = dla.conditionBuilder
		}
	} else {
		newConditionBuilder = conditionBuilder
	}

	for _, listAttributeIndex := range dla.orderOfListItems {
		listItem := dla.listAttributes[listAttributeIndex]
		switch listItemType := listItem.(type) {
		case Conditioner:
			if retConditionBuilder := listItemType.addCondition(newConditionBuilder); retConditionBuilder != nil {
				newConditionBuilder = retConditionBuilder
			}
		}
	}

	return newConditionBuilder
}

func (dla *DynamoListAttribute[T]) addUpdate(updateBuilder *expression.UpdateBuilder) (newUpdateBuilder *expression.UpdateBuilder, err error) {
	if !dla.buildExecuted {
		return nil, errors.New("build is not yet executed on attribute [" + dla.name + "], cannot update this list attribute")
	}

	if updateBuilder == nil {
		return nil, errors.New("nil update builder passed for attribute " + dla.documentPath)
	}

	newUpdateBuilder = updateBuilder
	valueBuilder := expression.Value(dla.value)

	switch dla.operation {
	case UPDATE_SET:
		switch valueType := interface{}(dla.value).(type) {
		case expression.SetValueBuilder: // this is when we want to perform list_append
			newUpdateBuilder = utils.PointerTo((newUpdateBuilder).Set(dla.nameBuilder, valueType))
		default:
			newUpdateBuilder = utils.PointerTo((newUpdateBuilder).Set(dla.nameBuilder, valueBuilder))
		}
	case UPDATE_REMOVE:
		newUpdateBuilder = utils.PointerTo((newUpdateBuilder).Remove(dla.nameBuilder))
	case UPDATE_ADD: // for numbers and set data structure
		newUpdateBuilder = utils.PointerTo((newUpdateBuilder).Add(dla.nameBuilder, valueBuilder))
	case UPDATE_DELETE: // for set data structure only
		newUpdateBuilder = utils.PointerTo((newUpdateBuilder).Delete(dla.nameBuilder, valueBuilder))
	case NO_OP:
		for _, listAttributeIndex := range dla.orderOfListItems {
			listItem := dla.listAttributes[listAttributeIndex]
			switch listItemType := listItem.(type) {
			case Updater:
				if retUpdateBuilder, err := listItemType.addUpdate(newUpdateBuilder); err != nil {
					return nil, err
				} else if retUpdateBuilder != nil {
					newUpdateBuilder = retUpdateBuilder
				}
			}
		}
	}

	return newUpdateBuilder, nil
}

func NewDDBItemExpressionBuilder[T TreeBuilder[T]](treeBuilder T) DDBItemExpressionBuilder[T] {
	return DDBItemExpressionBuilder[T]{
		root: treeBuilder.BuildTree(""),
	}
}
