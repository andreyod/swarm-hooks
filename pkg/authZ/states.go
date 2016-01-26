package authZ

const (
	//EventEnum - Describes type of events for the Validation logic
	notSupported EventEnum = iota
	containerCreate
	containersList
	containerInspect
	containerOthers
	passAsIs
	unauthorized
	streamOrHijack
)

const (
	//ApprovalEnum - Describes Validations verdict
	approved ApprovalEnum = iota
	notApproved
	conditionFilter
	conditionOverride
)
