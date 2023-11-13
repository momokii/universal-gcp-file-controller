package models

type InputFile struct {
	Filepath   string `json:"filepath"`
	BucketName string `json:"bucketName"`
}

type GetFile struct {
	Name        string `json:"name"`
	PublicURL   string `json:"publicURL"`
	DownloadURL string `json:"downloadURL"`
}

// type InputServiceAccount struct {
// 	// * not included type because default is service_account
// 	ProjectID               string `json:"project_id"`
// 	PrivateKeyID            string `json:"private_key_id"`
// 	PrivateKey              string `json:"private_key"`
// 	ClientEmail             string `json:"client_email"`
// 	ClientID                string `json:"client_id"`
// 	AuthURI                 string `json:"auth_uri"`
// 	TokenURI                string `json:"token_uri"`
// 	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
// 	ClientX509CertURL       string `json:"client_x509_cert_url"`
// 	UniverseDomain          string `json:"universe_domain"`
// }

type InputServiceAccount struct {
	// * not included type because default is service_account
	ProjectID   string `json:"project_id"`
	PrivateKey  string `json:"private_key"`
	ClientEmail string `json:"client_email"`
}

type InputUserData struct {
	UsingToken bool   `json:"using_token"`
	Token      string `json:"token"`
	InputServiceAccount
}

type InputUserAll struct {
	// * get data input
	Folderpath string `json:"folderpath"`
	WithInfo   bool   `json:"with_info"`

	// * delete info data
	Filepath string `json:"filepath"`

	// * general info must filled
	BucketName string `json:"bucketname"`
	InputUserData
}
