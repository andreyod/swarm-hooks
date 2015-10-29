package authZ

const (
	NOT_SUPPORTED EVENT_ENUM = iota
	CONTAINER_CREATE
	CONTAINERS_LIST
	CONTAINER_INSPECT
	CONTAINER_OTHERS
	PASS_AS_IS
	UNAUTHORIZED
	STREAM_OR_HIJACK
)

const (
	APPROVED APPROVAL_ENUM = iota
	NOT_APPROVED
	CONDITION_FILTER
	CONDITION_OVERRIDE
)