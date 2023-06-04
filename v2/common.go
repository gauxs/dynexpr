package v2

type DynamoOperation int

const (
	NO_OP DynamoOperation = iota
	GET
	UPDATE_SET
	UPDATE_REMOVE
	UPDATE_ADD
	UPDATE_DELETE
)
