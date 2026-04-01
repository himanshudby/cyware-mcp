package co

import (
	"github.com/cyware-labs/cyware-mcpserver/common"
	"resty.dev/v3"
)

const Login_endpoint = "/cpapi/rest-auth/login/"

type LoginPayload struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type LoginResponse struct {
	Email   string `json:"email"`
	Message string `json:"message"`
	Token   string `json:"token"`
}

func generateAuthHeaderForConfig(app common.Application) (string, error) {
	login_resp := LoginResponse{}
	login_payload := LoginPayload{
		Email:    app.Auth.Username,
		Password: common.Base64Encode(app.Auth.Password),
	}
	client := common.APIClient{
		BASE_URL: app.BASE_URL,
		Client:   resty.New(),
	}
	_, err := client.MakeRequest("POST", Login_endpoint, nil, &login_resp, login_payload, nil)
	if err != nil {
		return "", err
	}
	return common.FormatCywareToken(login_resp.Token), nil
}
