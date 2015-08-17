package authentication

type ProvisionAPI interface {

	//The Admin should first provision itself before starting to servce
	Init() error

	//Is valid and the label for the token if it is valid.
	ValidateToken(toekn string) (bool, string)


}