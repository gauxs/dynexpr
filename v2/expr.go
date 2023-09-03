package v2

import (
	"dynexpr/utils"
	"errors"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
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
	conditionBuilder expression.ConditionBuilder

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

func (da *DynamoAttribute[T]) constructDocumentPath(documentPathOfParent string) string {
	if strings.HasSuffix(da.GetName(), "]") { // this is a top level element of a list, don't use DDBAtributeNameCancatenator
		return documentPathOfParent + da.GetName()
	}

	if documentPathOfParent == "" { // this is top level attribute
		return da.GetName()
	}

	return documentPathOfParent + DDBAtributeNameCancatenator + da.GetName()
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
	if da.conditionBuilder.IsSet() {
		if conditionBuilder != nil && (*conditionBuilder).IsSet() {
			newConditionBuilder = utils.PointerTo(conditionBuilder.And(da.conditionBuilder))
		} else {
			newConditionBuilder = &da.conditionBuilder
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

// AndWithCondition adds a new condition to `this` attributes existing conditions using `AND`
func (da *DynamoAttribute[T]) AndWithCondition() func(conditionBuilder expression.ConditionBuilder) {
	return func(conditionBuilder expression.ConditionBuilder) {
		if da.conditionBuilder.IsSet() {
			if conditionBuilder.IsSet() {
				da.conditionBuilder = da.conditionBuilder.And(conditionBuilder)
			}
		} else {
			da.conditionBuilder = conditionBuilder
		}
	}
}

// AddValue adds a value which will be used to update `this` attributes
func (da *DynamoAttribute[T]) AddValue(operation DynamoOperation, value any) {
	da.operation = operation
	da.value = value
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
	keyConditionBuilder expression.KeyConditionBuilder
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

// Implicitly uses AND condition between conditions of different attribute
// func (dka *DynamoKeyAttribute[T]) addCondition(keyConditionBuilder *expression.KeyConditionBuilder) (newKeyConditionBuilder *expression.KeyConditionBuilder) {
// 	if !dka.keyConditionBuilder.IsSet() {
// 		return &dka.keyConditionBuilder
// 	}

// 	return utils.PointerTo(keyConditionBuilder.And(dka.keyConditionBuilder))
// }

// AndWithCondition adds a new condition to `this` attributes existing conditions using `AND`
func (dka *DynamoKeyAttribute[T]) AndWithCondition() func(keyConditionBuilder expression.KeyConditionBuilder) {
	return func(keyConditionBuilder expression.KeyConditionBuilder) {
		if dka.keyConditionBuilder.IsSet() {
			if keyConditionBuilder.IsSet() {
				dka.keyConditionBuilder = dka.keyConditionBuilder.And(keyConditionBuilder)
			}
		} else {
			dka.keyConditionBuilder = keyConditionBuilder
		}
	}
}

func (dka *DynamoKeyAttribute[T]) addKeyCondition(keyConditionBuilder *expression.KeyConditionBuilder) (newKeyConditionBuilder *expression.KeyConditionBuilder) {
	if dka.keyConditionBuilder.IsSet() {
		if keyConditionBuilder != nil && (*keyConditionBuilder).IsSet() {
			// parent condition will be L.H.S
			newKeyConditionBuilder = utils.PointerTo(keyConditionBuilder.And(dka.keyConditionBuilder))
		} else {
			newKeyConditionBuilder = &dka.keyConditionBuilder
		}
	}

	return newKeyConditionBuilder
}

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
	conditionBuilder expression.ConditionBuilder

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

func (dla *DynamoListAttribute[T]) constructDocumentPath(documentPathOfParent string) string {
	if strings.HasSuffix(dla.GetName(), "]") { // this is a top level element of a list, we can skip DDBAtributeNameCancatenator
		return documentPathOfParent + dla.GetName()
	}

	if documentPathOfParent == "" { // this is root's child
		return dla.GetName()
	}

	return documentPathOfParent + DDBAtributeNameCancatenator + dla.GetName()
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

func (dla *DynamoListAttribute[T]) addCondition(conditionBuilder *expression.ConditionBuilder) (newConditionBuilder *expression.ConditionBuilder) {
	if dla.conditionBuilder.IsSet() {
		if conditionBuilder != nil && (*conditionBuilder).IsSet() {
			newConditionBuilder = utils.PointerTo(conditionBuilder.And(dla.conditionBuilder))
		} else {
			newConditionBuilder = &dla.conditionBuilder
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

func (dla *DynamoListAttribute[T]) AndWithCondition() func(conditionBuilder expression.ConditionBuilder) {
	return func(conditionBuilder expression.ConditionBuilder) {
		if dla.conditionBuilder.IsSet() {
			if conditionBuilder.IsSet() {
				dla.conditionBuilder = dla.conditionBuilder.And(conditionBuilder)
			}
		} else {
			dla.conditionBuilder = conditionBuilder
		}
	}
}

func (dla *DynamoListAttribute[T]) AddValue(operation DynamoOperation, value any) {
	dla.operation = operation
	dla.value = value
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
	keyConditionBuilder := &expression.KeyConditionBuilder{}
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
	return d.root.addCondition(&expression.ConditionBuilder{})
}

// BuildUpdateBuilder builds a UpdateBuilder by aggregating all the update operation of this
// expression builder tree
func (d DDBItemExpressionBuilder[T]) BuildUpdateBuilder() (*expression.UpdateBuilder, error) {
	return d.root.addUpdate(&expression.UpdateBuilder{})
}
