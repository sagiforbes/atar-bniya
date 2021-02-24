package infra

//secretStruct base interface of all secrets
type secretStruct interface {
	GetType() string //Get the type of the secret as a string

}

//secretText Save a string as secret
type secretText struct {
	Text string
}

//GetType create the object from a string
func (t secretText) GetType() string {
	return "text"
}

//*******************************************************************************

//secretSSHWithPrivate holds ssh key with private key values
type secretSSHWithPrivate struct {
	User       string `json:"user,omitempty"`
	PrivateKey string `json:"private_key,omitempty"`
	Passphrase string `json:"passphrase,omitempty"`
}

//GetType create the object from a string
func (t secretSSHWithPrivate) GetType() string {
	return "ssh"
}

//*******************************************************************************

//secretSSHWithPassword holds ssh key with user and password key values
type secretUserPassword struct {
	User     string
	Password string
}

//GetType create the object from a string
func (t secretUserPassword) GetType() string {
	return "userpass"
}

//*******************************************************************************
