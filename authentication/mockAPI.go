package authentication

type MockAPI struct{}

func (*MockAPI) ValidateToken(token string) (bool, string) {
	if token == "" {
		return false, "Invalid user token!"
	}
	return true, token
}

func (*MockAPI) Init() error {
	return nil
}
