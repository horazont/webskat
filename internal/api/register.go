package api

type RegisterV1Request struct {
	ClientSecret   string `json:"clientSecret"`
	DisplayName    string `json:"displayName"`
	ServerPassword string `json:"serverPassword"`
}

type RegisterV1Response struct {
	ClientID string `json:"clientID"`
}
