package authZ

const (
	//eventEnum - Describes type of events for the Validation logic
	notSupported eventEnum = iota
	containerCreate
	containersList
	containerInspect
	containerOthers
	passAsIs
	unauthorized
	streamOrHijack
)

const (
	//approvalEnum - Describes Validations verdict
	approved approvalEnum = iota
	notApproved
	conditionFilter
	conditionOverride
)
